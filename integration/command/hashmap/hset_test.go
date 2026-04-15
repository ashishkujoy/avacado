package hashmap

import (
	"avacado/integration"
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var testClient *redis.Client

func TestMain(m *testing.M) {
	shutdown, err := integration.StartNewServer(6003)
	if err != nil {
		panic(err)
	}

	testClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6003",
		Password: "",
		DB:       0,
	})

	code := m.Run()

	if err := testClient.Close(); err != nil {
		panic(err)
	}
	shutdown()
	os.Exit(code)
}

// TestHSet_SetSingleField verifies that setting a single field on a new hash
// returns 1 (the number of new fields added).
func TestHSet_SetSingleField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	n, err := testClient.HSet(ctx, "hash1", "field1", "value1").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

// TestHSet_SetMultipleFields verifies that setting multiple fields on a new hash
// returns the count of all fields added.
func TestHSet_SetMultipleFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	n, err := testClient.HSet(ctx, "hash2", "field1", "value1", "field2", "value2", "field3", "value3").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), n)
}

// TestHSet_UpdateExistingField verifies that updating an already-existing field
// returns 0, since no new fields were added.
func TestHSet_UpdateExistingField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create the field first
	n, err := testClient.HSet(ctx, "hash3", "field1", "original").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)

	// Update the same field — no new fields, so count should be 0
	n, err = testClient.HSet(ctx, "hash3", "field1", "updated").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

// TestHSet_MixedNewAndExistingFields verifies that when setting a mix of new and
// existing fields, only the count of new fields is returned.
func TestHSet_MixedNewAndExistingFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Seed two fields
	_, err := testClient.HSet(ctx, "hash4", "field1", "value1", "field2", "value2").Result()
	assert.NoError(t, err)

	// Update field1 (existing) and add field3 (new) — only 1 new field
	n, err := testClient.HSet(ctx, "hash4", "field1", "newvalue1", "field3", "value3").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

