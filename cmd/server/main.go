package main

import (
	"betonetotbo/go-expert-labs-rate-limiter/internal/config"
	"betonetotbo/go-expert-labs-rate-limiter/pkg/ratelimiter"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
)

func main() {
	log.Println("Loading config...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Config: %+v\n", *cfg)

	rc := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
	})

	var r = chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(ratelimiter.NewMiddleware(
		ratelimiter.NewRateLimiter(ratelimiter.NewRedisStrategy(rc, "rate-limiter-by-ip-", cfg.Interval),
			ratelimiter.NewRedisStrategy(rc, "rate-limiter-by-token-", cfg.Interval),
			cfg,
		),
	))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("welcome"))
	})

	log.Printf("Listening on port %d\n", cfg.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r)
	if err != nil {
		log.Fatal(err)
	}
}
