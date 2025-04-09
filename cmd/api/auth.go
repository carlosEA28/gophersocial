package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/carlosEA28/Social/internal/mail"
	"github.com/carlosEA28/Social/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*repository.User
	Token string `json:"token"`
}

func (app *app) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	user := &repository.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	//hash the user password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()

	//hash the token for storage but keep the plain token for email
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	//store the user
	if err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp); err != nil {
		switch err {
		case repository.ErrorDuplicateEmail:
			app.badRequetResponse(w, r, err)
		case repository.ErrorDuplicateUsername:
			app.badRequetResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}

		return
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}

	activationUrl := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)

	isProdEnv := app.config.env == "production"

	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationUrl,
	}

	//send mail
	_, err := app.mail.Send(mail.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)

		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("error deleting user", "error", err)
		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

func (app *app) createTokenHandler(w http.ResponseWriter, r *http.Request) {

	//parse the payload credentials
	type CreateUserTokenPayload struct {
		Email    string `json:"email" validate:"required,email,max=255"`
		Password string `json:"password" validate:"required,max=72"`
	}
	var payload CreateUserTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	//fetch the user(check if the user exists) from the payload
	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case repository.ErrorNotFound:
			app.unauthorizedErrorReposnse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	//generate the token -> add claims
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.expDate).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.issuer,
		"aud": app.config.auth.token.issuer,
	}
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	//send to the client
	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}
