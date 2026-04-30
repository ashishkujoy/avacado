package hashmap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHExists_FieldExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hexists_hash1", "field1", "value1").Result()
	assert.NoError(t, err)

	exists, err := testClient.HExists(ctx, "hexists_hash1", "field1").Result()
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestHExists_FieldNotExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hexists_hash2", "field1", "value1").Result()
	assert.NoError(t, err)

	exists, err := testClient.HExists(ctx, "hexists_hash2", "missing").Result()
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestHExists_KeyNotExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	exists, err := testClient.HExists(ctx, "hexists_nonexistent", "field1").Result()
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestHExists_MultipleFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hexists_hash3", "field1", "value1", "field2", "value2").Result()
	assert.NoError(t, err)

	exists1, err := testClient.HExists(ctx, "hexists_hash3", "field1").Result()
	assert.NoError(t, err)
	assert.True(t, exists1)

	exists2, err := testClient.HExists(ctx, "hexists_hash3", "field2").Result()
	assert.NoError(t, err)
	assert.True(t, exists2)

	exists3, err := testClient.HExists(ctx, "hexists_hash3", "field3").Result()
	assert.NoError(t, err)
	assert.False(t, exists3)
}

func TestHExists_AfterDeletion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.HSet(ctx, "hexists_hash4", "field1", "value1").Result()
	assert.NoError(t, err)

	exists, err := testClient.HExists(ctx, "hexists_hash4", "field1").Result()
	assert.NoError(t, err)
	assert.True(t, exists)

	_, err = testClient.HDel(ctx, "hexists_hash4", "field1").Result()
	assert.NoError(t, err)

	exists, err = testClient.HExists(ctx, "hexists_hash4", "field1").Result()
	assert.NoError(t, err)
	assert.False(t, exists)
}
