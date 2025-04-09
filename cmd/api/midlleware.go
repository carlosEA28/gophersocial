package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

		claims := jwtToken.Claims.(jwt.MapClaims)

		//(nao sei direito),acho que tira o id do token e converte para int
		userId, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)

		if err != nil {
			app.unauthorizedErrorReposnse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		//procura se existe um user com o id retirado do token
		ctx := r.Context()
		user, err := app.store.Users.GetUserById(ctx, userId)
		if err != nil {
			app.unauthorizedErrorReposnse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
