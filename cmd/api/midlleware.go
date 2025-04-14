package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/carlosEA28/Social/internal/repository"
	"github.com/golang-jwt/jwt/v5"
)

func (app *app) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		//valida se o authorization existe
		if authHeader == "" {
			app.unauthorizedErrorReposnse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		//separa o prefixo "Bearer" do token em si
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedErrorReposnse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		//salva apenas a parte do token, e nao o "Bearer"
		token := parts[1]

		//valida o token
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedErrorReposnse(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		//(nao sei direito),acho que tira o id do token e converte para int
		userId, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)

		if err != nil {
			app.unauthorizedErrorReposnse(w, r, err)
			return
		}

		//procura se existe um user com o id retirado do token
		ctx := r.Context()
		user, err := app.store.Users.GetUserById(ctx, userId)
		if err != nil {
			app.unauthorizedErrorReposnse(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
func (app *app) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post := getPostFromCtx(r)

		if post.UserId == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbidenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *app) checkRolePrecedence(ctx context.Context, user *repository.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}
