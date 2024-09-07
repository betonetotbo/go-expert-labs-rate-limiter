package main

import (
	"betonetotbo/go-expert-labs-rate-limiter/pkg/ratelimiter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

func main() {
	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	var r = chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(ratelimiter.NewMiddleware(
		ratelimiter.NewRateLimiter(ratelimiter.NewRedisStrategy(rc, "ip", time.Second*1),
			ratelimiter.NewRedisStrategy(rc, "token", time.Second*1),
			10,
		),
	))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	http.ListenAndServe(":3000", r)
}
