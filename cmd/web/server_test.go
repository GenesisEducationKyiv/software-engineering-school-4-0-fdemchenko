package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fdemchenko/exchanger/internal/services/rate"
	"github.com/stretchr/testify/assert"
)

func TestRateEndpointIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	rateService := rate.NewRateService(
		rate.WithFetchers(
			rate.NewNBURateFetcher("nbu fetcher"),
			rate.NewFawazRateFetcher("fawaz fetcher"),
			rate.NewPrivatRateFetcher("privat fetcher"),
		),
		rate.WithUpdateInterval(RateCachingDuration),
	)
	app := application{
		rateService: rateService,
	}

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	rs, err := ts.Client().Get(ts.URL + "/rate")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, rs.StatusCode)

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	rate, err := strconv.ParseFloat(string(body), 32)
	assert.NoError(t, err)
	assert.Greater(t, rate, float64(0))
}
