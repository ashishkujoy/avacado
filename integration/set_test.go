package integration

import (
	"context"
	"testing"
	"time"

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

func TestSet_Expiry(t *testing.T) {
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
	err = rdb.Set(ctx, "key", "value", 1*time.Second).Err()
	assert.NoError(t, err)

	val, err := rdb.Get(ctx, "key").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	<-time.After(time.Second * 1)

	_, err = rdb.Get(ctx, "key").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestSet_NXOption(t *testing.T) {
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
	err = rdb.Set(ctx, "key1", "value", -1).Err()
	assert.NoError(t, err)

	// Should not update since key1 already exists
	err = rdb.SetNX(ctx, "key1", "newvalue", -1).Err()
	assert.NoError(t, err)

	val, err := rdb.Get(ctx, "key1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	// Should set since key2 does not exist
	err = rdb.SetNX(ctx, "key2", "value2", -1).Err()
	assert.NoError(t, err)

	val, err = rdb.Get(ctx, "key2").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)
}

func TestSet_XXOption(t *testing.T) {
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

	// Should not set since key1 does not exist
	err = rdb.SetXX(ctx, "key1", "value", -1).Err()
	assert.NoError(t, err)

	err = rdb.Set(ctx, "key1", "old value", -1).Err()
	assert.NoError(t, err)

	// Should update since key1 already exists
	err = rdb.SetXX(ctx, "key1", "new value", -1).Err()
	assert.NoError(t, err)

	val, err := rdb.Get(ctx, "key1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "new value", val)
}
