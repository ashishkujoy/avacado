package kv

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestDecrBy_DecrNonExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	counter1, err := testClient.DecrBy(ctx, "decrby_counter1", 5).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(-5), counter1)

	counter1, err = testClient.DecrBy(ctx, "decrby_counter1", 3).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(-8), counter1)
}

func TestDecrBy_DecrExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.SetArgs(ctx, "decrby_counter2", "100", redis.SetArgs{TTL: 2 * time.Second})
	counter2, err := testClient.DecrBy(ctx, "decrby_counter2", 20).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(80), counter2)

	counter2, err = testClient.DecrBy(ctx, "decrby_counter2", 30).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(50), counter2)
}

func TestDecrBy_DecrExpiredKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.SetArgs(ctx, "decrby_counter3", "50", redis.SetArgs{TTL: 1 * time.Second})
	counter3, err := testClient.DecrBy(ctx, "decrby_counter3", 10).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(40), counter3)

	<-time.After(1 * time.Second)

	counter3, err = testClient.DecrBy(ctx, "decrby_counter3", 15).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(-15), counter3)
}

func TestDecrBy_NonNumericResultInError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.SetArgs(ctx, "decrby_counter4", "hello", redis.SetArgs{TTL: 1 * time.Second})
	_, err := testClient.DecrBy(ctx, "decrby_counter4", 5).Result()
	assert.Error(t, err)
}
