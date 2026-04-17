package hashmap

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// TestHGet_GetExistingField verifies that getting a field that was previously set
// returns the correct value.
func TestHGet_GetExistingField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hget_hash1", "field1", "value1").Result()
	assert.NoError(t, err)

	val, err := testClient.HGet(ctx, "hget_hash1", "field1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)
}

// TestHGet_GetUpdatedField verifies that after updating a field, HGet returns
// the latest value.
func TestHGet_GetUpdatedField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hget_hash2", "field1", "original").Result()
	assert.NoError(t, err)

	_, err = testClient.HSet(ctx, "hget_hash2", "field1", "updated").Result()
	assert.NoError(t, err)

	val, err := testClient.HGet(ctx, "hget_hash2", "field1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "updated", val)
}

// TestHGet_GetNonExistentField verifies that getting a field that does not exist
// on an existing hash returns a nil (null bulk string) response.
func TestHGet_GetNonExistentField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hget_hash3", "field1", "value1").Result()
	assert.NoError(t, err)

	_, err = testClient.HGet(ctx, "hget_hash3", "missing").Result()
	assert.Equal(t, redis.Nil, err)
}

// TestHGet_GetFieldOnNonExistentHash verifies that getting a field from a hash
// that does not exist returns a nil (null bulk string) response.
func TestHGet_GetFieldOnNonExistentHash(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HGet(ctx, "hget_nonexistent", "field1").Result()
	assert.Equal(t, redis.Nil, err)
}

// TestHGet_GetMultipleFieldsFromSameHash verifies that distinct fields on the
// same hash each return their own value correctly.
func TestHGet_GetMultipleFieldsFromSameHash(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hget_hash4", "f1", "v1", "f2", "v2", "f3", "v3").Result()
	assert.NoError(t, err)

	val1, err := testClient.HGet(ctx, "hget_hash4", "f1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "v1", val1)

	val2, err := testClient.HGet(ctx, "hget_hash4", "f2").Result()
	assert.NoError(t, err)
	assert.Equal(t, "v2", val2)

	val3, err := testClient.HGet(ctx, "hget_hash4", "f3").Result()
	assert.NoError(t, err)
	assert.Equal(t, "v3", val3)
}

