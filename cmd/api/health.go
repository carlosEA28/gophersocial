package main

import (
	"net/http"
)

func (app *app) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Status":  "ok",
		"Env":     app.config.env,
		"version": version,
	}

	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
	}
}
