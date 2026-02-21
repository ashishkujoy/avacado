package list

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
	shutdown, err := integration.StartNewServer(6002)
	if err != nil {
		panic(err)
	}

	testClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6002",
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

func TestRPush_PushToNewList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	count, err := testClient.RPush(ctx, "list1", "a", "b", "c").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestRPush_PushToExistingList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	count, err := testClient.RPush(ctx, "list2", "a").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = testClient.RPush(ctx, "list2", "b", "c").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestLLen_ExistingList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "list3", "a", "b", "c")

	length, err := testClient.LLen(ctx, "list3").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), length)
}

func TestLLen_NonExistingList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	length, err := testClient.LLen(ctx, "nonexistent").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), length)
}

func TestRPop_SingleElement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "list4", "a", "b", "c")

	val, err := testClient.RPop(ctx, "list4").Result()
	assert.NoError(t, err)
	assert.Equal(t, "c", val)
}

func TestRPop_NonExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.RPop(ctx, "nonexistent").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestRPop_WithCount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "list5", "a", "b", "c", "d")

	vals, err := testClient.RPopCount(ctx, "list5", 2).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"d", "c"}, vals)
}

func TestLPush_PushToNewList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	count, err := testClient.LPush(ctx, "lpush1", "a", "b", "c").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestLPush_ElementOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// LPUSH mylist a b c -> list is [c, b, a]
	testClient.LPush(ctx, "lpush2", "a", "b", "c")

	// RPop should return from the tail: a, then b, then c
	val, err := testClient.RPop(ctx, "lpush2").Result()
	assert.NoError(t, err)
	assert.Equal(t, "a", val)

	val, err = testClient.RPop(ctx, "lpush2").Result()
	assert.NoError(t, err)
	assert.Equal(t, "b", val)

	val, err = testClient.RPop(ctx, "lpush2").Result()
	assert.NoError(t, err)
	assert.Equal(t, "c", val)
}

func TestLPush_PushToExistingList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	count, err := testClient.RPush(ctx, "lpush3", "a").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = testClient.LPush(ctx, "lpush3", "b").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// List should be [b, a], RPop returns a
	val, err := testClient.RPop(ctx, "lpush3").Result()
	assert.NoError(t, err)
	assert.Equal(t, "a", val)

	val, err = testClient.RPop(ctx, "lpush3").Result()
	assert.NoError(t, err)
	assert.Equal(t, "b", val)
}

func TestLPop_SingleElement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lpop1", "a", "b", "c")

	val, err := testClient.LPop(ctx, "lpop1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "a", val)
}

func TestLPop_NonExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.LPop(ctx, "lpop_nonexistent").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestLPop_WithCount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lpop2", "a", "b", "c", "d")

	vals, err := testClient.LPopCount(ctx, "lpop2", 2).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, vals)
}

func TestLIndex_ExistingElement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lindex1", "a", "b", "c")

	val, err := testClient.LIndex(ctx, "lindex1", 0).Result()
	assert.NoError(t, err)
	assert.Equal(t, "a", val)

	val, err = testClient.LIndex(ctx, "lindex1", 1).Result()
	assert.NoError(t, err)
	assert.Equal(t, "b", val)

	val, err = testClient.LIndex(ctx, "lindex1", 2).Result()
	assert.NoError(t, err)
	assert.Equal(t, "c", val)
}

func TestLIndex_NegativeIndex(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lindex2", "a", "b", "c")

	// -1 is the last element
	val, err := testClient.LIndex(ctx, "lindex2", -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, "c", val)

	// -2 is the second to last element
	val, err = testClient.LIndex(ctx, "lindex2", -2).Result()
	assert.NoError(t, err)
	assert.Equal(t, "b", val)

	// -3 is the first element
	val, err = testClient.LIndex(ctx, "lindex2", -3).Result()
	assert.NoError(t, err)
	assert.Equal(t, "a", val)
}

func TestLIndex_OutOfRange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lindex3", "a", "b", "c")

	_, err := testClient.LIndex(ctx, "lindex3", 10).Result()
	assert.Equal(t, redis.Nil, err)

	_, err = testClient.LIndex(ctx, "lindex3", -10).Result()
	assert.Equal(t, redis.Nil, err)
}

func TestLIndex_NonExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.LIndex(ctx, "lindex_nonexistent", 0).Result()
	assert.Equal(t, redis.Nil, err)
}

func TestList_PushPopLen(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Push elements
	count, err := testClient.RPush(ctx, "list6", "one", "two", "three").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Verify length
	length, err := testClient.LLen(ctx, "list6").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), length)

	// Pop last element
	val, err := testClient.RPop(ctx, "list6").Result()
	assert.NoError(t, err)
	assert.Equal(t, "three", val)

	// Verify length decreased
	length, err = testClient.LLen(ctx, "list6").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), length)

	// Pop remaining elements
	val, err = testClient.RPop(ctx, "list6").Result()
	assert.NoError(t, err)
	assert.Equal(t, "two", val)

	val, err = testClient.RPop(ctx, "list6").Result()
	assert.NoError(t, err)
	assert.Equal(t, "one", val)

	// List should be empty / gone
	length, err = testClient.LLen(ctx, "list6").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), length)

	// Pop from empty list returns nil
	_, err = testClient.RPop(ctx, "list6").Result()
	assert.Equal(t, redis.Nil, err)
}
