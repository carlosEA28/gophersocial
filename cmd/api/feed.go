package main

import (
	"net/http"

	"github.com/carlosEA28/Social/internal/repository"
)

func (app *app) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	fq := repository.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	ctx := r.Context()
	feed, err := app.store.Posts.GetUserFeed(ctx, int64(1), fq)

	if err != nil {
		app.internalServerError(w, r, err)
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
	}
}
