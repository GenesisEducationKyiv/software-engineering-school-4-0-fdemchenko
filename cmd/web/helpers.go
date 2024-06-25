package main

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	log.Error().Err(err).Send()
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
