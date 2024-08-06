package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/fdemchenko/exchanger/internal/communication"
	"github.com/fdemchenko/exchanger/internal/communication/customers"
	"github.com/fdemchenko/exchanger/internal/repositories"
	"github.com/fdemchenko/exchanger/internal/validator"
	"github.com/rs/zerolog/log"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /rate", app.getRate)
	mux.HandleFunc("POST /subscribe", app.subscribe)
	mux.HandleFunc("POST /unsubscribe", app.unsubscribe)

	return app.recoveryMiddleware(app.loggingMiddleware(app.secureHeadersMiddleware(mux)))
}

func (app *application) getRate(w http.ResponseWriter, r *http.Request) {
	rate, err := app.rateService.GetRate(r.Context(), "usd")
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.writeJSON(w, envelope{"rate": rate}, http.StatusOK)
	if err != nil {
		app.serverError(w, err)
	}
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

	id, err := app.emailService.Create(newEmail)
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicateEmail) {
			app.clientError(w, http.StatusConflict)
			return
		}
		app.serverError(w, err)
		return
	}

	msg := communication.Message[customers.CreateCustomerRequestPayload]{
		MessageHeader: communication.MessageHeader{Type: customers.CreateCustomerRequest, Timestamp: time.Now()},
		Payload:       customers.CreateCustomerRequestPayload{Email: newEmail, SubscriptionID: id},
	}
	err = app.customerProducer.SendMessage(msg, customers.CreateCustomerRequestQueue)
	if err != nil {
		log.Error().Err(err).Send()
	}
}

func (app *application) unsubscribe(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	email := r.PostForm.Get("email")
	err = app.emailService.DeleteByEmail(email)
	if err != nil {
		if errors.Is(err, repositories.ErrEmailDoesNotExist) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, err)
	}
}
