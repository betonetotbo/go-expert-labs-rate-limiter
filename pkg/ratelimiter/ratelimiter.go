package ratelimiter

import (
	"net/http"
	"strings"
)

type (
	RateLimiter interface {
		Allow(r *http.Request) bool
	}

	rateLimiter struct {
		ipStrategy    Strategy
		tokenStrategy Strategy
		limit         int64
	}
)

func NewRateLimiter(ipStrategy Strategy, tokenStrategy Strategy, limit int64) RateLimiter {
	return &rateLimiter{
		ipStrategy:    ipStrategy,
		tokenStrategy: tokenStrategy,
		limit:         limit,
	}
}

func (rl *rateLimiter) Allow(r *http.Request) bool {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	token := r.Header.Get("API_KEY")

	if token != "" {
		count, err := rl.tokenStrategy.Inc(r.Context(), token)
		if err != nil || count > rl.limit {
			return false
		}
	}

	count, err := rl.ipStrategy.Inc(r.Context(), ip)
	return err == nil && count <= rl.limit
}
