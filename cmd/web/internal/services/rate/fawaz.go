package rate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	FawazAhmedExchangeRateURL = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1/currencies"
)

type FawazAhmedResponse struct {
	Rates struct {
		UAH float32 `json:"uah"`
	} `json:"usd"`
}

type FawazRateFetcher struct {
	name string
}

func (frf FawazRateFetcher) Name() string {
	return frf.name
}

func NewFawazRateFetcher(name string) FawazRateFetcher {
	return FawazRateFetcher{name: name}
}

func (frf FawazRateFetcher) Fetch(ctx context.Context, code string, client *http.Client) (float32, error) {
	ctx, cancel := context.WithTimeout(ctx, RateFetchTimeout)
	defer cancel()

	reqURL := fmt.Sprintf("%s/%s.json", FawazAhmedExchangeRateURL, strings.ToLower(code))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return 0, ErrInvalidCurrencyCode
	}

	defer resp.Body.Close()
	var fawazAhmedResponse FawazAhmedResponse
	err = json.NewDecoder(resp.Body).Decode(&fawazAhmedResponse)
	if err != nil {
		return 0, err
	}
	return fawazAhmedResponse.Rates.UAH, nil
}
