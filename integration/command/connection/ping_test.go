package connection

import (
	"avacado/integration"
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var testClient *redis.Client

func TestMain(m *testing.M) {
	shutdown, err := integration.StartNewServer(6005)
	if err != nil {
		panic(err)
	}

	testClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6005",
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

func TestPing_NoMessage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	result, err := testClient.Ping(ctx).Result()

	assert.NoError(t, err)
	assert.Equal(t, "PONG", result)
}

func TestPing_WithMessage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	result, err := testClient.Do(ctx, "PING", "hello").Text()

	assert.NoError(t, err)
	assert.Equal(t, "hello", result)
}
