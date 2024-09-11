package ratelimiter

import (
	"betonetotbo/go-expert-labs-rate-limiter/internal/config"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type (
	rateLimiterMock struct {
		mock.Mock
	}
	httpHandlerMock struct {
		mock.Mock
	}

	strategyMemoryBasedItem struct {
		counter   int64
		expiresAt time.Time
	}
	strategyMemoryBased struct {
		counters map[string]*strategyMemoryBasedItem
		m        sync.Mutex
	}
	strategyFailureBased struct {
	}
)

func newStrategyMemoryBased() Strategy {
	return &strategyMemoryBased{
		counters: make(map[string]*strategyMemoryBasedItem),
	}
}

func (s *strategyMemoryBased) Inc(_ context.Context, value string) (int64, error) {
	s.m.Lock()
	defer s.m.Unlock()

	i, ok := s.counters[value]
	if !ok || time.Now().After(i.expiresAt) {
		i = &strategyMemoryBasedItem{
			counter:   1,
			expiresAt: time.Now().Add(time.Second),
		}
		s.counters[value] = i
	} else {
		i.counter++
	}
	return i.counter, nil
}

func (s *strategyFailureBased) Inc(_ context.Context, _ string) (int64, error) {
	panic("call not expected")
}

func (m *rateLimiterMock) Allow(r *http.Request) bool {
	return m.Called(r).Bool(0)
}

func (m *httpHandlerMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func TestMiddleware_AllowIp(t *testing.T) {
	// Assemble
	sm := &rateLimiterMock{}
	sm.On("Allow", mock.Anything).Return(true)

	hm := &httpHandlerMock{}
	hm.On("ServeHTTP", mock.Anything, mock.Anything)

	m := NewMiddleware(sm)(hm)

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Act
	m.ServeHTTP(w, r)

	// Verify
	assert.Equal(t, w.Code, http.StatusOK)
}

func TestMiddleware_DenyIp(t *testing.T) {
	// Assemble
	sm := &rateLimiterMock{}
	sm.On("Allow", mock.Anything).Return(false)

	hm := &httpHandlerMock{}
	hm.On("ServeHTTP", mock.Anything, mock.Anything)

	m := NewMiddleware(sm)(hm)

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Act
	m.ServeHTTP(w, r)

	// Verify
	assert.Equal(t, w.Code, http.StatusTooManyRequests)
	assert.Equal(t, rateLimiterMessage, string(w.Body.Bytes()))
}

func TestMiddlewareIntegration_AllowAndDenyByIp(t *testing.T) {
	// Assemble
	rl := NewRateLimiter(newStrategyMemoryBased(), &strategyFailureBased{}, &config.Config{
		Rps:      10,
		Interval: time.Minute,
		TokenRps: config.TokenRps{
			Values: map[string]int{
				"inexistent": 10,
			},
		},
	})
	sv := httptest.NewServer(NewMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	})))
	defer sv.Close()

	// Act
	failuresByThread := map[int]int{}
	var m sync.Mutex

	var wg sync.WaitGroup
	for idx := range 20 {
		wg.Add(1)
		go func() {
			r, _ := http.NewRequest("GET", sv.URL, nil)
			r.Header.Set("API_KEY", "token")
			c := &http.Client{}

			for timeout := time.After(time.Second * 5); ; {
				select {
				case <-timeout:
					return
				default:
				}
				time.Sleep(time.Millisecond * 10)

				resp, err := c.Do(r)

				if err != nil {
					fmt.Printf("%v - %d - Failed to make request: %v", time.Now(), idx, err)
				} else {
					fmt.Printf("%v - %d - Response status: %d\n", time.Now().Format("15:04:05.000"), idx, resp.StatusCode)
					if resp.StatusCode == 201 {
						break
					}
				}
				m.Lock()
				failuresByThread[idx]++
				m.Unlock()
			}

			wg.Done()
		}()
	}

	// Verify
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		assert.NotEmpty(t, failuresByThread)
	case <-time.After(time.Second * 10):
		assert.Fail(t, "test timeout")
	}
}

func TestMiddlewareIntegration_AllowAndDenyByToken(t *testing.T) {
	// Assemble
	rl := NewRateLimiter(&strategyFailureBased{}, newStrategyMemoryBased(), &config.Config{
		Rps:      10,
		Interval: time.Minute,
		TokenRps: config.TokenRps{
			Values: map[string]int{
				"token": 10,
			},
		},
	})
	sv := httptest.NewServer(NewMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	})))
	defer sv.Close()

	// Act
	failuresByThread := map[int]int{}
	var m sync.Mutex

	var wg sync.WaitGroup
	for idx := range 20 {
		wg.Add(1)
		go func() {
			r, _ := http.NewRequest("GET", sv.URL, nil)
			r.Header.Set("API_KEY", "token")
			c := &http.Client{}

			for timeout := time.After(time.Second * 5); ; {
				select {
				case <-timeout:
					return
				default:
				}
				time.Sleep(time.Millisecond * 10)
				resp, err := c.Do(r)

				if err != nil {
					fmt.Printf("%v - %d - Failed to make request: %v", time.Now(), idx, err)
				} else {
					fmt.Printf("%v - %d - Response status: %d\n", time.Now().Format("15:04:05.000"), idx, resp.StatusCode)
					if resp.StatusCode == 201 {
						break
					}
				}
				m.Lock()
				failuresByThread[idx]++
				m.Unlock()
			}

			wg.Done()
		}()
	}

	// Verify
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		assert.NotEmpty(t, failuresByThread)
	case <-time.After(time.Second * 10):
		assert.Fail(t, "test timeout")
	}
}
