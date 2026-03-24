package list

import (
	"avacado/integration"
	"context"
	"os"
	"testing"
	"time"

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

func TestBLPop_ImmediatelyAvailable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "blpop1", "a", "b", "c")

	result, err := testClient.BLPop(ctx, time.Second, "blpop1").Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"blpop1", "a"}, result)
}

func TestBLPop_BlocksUntilPush(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	go func() {
		time.Sleep(100 * time.Millisecond)
		testClient.RPush(ctx, "blpop2", "hello")
	}()

	result, err := testClient.BLPop(ctx, 2*time.Second, "blpop2").Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"blpop2", "hello"}, result)
}

func TestBLPop_Timeout(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.BLPop(ctx, 200*time.Millisecond, "blpop_empty").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestBLPop_MultipleKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "blpop_mk2", "first")

	result, err := testClient.BLPop(ctx, time.Second, "blpop_mk1", "blpop_mk2").Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"blpop_mk2", "first"}, result)
}

func TestLRange_FullList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lrange1", "a", "b", "c")

	vals, err := testClient.LRange(ctx, "lrange1", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, vals)
}

func TestLRange_PositiveIndices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lrange2", "a", "b", "c", "d", "e")

	vals, err := testClient.LRange(ctx, "lrange2", 1, 3).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"b", "c", "d"}, vals)
}

func TestLRange_NegativeIndices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lrange3", "a", "b", "c", "d", "e")

	vals, err := testClient.LRange(ctx, "lrange3", -3, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"c", "d", "e"}, vals)
}

func TestLRange_OutOfBoundsClampedToList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lrange4", "a", "b", "c")

	vals, err := testClient.LRange(ctx, "lrange4", -100, 100).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, vals)
}

func TestLRange_NonExistingKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	vals, err := testClient.LRange(ctx, "lrange_nonexistent", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{}, vals)
}

func TestLRange_StartGreaterThanEnd(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lrange5", "a", "b", "c")

	vals, err := testClient.LRange(ctx, "lrange5", 3, 1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{}, vals)
}

func TestLMove_LeftToLeft(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lmove1_src", "a", "b", "c")

	val, err := testClient.LMove(ctx, "lmove1_src", "lmove1_dst", "LEFT", "LEFT").Result()
	assert.NoError(t, err)
	assert.Equal(t, "a", val)

	// src should now be [b, c]
	srcVals, _ := testClient.LRange(ctx, "lmove1_src", 0, -1).Result()
	assert.Equal(t, []string{"b", "c"}, srcVals)

	// dst should be [a]
	dstVals, _ := testClient.LRange(ctx, "lmove1_dst", 0, -1).Result()
	assert.Equal(t, []string{"a"}, dstVals)
}

func TestLMove_RightToLeft(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lmove2_src", "a", "b", "c")

	val, err := testClient.LMove(ctx, "lmove2_src", "lmove2_dst", "RIGHT", "LEFT").Result()
	assert.NoError(t, err)
	assert.Equal(t, "c", val)

	// src should now be [a, b]
	srcVals, _ := testClient.LRange(ctx, "lmove2_src", 0, -1).Result()
	assert.Equal(t, []string{"a", "b"}, srcVals)

	// dst should be [c]
	dstVals, _ := testClient.LRange(ctx, "lmove2_dst", 0, -1).Result()
	assert.Equal(t, []string{"c"}, dstVals)
}

func TestLMove_LeftToRight(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lmove3_src", "a", "b", "c")
	testClient.RPush(ctx, "lmove3_dst", "x", "y")

	val, err := testClient.LMove(ctx, "lmove3_src", "lmove3_dst", "LEFT", "RIGHT").Result()
	assert.NoError(t, err)
	assert.Equal(t, "a", val)

	// dst should be [x, y, a]
	dstVals, _ := testClient.LRange(ctx, "lmove3_dst", 0, -1).Result()
	assert.Equal(t, []string{"x", "y", "a"}, dstVals)
}

func TestLMove_RightToRight(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lmove4_src", "a", "b", "c")
	testClient.RPush(ctx, "lmove4_dst", "x", "y")

	val, err := testClient.LMove(ctx, "lmove4_src", "lmove4_dst", "RIGHT", "RIGHT").Result()
	assert.NoError(t, err)
	assert.Equal(t, "c", val)

	// dst should be [x, y, c]
	dstVals, _ := testClient.LRange(ctx, "lmove4_dst", 0, -1).Result()
	assert.Equal(t, []string{"x", "y", "c"}, dstVals)
}

func TestLMove_NonExistingSource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := testClient.LMove(ctx, "lmove_nonexistent_src", "lmove5_dst", "LEFT", "LEFT").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestLMove_SameKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lmove6", "a", "b", "c")

	// LMOVE src src LEFT RIGHT rotates the list: [a,b,c] -> [b,c,a]
	val, err := testClient.LMove(ctx, "lmove6", "lmove6", "LEFT", "RIGHT").Result()
	assert.NoError(t, err)
	assert.Equal(t, "a", val)

	vals, _ := testClient.LRange(ctx, "lmove6", 0, -1).Result()
	assert.Equal(t, []string{"b", "c", "a"}, vals)
}

func TestLMove_SingleElementSource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testClient.RPush(ctx, "lmove7_src", "only")

	val, err := testClient.LMove(ctx, "lmove7_src", "lmove7_dst", "LEFT", "LEFT").Result()
	assert.NoError(t, err)
	assert.Equal(t, "only", val)

	// src should now be empty (key gone)
	srcLen, _ := testClient.LLen(ctx, "lmove7_src").Result()
	assert.Equal(t, int64(0), srcLen)

	dstVals, _ := testClient.LRange(ctx, "lmove7_dst", 0, -1).Result()
	assert.Equal(t, []string{"only"}, dstVals)
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
