package rate

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	TestingRateServiceFetchInterval   = time.Second * 1
	TestingRateServiceWaitingDuration = time.Second * 2
	ExpectedExchangeRate              = float32(8.0)
)

type MockNBUFetcher struct {
	mock.Mock
}

func (mnf *MockNBUFetcher) Fetch(ctx context.Context, code string, client *http.Client) (float32, error) {
	args := mnf.Called()
	return args.Get(0).(float32), args.Error(1)
}

func (mnf *MockNBUFetcher) Name() string {
	return "nbu fetcher"
}

type MockFawazFetcher struct {
	mock.Mock
}

func (mff *MockFawazFetcher) Fetch(ctx context.Context, code string, client *http.Client) (float32, error) {
	args := mff.Called()
	return args.Get(0).(float32), args.Error(1)
}

func (mff *MockFawazFetcher) Name() string {
	return "fawaz fetcher"
}

type MockPrivatFetcher struct {
	mock.Mock
}

func (mpf *MockPrivatFetcher) Fetch(ctx context.Context, code string, client *http.Client) (float32, error) {
	args := mpf.Called()
	return args.Get(0).(float32), args.Error(1)
}

func (mpf *MockPrivatFetcher) Name() string {
	return "privat fetcher"
}

func TestRateService_RateIsCorrect(t *testing.T) {
	ctx := context.Background()
	mockNBUFetcher := new(MockNBUFetcher)
	mockNBUFetcher.On("Fetch").Return(ExpectedExchangeRate, nil)

	rateService := NewRateService(WithFetchers(mockNBUFetcher))
	rate, err := rateService.GetRate(ctx, "usd")

	assert.Equal(t, float32(8.0), rate)
	assert.NoError(t, err)
}

func TestRateService_RateIsCached(t *testing.T) {
	ctx := context.Background()
	mockNBUFetcher := new(MockNBUFetcher)
	mockNBUFetcher.On("Fetch").Return(float32(8.0), nil)

	rateService := NewRateService(WithFetchers(mockNBUFetcher))
	_, _ = rateService.GetRate(ctx, "usd")
	_, _ = rateService.GetRate(ctx, "usd")

	// Make sure service cached result from previoues calls.
	mockNBUFetcher.AssertNumberOfCalls(t, "Fetch", 1)
}

func TestRateService_RateIsFetchedAfterInterval(t *testing.T) {
	ctx := context.Background()
	mockNBUFetcher := new(MockNBUFetcher)
	mockNBUFetcher.On("Fetch").Return(ExpectedExchangeRate, nil)

	rateService := NewRateService(
		WithFetchers(mockNBUFetcher),
		WithUpdateInterval(TestingRateServiceFetchInterval))

	// Make sure service re-fetch after update interval.
	_, _ = rateService.GetRate(ctx, "usd")
	time.Sleep(TestingRateServiceWaitingDuration)
	_, _ = rateService.GetRate(ctx, "usd")

	mockNBUFetcher.AssertNumberOfCalls(t, "Fetch", 2)
}

func TestRateService_FallbackToAnotherFetcher(t *testing.T) {
	ctx := context.Background()
	mockNBUFetcher := new(MockNBUFetcher)
	mockNBUFetcher.On("Fetch").Return(float32(0), ErrInvalidCurrencyCode)

	mockFawazFetcher := new(MockFawazFetcher)
	mockFawazFetcher.On("Fetch").Return(float32(0), ErrInvalidCurrencyCode)

	mockPrivatFetcher := new(MockPrivatFetcher)
	mockPrivatFetcher.On("Fetch").Return(ExpectedExchangeRate, nil)

	rateService := NewRateService(
		WithFetchers(mockNBUFetcher, mockFawazFetcher, mockPrivatFetcher),
		WithUpdateInterval(TestingRateServiceFetchInterval))

	rate, err := rateService.GetRate(ctx, "usd")
	assert.NoError(t, err)
	assert.Equal(t, ExpectedExchangeRate, rate)

	mockNBUFetcher.AssertCalled(t, "Fetch")
	mockFawazFetcher.AssertCalled(t, "Fetch")
	mockPrivatFetcher.AssertCalled(t, "Fetch")
}
