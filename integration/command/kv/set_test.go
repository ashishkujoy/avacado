package kv

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestSet_WithIFEQMatchingValue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set initial value
	err := testClient.Set(ctx, "ifeq_key1", "oldvalue", 0).Err()
	assert.NoError(t, err)

	// Use Do to call SET with IFEQ option
	result, err := testClient.Do(ctx, "SET", "ifeq_key1", "newvalue", "IFEQ", "oldvalue").Result()
	assert.NoError(t, err)
	assert.Equal(t, "OK", result)

	// Verify the value was updated
	val, err := testClient.Get(ctx, "ifeq_key1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "newvalue", val)
}

func TestSet_WithIFEQNonMatchingValue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set initial value
	err := testClient.Set(ctx, "ifeq_key2", "oldvalue", 0).Err()
	assert.NoError(t, err)

	// Try to set with non-matching IFEQ value
	_, err = testClient.Do(ctx, "SET", "ifeq_key2", "newvalue", "IFEQ", "wrongvalue").Result()
	assert.Error(t, err)

	// Verify the value was not updated
	val, err := testClient.Get(ctx, "ifeq_key2").Result()
	assert.NoError(t, err)
	assert.Equal(t, "oldvalue", val)
}

func TestSet_WithIFEQNonExistentKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Try to set with IFEQ on non-existent key
	_, err := testClient.Do(ctx, "SET", "ifeq_nonexistent", "newvalue", "IFEQ", "somevalue").Result()
	assert.Error(t, err)

	// Verify the key was not created
	_, err = testClient.Get(ctx, "ifeq_nonexistent").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestSet_WithIFEQExpiredKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set a key with short expiry
	err := testClient.SetArgs(ctx, "ifeq_expiring", "oldvalue", redis.SetArgs{TTL: 1 * time.Second}).Err()
	assert.NoError(t, err)

	// Wait for expiry
	time.Sleep(1100 * time.Millisecond)

	// Try to set with IFEQ on expired key (should be treated as non-existent)
	_, err = testClient.Do(ctx, "SET", "ifeq_expiring", "newvalue", "IFEQ", "oldvalue").Result()
	assert.Error(t, err)

	// Verify the key is still expired/non-existent
	_, err = testClient.Get(ctx, "ifeq_expiring").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestSet_WithIFEQAndGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set initial value
	err := testClient.Set(ctx, "ifeq_key3", "oldvalue", 0).Err()
	assert.NoError(t, err)

	// Set with IFEQ and GET option
	result, err := testClient.Do(ctx, "SET", "ifeq_key3", "newvalue", "IFEQ", "oldvalue", "GET").Result()
	assert.NoError(t, err)
	assert.Equal(t, "oldvalue", result)

	// Verify the value was updated
	val, err := testClient.Get(ctx, "ifeq_key3").Result()
	assert.NoError(t, err)
	assert.Equal(t, "newvalue", val)
}

func TestSet_WithIFEQAndEX(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set initial value
	err := testClient.Set(ctx, "ifeq_key4", "oldvalue", 0).Err()
	assert.NoError(t, err)

	// Set with IFEQ and EX option
	result, err := testClient.Do(ctx, "SET", "ifeq_key4", "newvalue", "IFEQ", "oldvalue", "EX", 2).Result()
	assert.NoError(t, err)
	assert.Equal(t, "OK", result)

	// Verify the value was updated
	val, err := testClient.Get(ctx, "ifeq_key4").Result()
	assert.NoError(t, err)
	assert.Equal(t, "newvalue", val)

	// Wait for expiry
	time.Sleep(2100 * time.Millisecond)

	// Verify the key expired
	_, err = testClient.Get(ctx, "ifeq_key4").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestSet_WithIFEQIntegerValues(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set initial integer value
	err := testClient.Set(ctx, "ifeq_key5", "123", 0).Err()
	assert.NoError(t, err)

	// Set with IFEQ matching integer value
	result, err := testClient.Do(ctx, "SET", "ifeq_key5", "456", "IFEQ", "123").Result()
	assert.NoError(t, err)
	assert.Equal(t, "OK", result)

	// Verify the value was updated
	val, err := testClient.Get(ctx, "ifeq_key5").Result()
	assert.NoError(t, err)
	assert.Equal(t, "456", val)
}
