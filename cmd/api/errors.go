package main

import (
	"log"
	"net/http"
)

func (app *app) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s path: %s, error: %s", r.Method, r.URL.Path, err)

	app.logger.Errorw("Internal error", "method:%s", "path:%s", "error", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusInternalServerError, "the server encounterd a problem")

}
func (app *app) forbidenResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.Warnw("Forbidden: %s path: %s, error: %s", r.Method, r.URL.Path)

	app.logger.Errorw("Internal error", "method:%s", "path:%s", "error", r.Method, r.URL.Path)
	writeJSONError(w, http.StatusInternalServerError, "forbidden")

}

func (app *app) badRequetResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("Internal error", "method:%s", "path:%s", "error", r.Method, r.URL.Path, err)

	writeJSONError(w, http.StatusBadRequest, err.Error())

}

func (app *app) notFounResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("Internal error", "method:%s", "path:%s", "error", r.Method, r.URL.Path, err)

	writeJSONError(w, http.StatusNotFound, "not found")

}

func (app *app) unauthorizedErrorReposnse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("Internal error", "method:%s", "path:%s", "error", r.Method, r.URL.Path, err)

	writeJSONError(w, http.StatusNotFound, "Unauthorized")

}
