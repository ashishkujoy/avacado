package integration

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func Test_SetAndGetAKey(t *testing.T) {
	shutdown, err := StartNewServer(6000)
	assert.NoError(t, err)
	defer shutdown()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6000",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()
	var ctx = context.Background()
	err = rdb.Set(ctx, "key", "value", 0).Err()
	assert.NoError(t, err)

	val, err := rdb.Get(ctx, "key").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	_, err = rdb.Get(ctx, "nonexistent").Result()
	assert.Equal(t, redis.Nil, err)
}
