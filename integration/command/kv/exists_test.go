package kv

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestExists_SingleExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set a key
	testClient.Set(ctx, "exists_key1", "value1", 0)

	// Check if it exists
	count, err := testClient.Exists(ctx, "exists_key1").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestExists_SingleNonExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Check non-existing key
	count, err := testClient.Exists(ctx, "exists_nonexistent").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestExists_MultipleExistingKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set multiple keys
	testClient.Set(ctx, "exists_key2", "value2", 0)
	testClient.Set(ctx, "exists_key3", "value3", 0)
	testClient.Set(ctx, "exists_key4", "value4", 0)

	// Check if they exist
	count, err := testClient.Exists(ctx, "exists_key2", "exists_key3", "exists_key4").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestExists_MixOfExistingAndNonExisting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set some keys
	testClient.Set(ctx, "exists_key5", "value5", 0)
	testClient.Set(ctx, "exists_key6", "value6", 0)

	// Check mix of existing and non-existing keys
	count, err := testClient.Exists(ctx, "exists_key5", "exists_nonexistent1", "exists_key6", "exists_nonexistent2").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestExists_ExpiredKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set a key with short TTL
	testClient.SetArgs(ctx, "exists_key7", "value7", redis.SetArgs{TTL: 1 * time.Second})

	// Check it exists before expiry
	count, err := testClient.Exists(ctx, "exists_key7").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Wait for it to expire
	<-time.After(1100 * time.Millisecond)

	// Check it doesn't exist after expiry
	count, err = testClient.Exists(ctx, "exists_key7").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestExists_DuplicateKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set a key
	testClient.Set(ctx, "exists_key8", "value8", 0)

	// Check the same key multiple times - should count each mention
	count, err := testClient.Exists(ctx, "exists_key8", "exists_key8", "exists_key8").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestExists_DuplicateKeysWithNonExisting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set a key
	testClient.Set(ctx, "exists_key9", "value9", 0)

	// Check duplicate keys with some non-existing
	count, err := testClient.Exists(ctx, "exists_key9", "exists_nonexistent3", "exists_key9").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestExists_AllNonExisting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Check multiple non-existing keys
	count, err := testClient.Exists(ctx, "exists_nonexistent4", "exists_nonexistent5", "exists_nonexistent6").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestExists_AfterDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set a key
	testClient.Set(ctx, "exists_key10", "value10", 0)

	// Check it exists
	count, err := testClient.Exists(ctx, "exists_key10").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete the key
	testClient.Del(ctx, "exists_key10")

	// Check it doesn't exist anymore
	count, err = testClient.Exists(ctx, "exists_key10").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
