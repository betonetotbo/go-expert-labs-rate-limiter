package ratelimiter

import (
	"betonetotbo/go-expert-labs-rate-limiter/internal/config"
	"log"
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
					if err != nil {
						log.Printf("Failed to apply rating limit rules: %v\n", err)
					} else {
						log.Printf("Rate limit exceeded for token '%s' (count %d)\n", token, count)
					}
					return false
				}
				return true
			}
		}
	}

	count, err := rl.ipStrategy.Inc(r.Context(), ip)
	if err != nil {
		log.Printf("Failed to apply rating limit rules by IP: %v\n", err)
		return false
	}
	allowed := count <= int64(rl.cfg.Rps)
	if !allowed {
		log.Printf("Rate limit exceeded for IP '%s' (count %d)\n", ip, count)
	}
	return allowed
}
