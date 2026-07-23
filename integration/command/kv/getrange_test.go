package kv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRange_ExistingKeyPositiveRange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "getrange_key1", "Hello World", 0)

	val, err := testClient.GetRange(ctx, "getrange_key1", 0, 4).Result()
	assert.NoError(t, err)
	assert.Equal(t, "Hello", val)
}

func TestGetRange_NegativeIndices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "getrange_key2", "Hello World", 0)

	val, err := testClient.GetRange(ctx, "getrange_key2", -5, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, "World", val)
}

func TestGetRange_NonExistentKeyReturnsEmptyString(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	val, err := testClient.GetRange(ctx, "getrange_missing_key", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, "", val)
}

func TestGetRange_EndBeyondLengthClamps(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "getrange_key3", "Hello World", 0)

	val, err := testClient.GetRange(ctx, "getrange_key3", 6, 100).Result()
	assert.NoError(t, err)
	assert.Equal(t, "World", val)
}

func TestSubstr_AliasMatchesGetRange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "getrange_key4", "Hello World", 0)

	getRangeVal, err := testClient.GetRange(ctx, "getrange_key4", -5, -1).Result()
	assert.NoError(t, err)

	substrResult, err := testClient.Do(ctx, "SUBSTR", "getrange_key4", -5, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, getRangeVal, substrResult)
}
