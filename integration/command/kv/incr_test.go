package kv

import (
	"avacado/integration"
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var testClient *redis.Client

func TestMain(m *testing.M) {
	shutdown, err := integration.StartNewServer(6001)
	if err != nil {
		panic(err)
	}

	testClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6001",
		Password: "",
		DB:       0,
	})

	code := m.Run()

	if err := testClient.Close(); err != nil {
		panic(err)
	}
	shutdown()
	os.Exit(code)
}

func TestIncr_IncrNonExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	counter1, err := testClient.Incr(ctx, "counter1").Result()
	assert.NoError(t, err)
	assert.Equal(t, counter1, int64(1))

	counter1, err = testClient.Incr(ctx, "counter1").Result()
	assert.NoError(t, err)
	assert.Equal(t, counter1, int64(2))
}

func TestIncr_IncrExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.SetArgs(ctx, "counter2", "10", redis.SetArgs{TTL: 1 * time.Second})
	counter2, err := testClient.Incr(ctx, "counter2").Result()
	assert.NoError(t, err)
	assert.Equal(t, counter2, int64(11))

	<-time.After(1 * time.Second)

	counter2, err = testClient.Incr(ctx, "counter2").Result()
	assert.NoError(t, err)
	assert.Equal(t, counter2, int64(1))
}

func TestIncr_NonNumericResultInError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.SetArgs(ctx, "counter3", "hello", redis.SetArgs{TTL: 1 * time.Second})
	_, err := testClient.Incr(ctx, "counter3").Result()
	assert.Error(t, err)
}
