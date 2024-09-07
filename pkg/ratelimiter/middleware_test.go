package ratelimiter

import (
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

func (m *rateLimiterMock) Allow(r *http.Request) bool {
	return m.Called(r).Bool(0)
}

func (m *httpHandlerMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func TestMiddleware_Allow(t *testing.T) {
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

func TestMiddleware_Deny(t *testing.T) {
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
}

func TestIntegrationMiddleware_Allow(t *testing.T) {
	rl := NewRateLimiter(newStrategyMemoryBased(), newStrategyMemoryBased(), 10)
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

			for {
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
	wg.Wait()
	assert.NotEmpty(t, failuresByThread)
}
