package ratelimiter

import (
	"betonetotbo/go-expert-labs-rate-limiter/internal/config"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http/httptest"
	"testing"
	"time"
)

type (
	strategyMock struct {
		mock.Mock
	}
)

func (sm *strategyMock) Inc(ctx context.Context, value string) (int64, error) {
	results := sm.Called(ctx, value)
	return results.Get(0).(int64), results.Error(1)
}

func TestRateLimiter_Allow(t *testing.T) {
	// Assemble
	tokenStrategyMock := &strategyMock{}
	tokenStrategyMock.On("Inc", mock.Anything, "token").Return(int64(1), nil)

	ipStrategyMock := &strategyMock{}
	ipStrategyMock.On("Inc", mock.Anything, "127.0.0.1").Return(int64(1), nil)

	rm := NewRateLimiter(ipStrategyMock, tokenStrategyMock, &config.Config{Rps: 10, Interval: time.Minute})

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1"
	r.Header.Set("API_KEY", "token")

	// Act
	allow := rm.Allow(r)

	// Verify
	assert.True(t, allow)
}

func TestRateLimiter_DenyByIp(t *testing.T) {
	// Assemble
	tokenStrategyMock := &strategyMock{}
	tokenStrategyMock.On("Inc", mock.Anything, mock.Anything).Panic("call not expected")

	ipStrategyMock := &strategyMock{}
	ipStrategyMock.On("Inc", mock.Anything, "127.0.0.1").Return(int64(11), nil)

	rm := NewRateLimiter(ipStrategyMock, tokenStrategyMock, &config.Config{
		Rps:      10,
		Interval: time.Minute,
		TokenRps: config.TokenRps{
			Values: map[string]int{
				"token": 10,
			},
		},
	})

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1"
	r.Header.Set("API_KEY", "token2")

	// Act
	allow := rm.Allow(r)

	// Verify
	assert.False(t, allow)
}

func TestRateLimiter_DenyByToken(t *testing.T) {
	// Assemble
	tokenStrategyMock := &strategyMock{}
	tokenStrategyMock.On("Inc", mock.Anything, "token").Return(int64(11), nil)

	ipStrategyMock := &strategyMock{}
	ipStrategyMock.On("Inc", mock.Anything, mock.Anything).Panic("call not expected")

	rm := NewRateLimiter(ipStrategyMock, tokenStrategyMock, &config.Config{
		Rps:      10,
		Interval: time.Minute,
		TokenRps: config.TokenRps{
			Values: map[string]int{
				"token": 10,
			},
		},
	})

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1"
	r.Header.Set("API_KEY", "token")

	// Act
	allow := rm.Allow(r)

	// Verify
	assert.False(t, allow)
}
