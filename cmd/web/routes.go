package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fdemchenko/exchanger/internal/repositories"
	"github.com/fdemchenko/exchanger/internal/validator"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /rate", app.getRate)
	mux.HandleFunc("POST /subscribe", app.subscribe)

	return app.recoveryMiddleware(app.loggingMiddleware(app.secureHeadersMiddleware(mux)))
}

func (app *application) getRate(w http.ResponseWriter, r *http.Request) {
	rate, err := app.rateService.GetRate(r.Context(), "usd")
	if err != nil {
		app.serverError(w, err)
		return
	}
	fmt.Fprintf(w, "%f", rate)
}

func (app *application) subscribe(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	newEmail := r.PostForm.Get("email")
	v := validator.New()
	v.Check(validator.IsValidEmail(newEmail), "email", "invalid email")
	if !v.IsValid() {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.emailService.Create(newEmail)
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicateEmail) {
			app.clientError(w, http.StatusConflict)
			return
		}
		app.serverError(w, err)
	}
}
