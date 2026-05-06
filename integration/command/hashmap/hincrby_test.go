package hashmap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHIncrBy_NonExistentField increments a field on a non-existent hash,
// initializing it to 0 and then adding the increment.
func TestHIncrBy_NonExistentField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	val, err := testClient.HIncrBy(ctx, "myhash1", "field1", 5).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(5), val)

	// Verify the value was set correctly
	getVal, err := testClient.HGet(ctx, "myhash1", "field1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "5", getVal)
}

// TestHIncrBy_ExistingField increments an existing field in a hash.
func TestHIncrBy_ExistingField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set initial value
	_, err := testClient.HSet(ctx, "myhash2", "field1", "10").Result()
	assert.NoError(t, err)

	// Increment the field
	val, err := testClient.HIncrBy(ctx, "myhash2", "field1", 5).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(15), val)
}

// TestHIncrBy_NegativeIncrement decrements using a negative increment.
func TestHIncrBy_NegativeIncrement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set initial value
	_, err := testClient.HSet(ctx, "myhash3", "field1", "20").Result()
	assert.NoError(t, err)

	// Decrement using negative increment
	val, err := testClient.HIncrBy(ctx, "myhash3", "field1", -5).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(15), val)
}

// TestHIncrBy_MultipleIncrements performs multiple increments on the same field.
func TestHIncrBy_MultipleIncrements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// First increment on non-existent field
	val1, err := testClient.HIncrBy(ctx, "myhash4", "counter", 1).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val1)

	// Second increment
	val2, err := testClient.HIncrBy(ctx, "myhash4", "counter", 2).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), val2)

	// Third increment
	val3, err := testClient.HIncrBy(ctx, "myhash4", "counter", 4).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(7), val3)
}

// TestHIncrBy_NegativeValue increments a field to a negative value.
func TestHIncrBy_NegativeValue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set initial negative value
	_, err := testClient.HSet(ctx, "myhash5", "field1", "-10").Result()
	assert.NoError(t, err)

	// Increment (add less than the absolute value of the negative)
	val, err := testClient.HIncrBy(ctx, "myhash5", "field1", 3).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(-7), val)
}

// TestHIncrBy_ZeroIncrement increments by zero.
func TestHIncrBy_ZeroIncrement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set initial value
	_, err := testClient.HSet(ctx, "myhash6", "field1", "42").Result()
	assert.NoError(t, err)

	// Increment by zero
	val, err := testClient.HIncrBy(ctx, "myhash6", "field1", 0).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(42), val)
}

// TestHIncrBy_NonIntegerField tries to increment a field with a non-integer value.
func TestHIncrBy_NonIntegerField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set a string value
	_, err := testClient.HSet(ctx, "myhash7", "field1", "notanumber").Result()
	assert.NoError(t, err)

	// Try to increment - should fail
	val, err := testClient.HIncrBy(ctx, "myhash7", "field1", 5).Result()
	assert.Error(t, err)
	assert.Equal(t, int64(0), val)
}

// TestHIncrBy_MultipleFields increments different fields in the same hash.
func TestHIncrBy_MultipleFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Increment multiple fields
	val1, err := testClient.HIncrBy(ctx, "myhash8", "field1", 10).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(10), val1)

	val2, err := testClient.HIncrBy(ctx, "myhash8", "field2", 20).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(20), val2)

	// Verify both fields
	all, err := testClient.HGetAll(ctx, "myhash8").Result()
	assert.NoError(t, err)
	assert.Equal(t, "10", all["field1"])
	assert.Equal(t, "20", all["field2"])
}

// TestHIncrBy_LargeValues tests with large integer values.
func TestHIncrBy_LargeValues(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set a large initial value
	largeVal := int64(9223372036854775700)
	_, err := testClient.HSet(ctx, "myhash9", "field1", "9223372036854775700").Result()
	assert.NoError(t, err)

	// Increment with a safe value
	val, err := testClient.HIncrBy(ctx, "myhash9", "field1", 100).Result()
	assert.NoError(t, err)
	assert.Equal(t, largeVal+100, val)
}
