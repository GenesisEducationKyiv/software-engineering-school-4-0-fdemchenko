package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/fdemchenko/exchanger/cmd/web/internal/repositories"
	"github.com/fdemchenko/exchanger/internal/communication"
	"github.com/fdemchenko/exchanger/internal/communication/customers"
	"github.com/fdemchenko/exchanger/internal/validator"
	"github.com/justinas/alice"
	"github.com/rs/zerolog/log"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /rate", app.getRate)
	mux.HandleFunc("POST /subscribe", app.subscribe)
	mux.HandleFunc("POST /unsubscribe", app.unsubscribe)
	mux.HandleFunc("GET /metrics", app.metrics)

	middlewares := alice.New(
		app.recoveryMiddleware,
		app.loggingMiddleware,
		app.secureHeadersMiddleware,
		app.RequestCounterMiddleware,
	)
	return middlewares.Then(mux)
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
	s := fmt.Sprintf(`total_subscribers{success="%v"}`, err == nil)
	metrics.GetOrCreateCounter(s).Inc()

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
	s := fmt.Sprintf(`total_unsubscribers{success="%v"}`, err == nil)
	metrics.GetOrCreateCounter(s).Inc()
}

func (app *application) metrics(w http.ResponseWriter, _ *http.Request) {
	metrics.WritePrometheus(w, true)
}
