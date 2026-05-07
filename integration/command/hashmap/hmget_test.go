package hashmap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHMGet_AllFieldsExist verifies that HMGET returns values in order for all existing fields.
func TestHMGet_AllFieldsExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hmget_hash1", "field1", "v1", "field2", "v2").Result()
	assert.NoError(t, err)

	vals, err := testClient.HMGet(ctx, "hmget_hash1", "field1", "field2").Result()
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"v1", "v2"}, vals)
}

// TestHMGet_SomeMissing verifies that missing fields return nil at the correct positions.
func TestHMGet_SomeMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hmget_hash2", "field1", "v1").Result()
	assert.NoError(t, err)

	vals, err := testClient.HMGet(ctx, "hmget_hash2", "field1", "nofield").Result()
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"v1", nil}, vals)
}

// TestHMGet_KeyNotFound verifies that a non-existent key returns all nil values.
func TestHMGet_KeyNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	vals, err := testClient.HMGet(ctx, "hmget_nonexistent", "field1").Result()
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{nil}, vals)
}

// TestHMGet_SingleField verifies that HMGET works correctly with exactly one field.
func TestHMGet_SingleField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hmget_hash3", "only", "val").Result()
	assert.NoError(t, err)

	vals, err := testClient.HMGet(ctx, "hmget_hash3", "only").Result()
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"val"}, vals)
}

// TestHMGet_DuplicateFields verifies that requesting the same field twice returns its value at both positions.
func TestHMGet_DuplicateFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hmget_hash4", "f1", "v1").Result()
	assert.NoError(t, err)

	vals, err := testClient.HMGet(ctx, "hmget_hash4", "f1", "f1").Result()
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"v1", "v1"}, vals)
}
