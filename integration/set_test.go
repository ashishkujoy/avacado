package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var testClient *redis.Client

func TestMain(m *testing.M) {
	shutdown, err := StartNewServer(6000)
	if err != nil {
		panic(err)
	}

	testClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6000",
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

func Test_SetAndGetAKey(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		testClient.FlushDB(ctx)
	})

	err := testClient.Set(ctx, "key", "value", 0).Err()
	assert.NoError(t, err)

	val, err := testClient.Get(ctx, "key").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	_, err = testClient.Get(ctx, "nonexistent").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestSet_Expiry(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		testClient.FlushDB(ctx)
	})

	err := testClient.Set(ctx, "key", "value", 1*time.Second).Err()
	assert.NoError(t, err)

	val, err := testClient.Get(ctx, "key").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	<-time.After(time.Second * 1)

	_, err = testClient.Get(ctx, "key").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestSet_NXOption(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		testClient.FlushDB(ctx)
	})

	err := testClient.Set(ctx, "key1", "value", -1).Err()
	assert.NoError(t, err)

	// Should not update since key1 already exists
	err = testClient.SetNX(ctx, "key1", "newvalue", -1).Err()
	assert.NoError(t, err)

	val, err := testClient.Get(ctx, "key1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	// Should set since key2 does not exist
	err = testClient.SetNX(ctx, "key2", "value2", -1).Err()
	assert.NoError(t, err)

	val, err = testClient.Get(ctx, "key2").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)
}

func TestSet_XXOption(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		testClient.FlushDB(ctx)
	})

	// Should not set since key1 does not exist
	err := testClient.SetXX(ctx, "key1", "value", -1).Err()
	assert.NoError(t, err)

	err = testClient.Set(ctx, "key1", "old value", -1).Err()
	assert.NoError(t, err)

	// Should update since key1 already exists
	err = testClient.SetXX(ctx, "key1", "new value", -1).Err()
	assert.NoError(t, err)

	val, err := testClient.Get(ctx, "key1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "new value", val)
}
