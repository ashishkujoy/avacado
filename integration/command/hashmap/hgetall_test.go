package hashmap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHGetAll_ExistingHash verifies that HGetAll returns all fields and values
// for a hash that was previously set.
func TestHGetAll_ExistingHash(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hgetall_hash1", "field1", "value1", "field2", "value2").Result()
	assert.NoError(t, err)

	result, err := testClient.HGetAll(ctx, "hgetall_hash1").Result()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"field1": "value1",
		"field2": "value2",
	}, result)
}

// TestHGetAll_SingleField verifies that HGetAll works correctly for a hash
// with only one field.
func TestHGetAll_SingleField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hgetall_hash2", "only_field", "only_value").Result()
	assert.NoError(t, err)

	result, err := testClient.HGetAll(ctx, "hgetall_hash2").Result()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"only_field": "only_value"}, result)
}

// TestHGetAll_NonExistentHash verifies that HGetAll returns an empty map
// when the hash does not exist.
func TestHGetAll_NonExistentHash(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	result, err := testClient.HGetAll(ctx, "hgetall_nonexistent").Result()
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// TestHGetAll_AfterUpdate verifies that HGetAll reflects the latest values
// after a field has been updated.
func TestHGetAll_AfterUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hgetall_hash3", "field1", "original").Result()
	assert.NoError(t, err)

	_, err = testClient.HSet(ctx, "hgetall_hash3", "field1", "updated").Result()
	assert.NoError(t, err)

	result, err := testClient.HGetAll(ctx, "hgetall_hash3").Result()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"field1": "updated"}, result)
}

// TestHGetAll_MultipleFields verifies that HGetAll returns all fields correctly
// when a hash has many fields set at once.
func TestHGetAll_MultipleFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hgetall_hash4",
		"f1", "v1",
		"f2", "v2",
		"f3", "v3",
		"f4", "v4",
	).Result()
	assert.NoError(t, err)

	result, err := testClient.HGetAll(ctx, "hgetall_hash4").Result()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"f1": "v1",
		"f2": "v2",
		"f3": "v3",
		"f4": "v4",
	}, result)
}

