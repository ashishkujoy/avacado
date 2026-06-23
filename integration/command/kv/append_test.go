package kv

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAppend_NonExistentKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	n, err := testClient.Append(ctx, "append_new_key", "Hello").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(5), n)

	val, err := testClient.Get(ctx, "append_new_key").Result()
	assert.NoError(t, err)
	assert.Equal(t, "Hello", val)
}

func TestAppend_ExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "append_existing", "Hello", 0)

	n, err := testClient.Append(ctx, "append_existing", " World").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(11), n)

	val, err := testClient.Get(ctx, "append_existing").Result()
	assert.NoError(t, err)
	assert.Equal(t, "Hello World", val)
}

func TestAppend_IntegerEncodedKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "append_int_key", "10", 0)

	n, err := testClient.Append(ctx, "append_int_key", " items").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(8), n)

	val, err := testClient.Get(ctx, "append_int_key").Result()
	assert.NoError(t, err)
	assert.Equal(t, "10 items", val)
}

func TestAppend_TTLPreserved(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "append_ttl_key", "hello", 10*time.Second)

	n, err := testClient.Append(ctx, "append_ttl_key", "!").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(6), n)

	ttl, err := testClient.TTL(ctx, "append_ttl_key").Result()
	assert.NoError(t, err)
	assert.Positive(t, ttl)
}
