package rate

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const (
	PrivatExchangeRateURL = "https://api.privatbank.ua/p24api/pubinfo?json&exchange&coursid=5"
)

type PrivatResponse struct {
	Base     string `json:"base"`
	Currency string `json:"ccy"`
	Buy      string `json:"buy"`
	Sale     string `json:"sale"`
}

type PrivatRateFetcher struct {
	name string
}

func (prf PrivatRateFetcher) Name() string {
	return prf.name
}

func NewPrivatRateFetcher(name string) PrivatRateFetcher {
	return PrivatRateFetcher{name: name}
}

func (prf PrivatRateFetcher) Fetch(ctx context.Context, code string, client *http.Client) (float32, error) {
	ctx, cancel := context.WithTimeout(ctx, RateFetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, PrivatExchangeRateURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var privatResponse []PrivatResponse
	err = json.NewDecoder(resp.Body).Decode(&privatResponse)
	if err != nil {
		return 0, err
	}
	for _, rate := range privatResponse {
		if strings.EqualFold(code, rate.Currency) {
			rateFloat64, err := strconv.ParseFloat(rate.Buy, 32)
			if err != nil {
				return 0, err
			}
			return float32(rateFloat64), nil
		}
	}
	return 0, ErrInvalidCurrencyCode
}
