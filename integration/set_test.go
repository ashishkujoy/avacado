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

	err := testClient.Set(ctx, "key1", "value", 0).Err()
	assert.NoError(t, err)

	val, err := testClient.Get(ctx, "key1").Result()
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

	err := testClient.Set(ctx, "key2", "value", 1*time.Second).Err()
	assert.NoError(t, err)

	val, err := testClient.Get(ctx, "key2").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	<-time.After(time.Second * 1)

	_, err = testClient.Get(ctx, "key2").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestSet_NXOption(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		testClient.FlushDB(ctx)
	})

	err := testClient.Set(ctx, "key3", "value", -1).Err()
	assert.NoError(t, err)

	// Should not update since key3 already exists
	err = testClient.SetNX(ctx, "key3", "newvalue", -1).Err()
	assert.NoError(t, err)

	val, err := testClient.Get(ctx, "key3").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	// Should set since key4 does not exist
	err = testClient.SetNX(ctx, "key4", "value2", -1).Err()
	assert.NoError(t, err)

	val, err = testClient.Get(ctx, "key4").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)
}

func TestSet_XXOption(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		testClient.FlushDB(ctx)
	})

	// Should not set since key1 does not exist
	err := testClient.SetXX(ctx, "key5", "value", -1).Err()
	assert.NoError(t, err)

	err = testClient.Set(ctx, "key5", "old value", -1).Err()
	assert.NoError(t, err)

	// Should update since key1 already exists
	err = testClient.SetXX(ctx, "key5", "new value", -1).Err()
	assert.NoError(t, err)

	val, err := testClient.Get(ctx, "key5").Result()
	assert.NoError(t, err)
	assert.Equal(t, "new value", val)
}

func TestSet_GetOption(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
		testClient.FlushDB(ctx)
	})

	err := testClient.Set(ctx, "key6", "value1", -1).Err()
	assert.NoError(t, err)

	// Get old value while setting new value
	oldVal, err := testClient.SetArgs(ctx, "key6", "value2", redis.SetArgs{Get: true}).Result()
	assert.NoError(t, err)
	assert.Equal(t, "value1", oldVal)

	// Get old value while setting new value for non-existent key
	oldVal, err = testClient.SetArgs(ctx, "key7", "value3", redis.SetArgs{Get: true}).Result()
	assert.Equal(t, redis.Nil, err)
	assert.Equal(t, "", oldVal)
}
