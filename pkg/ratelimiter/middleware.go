package ratelimiter

import "net/http"

const (
	rateLimiterMessage = "you have reached the maximum number of requests or actions allowed within a certain time frame"
)

func NewMiddleware(limiter RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(r) {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(rateLimiterMessage))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
