package main

import (
	"net/http"
	"strconv"

	"github.com/carlosEA28/Social/internal/repository"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

func (app *app) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	followedUserId, err := strconv.ParseInt(chi.URLParam(r, "UserId"), 10, 64)

	if err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, followerUser); err != nil {
		app.internalServerError(w, r, err)
	}

	ctx := r.Context()

	if err := app.store.Followers.Follow(ctx, followerUser.ID, followedUserId); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	unfollowedUserId, err := strconv.ParseInt(chi.URLParam(r, "UserId"), 10, 64)
	if err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.Followers.Unfollow(ctx, followerUser.ID, unfollowedUserId); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)

	if err != nil {
		switch err {
		case repository.ErrorNotFound:
			app.badRequetResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, ""); err != nil {
		app.internalServerError(w, r, err)
	}
}

func getUserFromContext(r *http.Request) *repository.User {
	user, _ := r.Context().Value(userCtx).(*repository.User)
	return user
}
