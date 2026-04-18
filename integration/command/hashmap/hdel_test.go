package hashmap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHDel_DeleteSingleField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hdel_hash1", "field1", "value1").Result()
	assert.NoError(t, err)

	n, err := testClient.HDel(ctx, "hdel_hash1", "field1").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestHDel_DeleteMultipleFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hdel_hash2", "field1", "value1", "field2", "value2", "field3", "value3").Result()
	assert.NoError(t, err)

	n, err := testClient.HDel(ctx, "hdel_hash2", "field1", "field2").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), n)
}

func TestHDel_DeleteNonExistentField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hdel_hash3", "field1", "value1").Result()
	assert.NoError(t, err)

	n, err := testClient.HDel(ctx, "hdel_hash3", "missing").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestHDel_DeleteMixedExistingAndNonExistentFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hdel_hash4", "field1", "value1", "field2", "value2").Result()
	assert.NoError(t, err)

	n, err := testClient.HDel(ctx, "hdel_hash4", "field1", "missing").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestHDel_DeleteFromNonExistentHash(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	n, err := testClient.HDel(ctx, "hdel_nonexistent", "field1").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestHDel_FieldNoLongerAccessibleAfterDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hdel_hash5", "field1", "value1").Result()
	assert.NoError(t, err)

	_, err = testClient.HDel(ctx, "hdel_hash5", "field1").Result()
	assert.NoError(t, err)

	val, err := testClient.HGet(ctx, "hdel_hash5", "field1").Result()
	assert.Error(t, err)
	assert.Empty(t, val)
}
