package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

const ServerShutdownTimeout = 5 * time.Second

func (app *application) serveHTTP() error {
	server := http.Server{
		Handler:           app.routes(),
		Addr:              app.cfg.addr,
		WriteTimeout:      ServerTimeout,
		ReadHeaderTimeout: ServerTimeout,
	}

	shutdownErrors := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		quitSignal := <-quit

		log.Debug().Str("signal", quitSignal.String()).Msg("app terminated")

		ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
		defer cancel()

		shutdownErrors <- server.Shutdown(ctx)
	}()

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return <-shutdownErrors
}
