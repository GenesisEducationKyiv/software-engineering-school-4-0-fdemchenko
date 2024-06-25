package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecureHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			t.Fatal(err)
		}
	})

	app := application{}

	app.secureHeadersMiddleware(handler).ServeHTTP(recorder, request)
	response := recorder.Result()
	assert.Equal(t, response.StatusCode, http.StatusOK)

	for headerKey, headerValue := range secureHeaders {
		assert.Equal(t, headerValue, response.Header.Get(headerKey))
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, string(body), "OK")
}
