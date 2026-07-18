package kv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrlen_ExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "strlen_existing", "Hello World", 0)

	n, err := testClient.StrLen(ctx, "strlen_existing").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(11), n)
}

func TestStrlen_NonExistentKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	n, err := testClient.StrLen(ctx, "strlen_missing").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestStrlen_IntegerEncodedKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.Set(ctx, "strlen_int_key", "12345", 0)

	n, err := testClient.StrLen(ctx, "strlen_int_key").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(5), n)
}
