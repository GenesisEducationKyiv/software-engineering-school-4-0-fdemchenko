package rate

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

const (
	NBUExchangeRateURL = "https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange?json"
)

type NBUResponse struct {
	Rate float32 `json:"rate"`
	Code string  `json:"cc"`
}

type NBURateFetcher struct {
	name string
}

func (nrf NBURateFetcher) Name() string {
	return nrf.name
}

func NewNBURateFetcher(name string) NBURateFetcher {
	return NBURateFetcher{name: name}
}

func (nrf NBURateFetcher) Fetch(ctx context.Context, code string, client *http.Client) (float32, error) {
	ctx, cancel := context.WithTimeout(ctx, RateFetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, NBUExchangeRateURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var nbuResponse []NBUResponse
	err = json.NewDecoder(resp.Body).Decode(&nbuResponse)
	if err != nil {
		return 0, err
	}
	for _, rate := range nbuResponse {
		if strings.EqualFold(code, rate.Code) {
			return rate.Rate, nil
		}
	}
	return 0, ErrInvalidCurrencyCode
}
