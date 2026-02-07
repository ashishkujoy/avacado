package kv

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestDecr_DecrNonExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	counter1, err := testClient.Decr(ctx, "decr_counter1").Result()
	assert.NoError(t, err)
	assert.Equal(t, counter1, int64(-1))

	counter1, err = testClient.Decr(ctx, "decr_counter1").Result()
	assert.NoError(t, err)
	assert.Equal(t, counter1, int64(-2))
}

func TestDecr_DecrExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.SetArgs(ctx, "decr_counter2", "10", redis.SetArgs{TTL: 1 * time.Second})
	counter2, err := testClient.Decr(ctx, "decr_counter2").Result()
	assert.NoError(t, err)
	assert.Equal(t, counter2, int64(9))

	<-time.After(1 * time.Second)

	counter2, err = testClient.Decr(ctx, "decr_counter2").Result()
	assert.NoError(t, err)
	assert.Equal(t, counter2, int64(-1))
}

func TestDecr_NonNumericResultInError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.SetArgs(ctx, "decr_counter3", "hello", redis.SetArgs{TTL: 1 * time.Second})
	_, err := testClient.Decr(ctx, "decr_counter3").Result()
	assert.Error(t, err)
}
