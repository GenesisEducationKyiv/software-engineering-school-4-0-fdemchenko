package main

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, data envelope, statusCode int) error {
	jsBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	jsBytes = append(jsBytes, '\n') // for better terminal output

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(jsBytes)
	return err
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	log.Error().Err(err).Send()
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
