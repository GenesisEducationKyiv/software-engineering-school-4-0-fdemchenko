package rate

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/fdemchenko/exchanger/internal/cache"
	"github.com/rs/zerolog/log"
)

const (
	DefaultCachingDuration = 15 * time.Minute
	RateFetchTimeout       = 10 * time.Second
)

var ErrInvalidCurrencyCode = errors.New("invalid currency code")

type RateFetcher interface {
	Fetch(ctx context.Context, code string, client *http.Client) (float32, error)
	Name() string
}

type cachingRateService struct {
	fetchers       []RateFetcher
	client         *http.Client
	updateInterval time.Duration
	cache          *cache.Cache[string, float32]
}

type Option func(*cachingRateService)

func WithClient(client *http.Client) Option {
	return func(crs *cachingRateService) {
		crs.client = client
	}
}

func WithUpdateInterval(updateInterval time.Duration) Option {
	return func(crs *cachingRateService) {
		crs.updateInterval = updateInterval
	}
}

func WithFetchers(fetchers ...RateFetcher) Option {
	return func(crs *cachingRateService) {
		crs.fetchers = fetchers
	}
}

func NewRateService(options ...Option) *cachingRateService {
	// caching rate service with default values.
	service := &cachingRateService{
		client:         http.DefaultClient,
		updateInterval: DefaultCachingDuration,
		fetchers:       []RateFetcher{NewNBURateFetcher("nbu fetcher")},
		cache:          cache.New[string, float32](),
	}

	for _, option := range options {
		option(service)
	}
	return service
}

func (crs *cachingRateService) GetRate(ctx context.Context, currencyCode string) (float32, error) {
	if rate, exists := crs.cache.Get(strings.ToLower(currencyCode)); exists {
		return rate, nil
	}

	var err error
	var rate float32
	for _, fetcher := range crs.fetchers {
		rate, err = fetcher.Fetch(ctx, currencyCode, crs.client)
		log.Debug().Str("name", fetcher.Name()).Float32("rate", rate).Err(err).Send()
		if err == nil {
			crs.cache.Set(strings.ToLower(currencyCode), rate, crs.updateInterval)
			return rate, nil
		}
		log.Warn().Str("provider", fetcher.Name()).Err(err).Msg("Fallback to another provider")
	}

	return 0, err
}
