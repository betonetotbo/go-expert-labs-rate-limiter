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
	"time"
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
		ratelimiter.NewRateLimiter(ratelimiter.NewRedisStrategy(rc, "ip", time.Second*1),
			ratelimiter.NewRedisStrategy(rc, "token", time.Second*1),
			10,
		),
	))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	log.Printf("Listening on port %d\n", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r)
}
