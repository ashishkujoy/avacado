package kv

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestDel_DeleteSingleExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set a key
	testClient.Set(ctx, "del_key1", "value1", 0)

	// Delete it
	count, err := testClient.Del(ctx, "del_key1").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify it's deleted
	val, err := testClient.Get(ctx, "del_key1").Result()
	assert.Error(t, err) // redis.Nil error
	assert.Equal(t, "", val)
}

func TestDel_DeleteMultipleExistingKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set multiple keys
	testClient.Set(ctx, "del_key2", "value2", 0)
	testClient.Set(ctx, "del_key3", "value3", 0)
	testClient.Set(ctx, "del_key4", "value4", 0)

	// Delete them
	count, err := testClient.Del(ctx, "del_key2", "del_key3", "del_key4").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Verify they're deleted
	val, err := testClient.Get(ctx, "del_key2").Result()
	assert.Error(t, err)
	assert.Equal(t, "", val)

	val, err = testClient.Get(ctx, "del_key3").Result()
	assert.Error(t, err)
	assert.Equal(t, "", val)

	val, err = testClient.Get(ctx, "del_key4").Result()
	assert.Error(t, err)
	assert.Equal(t, "", val)
}

func TestDel_DeleteNonExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Delete non-existing key
	count, err := testClient.Del(ctx, "del_nonexistent").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestDel_DeleteMixOfExistingAndNonExisting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set one key
	testClient.Set(ctx, "del_key5", "value5", 0)

	// Delete mix of existing and non-existing keys
	count, err := testClient.Del(ctx, "del_key5", "del_nonexistent1", "del_nonexistent2").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count) // Only del_key5 was deleted

	// Verify the existing key was deleted
	val, err := testClient.Get(ctx, "del_key5").Result()
	assert.Error(t, err)
	assert.Equal(t, "", val)
}

func TestDel_DeleteExpiredKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set a key with short TTL
	testClient.SetArgs(ctx, "del_key6", "value6", redis.SetArgs{TTL: 1 * time.Second})

	// Wait for it to expire
	<-time.After(1100 * time.Millisecond)

	// Try to delete it - should return 0 since it's already expired
	count, err := testClient.Del(ctx, "del_key6").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
