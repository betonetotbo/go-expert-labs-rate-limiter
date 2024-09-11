package ratelimiter

import (
	"betonetotbo/go-expert-labs-rate-limiter/internal/config"
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
		cfg           *config.Config
	}
)

func NewRateLimiter(ipStrategy Strategy, tokenStrategy Strategy, cfg *config.Config) RateLimiter {
	return &rateLimiter{
		ipStrategy:    ipStrategy,
		tokenStrategy: tokenStrategy,
		cfg:           cfg,
	}
}

func (rl *rateLimiter) Allow(r *http.Request) bool {
	ip := strings.Split(r.RemoteAddr, ":")[0]

	if rl.cfg.TokenRps.Values != nil {
		token := r.Header.Get("API_KEY")
		if token != "" {
			rps, found := rl.cfg.TokenRps.Values[token]
			if found {
				count, err := rl.tokenStrategy.Inc(r.Context(), token)
				if err != nil || count > int64(rps) {
					return false
				}
				return true
			}
		}
	}

	count, err := rl.ipStrategy.Inc(r.Context(), ip)
	return err == nil && count <= int64(rl.cfg.Rps)
}
