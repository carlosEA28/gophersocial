package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/carlosEA28/Social/internal/repository"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

func (app *app) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)

	post := &repository.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserId:  user.ID,
	}

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}
func (app *app) getPostHandler(w http.ResponseWriter, r *http.Request) {

	post := getPostFromCtx(r) // ver se nao quebrou

	comments, err := app.store.Comment.GetByPostId(r.Context(), post.ID)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

func (app *app) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	// Obtém o ID do post nos parâmetros da URL
	idParam := chi.URLParam(r, "postId")

	// Converte para inteiro
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequetResponse(w, r, errors.New("invalid post ID")) // Retorna 400 em vez de 500
		return
	}

	ctx := r.Context()

	// Tenta excluir o post
	err = app.store.Posts.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrorNotFound):
			app.notFounResponse(w, r, err) // Correção do nome da função
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 - Sucesso sem conteúdo
}

func (app *app) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequetResponse(w, r, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if err := app.store.Posts.Update(r.Context(), post); err != nil {
		switch {
		case errors.Is(err, repository.ErrorNotFound):
			app.notFounResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}

		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}

}
func (app *app) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postId")         // paga o id nos parametros
		id, err := strconv.ParseInt(idParam, 10, 64) //converte para int

		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		post, err := app.store.Posts.GetById(ctx, id)

		if err != nil {
			switch {
			case errors.Is(err, repository.ErrorNotFound):
				app.notFounResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)

			}

			return
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

func getPostFromCtx(r *http.Request) *repository.Post {
	post, _ := r.Context().Value(postCtx).(*repository.Post)
	return post
}
