package kv

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetRange_NonExistentKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	n, err := testClient.SetRange(ctx, "setrange_new_key", 6, "Redis").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(11), n)

	val, err := testClient.Get(ctx, "setrange_new_key").Result()
	assert.NoError(t, err)
	assert.Equal(t, "\x00\x00\x00\x00\x00\x00Redis", val)
}

func TestSetRange_ExistingKeyWithinBounds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "setrange_existing", "Hello World", 0)

	n, err := testClient.SetRange(ctx, "setrange_existing", 6, "Redis").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(11), n)

	val, err := testClient.Get(ctx, "setrange_existing").Result()
	assert.NoError(t, err)
	assert.Equal(t, "Hello Redis", val)
}

func TestSetRange_OffsetBeyondLengthPadsWithZeroBytes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "setrange_padding", "Hello", 0)

	n, err := testClient.SetRange(ctx, "setrange_padding", 10, "World").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(15), n)

	val, err := testClient.Get(ctx, "setrange_padding").Result()
	assert.NoError(t, err)
	assert.Equal(t, "Hello\x00\x00\x00\x00\x00World", val)
}

func TestSetRange_EmptyValueIsNoop(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "setrange_empty_value", "Hello World", 0)

	n, err := testClient.SetRange(ctx, "setrange_empty_value", 6, "").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(11), n)

	val, err := testClient.Get(ctx, "setrange_empty_value").Result()
	assert.NoError(t, err)
	assert.Equal(t, "Hello World", val)
}

func TestSetRange_TTLPreserved(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "setrange_ttl_key", "Hello World", 10*time.Second)

	n, err := testClient.SetRange(ctx, "setrange_ttl_key", 6, "Redis").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(11), n)

	ttl, err := testClient.TTL(ctx, "setrange_ttl_key").Result()
	assert.NoError(t, err)
	assert.Positive(t, ttl)
}
