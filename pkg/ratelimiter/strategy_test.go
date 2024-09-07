package ratelimiter

import (
	"context"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestStrategy(t *testing.T) {
	// Assemble
	key := "key"
	completeKey := "keyvalue"
	rc, mock := redismock.NewClientMock()

	mock.MatchExpectationsInOrder(true)

	mock.ExpectIncr(completeKey).SetVal(int64(1))
	mock.ExpectExpire(completeKey, time.Second).SetVal(true)

	rm := NewRedisStrategy(rc, key, time.Second)

	// Act
	val, err := rm.Inc(context.Background(), "value")

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)
}
