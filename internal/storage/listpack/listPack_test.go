package listpack

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListPack_NewEmptyListPack(t *testing.T) {
	lp := NewEmptyListPack(256)
	assert.Equal(t, 0, lp.Length())
}

func TestListPack_NewListPack(t *testing.T) {
	elements := [][]byte{[]byte("hello"), []byte("world")}
	lp := NewListPack(256, elements...)
	assert.Equal(t, 2, lp.Length())
}

func TestListPack_PushElements(t *testing.T) {
	initialElements := [][]byte{[]byte("hello"), []byte("world")}
	lp := NewListPack(256, initialElements...)

	_, _ = lp.Push([]byte("avacado"))
	size, err := lp.Push([]byte("ListPack"))
	assert.NoError(t, err)

	assert.Equal(t, 4, size)
	assert.Equal(t, 4, lp.Length())
}

func TestListsMemoryStore_PopElements(t *testing.T) {
	initialElements := [][]byte{[]byte("hello"), []byte("world")}
	lp := NewListPack(256, initialElements...)
	_, _ = lp.Push([]byte("avacado"))
	_, _ = lp.Push([]byte("ListPack"))

	assert.Equal(t, []byte("ListPack"), lp.Pop())
	assert.Equal(t, []byte("avacado"), lp.Pop())
	assert.Equal(t, []byte("world"), lp.Pop())
	assert.Equal(t, []byte("hello"), lp.Pop())
}

func TestListsMemoryStore_LPush(t *testing.T) {
	lp := NewEmptyListPack(1024)
	_, _ = lp.Push([]byte("world"))
	_, _ = lp.Push([]byte("-124"))
	_, _ = lp.LPush([]byte("hello"))
	_, _ = lp.LPush([]byte("1231313"))

	assert.Equal(t, 4, lp.Length())
	assert.Equal(t, []byte("-124"), lp.Pop())
	assert.Equal(t, []byte("world"), lp.Pop())
	assert.Equal(t, []byte("hello"), lp.Pop())
	assert.Equal(t, []byte("1231313"), lp.Pop())
}

func TestListPack_PushOverflow(t *testing.T) {
	t.Run("single value too large for empty listpack", func(t *testing.T) {
		// maxSize=10: freeBytes = 10-7 = 3; "hello" needs 7 bytes
		lp := NewEmptyListPack(10)
		count, err := lp.Push([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 0, count)       // original elemCount returned
		assert.Equal(t, 0, lp.Length()) // listpack not modified
	})

	t.Run("second value overflows, first value committed", func(t *testing.T) {
		// maxSize=19: freeBytes=12; "hello"(7) fits → totalSize=14, freeBytes=5; second "hello"(7) doesn't
		lp := NewEmptyListPack(19)
		count, err := lp.Push([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		count, err = lp.Push([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 1, count)       // original elemCount before failed Push
		assert.Equal(t, 1, lp.Length()) // first value still intact
	})

	t.Run("listpack not corrupted after overflow error", func(t *testing.T) {
		lp := NewEmptyListPack(19)
		_, _ = lp.Push([]byte("hello"))
		_, _ = lp.Push([]byte("hello")) // overflows, ignored

		assert.Equal(t, []byte("hello"), lp.Pop())
	})
}

func TestListPack_LPushOverflow(t *testing.T) {
	t.Run("single value too large for empty listpack", func(t *testing.T) {
		// maxSize=10: freeBytes=3; "hello" needs 7 bytes
		lp := NewEmptyListPack(10)
		count, err := lp.LPush([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, 0, lp.Length())
	})

	t.Run("second value overflows, first value committed", func(t *testing.T) {
		lp := NewEmptyListPack(19)
		count, err := lp.LPush([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		count, err = lp.LPush([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, lp.Length())
	})
}

func TestListsMemoryStore_LPop(t *testing.T) {
	elements := [][]byte{[]byte("hello"), []byte("world"), []byte("124"), []byte("JamesBond")}
	lp := NewListPack(1024, elements...)

	popped := lp.LPop(2)
	assert.Equal(t, 2, len(popped))
	assert.Equal(t, elements[0], popped[0])
	assert.Equal(t, elements[1], popped[1])

	popped = lp.LPop(3)
	//assert.Equal(t, 2, len(popped))
	//assert.Equal(t, string(elements[2]), string(popped[0]))
	//assert.Equal(t, elements[3], popped[1])
	//
	//popped = lp.LPop(4)
	//assert.Equal(t, 0, len(popped))
}

func TestListsMemoryStore_IsEmpty(t *testing.T) {
	lp := NewEmptyListPack(24)
	assert.True(t, lp.IsEmpty())

	_, _ = lp.Push([]byte("12"))
	assert.False(t, lp.IsEmpty())

	lp.Pop()
	assert.True(t, lp.IsEmpty())
}

func TestListMemoryStore_LIndex(t *testing.T) {
	lp := NewEmptyListPack(24)
	_, _ = lp.Push([]byte("Hi"))
	_, _ = lp.Push([]byte("120"))
	_, _ = lp.Push([]byte("300"))
	_, _ = lp.Push([]byte("hi bye"))

	t.Run("Positive index only", func(t *testing.T) {
		element, found := lp.AtIndex(3)
		assert.True(t, found)
		assert.Equal(t, []byte("hi bye"), element)

		element, found = lp.AtIndex(1)
		assert.True(t, found)
		assert.Equal(t, []byte("120"), element)

		element, found = lp.AtIndex(0)
		assert.True(t, found)
		assert.Equal(t, []byte("Hi"), element)

		element, found = lp.AtIndex(4)
		assert.False(t, found)
		assert.Nil(t, element)
	})

	t.Run("Negative index", func(t *testing.T) {
		element, found := lp.AtIndex(-1)
		assert.True(t, found)
		assert.Equal(t, []byte("hi bye"), element)

		element, found = lp.AtIndex(-2)
		assert.True(t, found)
		assert.Equal(t, []byte("300"), element)

		element, found = lp.AtIndex(-3)
		assert.True(t, found)
		assert.Equal(t, []byte("120"), element)

		element, found = lp.AtIndex(-4)
		assert.True(t, found)
		assert.Equal(t, []byte("Hi"), element)

		element, found = lp.AtIndex(-5)
		assert.False(t, found)
		assert.Nil(t, element)
	})
}

func TestListPack_LRange(t *testing.T) {
	t.Run("from zero index to last", func(t *testing.T) {
		lp := NewListPack(60, []byte("Hi"), []byte("hello World"), []byte("12"))
		elements, err := lp.LRange(0, 3)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(elements))
		assert.Equal(t, []byte("Hi"), elements[0])
		assert.Equal(t, []byte("hello World"), elements[1])
		assert.Equal(t, []byte("12"), elements[2])
	})

	t.Run("from non zero index to last", func(t *testing.T) {
		lp := NewListPack(60, []byte("Hi"), []byte("hello World"), []byte("12"), []byte("43"))
		elements, err := lp.LRange(1, 3)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(elements))
		assert.Equal(t, []byte("hello World"), elements[0])
		assert.Equal(t, []byte("12"), elements[1])
		assert.Equal(t, []byte("43"), elements[2])
	})

	t.Run("from non zero index to non last", func(t *testing.T) {
		lp := NewListPack(60, []byte("Hi"), []byte("hello World"), []byte("12"), []byte("43"))
		elements, err := lp.LRange(1, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(elements))
		assert.Equal(t, []byte("hello World"), elements[0])
		assert.Equal(t, []byte("12"), elements[1])
	})
}

func TestListPack_InsertAt(t *testing.T) {
	t.Run("Insert at the start", func(t *testing.T) {
		lp := NewListPack(1024, []byte("First Value"))
		err := lp.InsertAt(0, []byte("Zero Value"))
		assert.NoError(t, err)

		v, _ := lp.AtIndex(0)
		assert.Equal(t, "Zero Value", string(v))
		v, _ = lp.AtIndex(1)
		assert.Equal(t, "First Value", string(v))
	})
	t.Run("Insert at the end", func(t *testing.T) {
		lp := NewListPack(1024, []byte("First Value"))
		err := lp.InsertAt(1, []byte("Second Value"))
		assert.NoError(t, err)

		v, _ := lp.AtIndex(0)
		assert.Equal(t, "First Value", string(v))
		v, _ = lp.AtIndex(1)
		assert.Equal(t, "Second Value", string(v))
	})
	t.Run("Insert in middle", func(t *testing.T) {
		lp := NewListPack(1024, []byte("First Value"), []byte("Second Value"), []byte("Fourth Value"))

		err := lp.InsertAt(2, []byte("Third Value"))
		assert.NoError(t, err)
		v, _ := lp.AtIndex(0)
		assert.Equal(t, "First Value", string(v))
		v, _ = lp.AtIndex(1)
		assert.Equal(t, "Second Value", string(v))
		v, _ = lp.AtIndex(2)
		assert.Equal(t, "Third Value", string(v))
		v, _ = lp.AtIndex(3)
		assert.Equal(t, "Fourth Value", string(v))
	})
	t.Run("Insert beyond length", func(t *testing.T) {
		lp := NewListPack(1024, []byte("First Value"))

		err := lp.InsertAt(2, []byte("Third Value"))
		assert.NoError(t, err)

		v, _ := lp.AtIndex(1)
		assert.Equal(t, "Third Value", string(v))
	})
	t.Run("Insert in empty listpack", func(t *testing.T) {
		lp := NewEmptyListPack(1024)

		err := lp.InsertAt(0, []byte("Hello World"))
		assert.NoError(t, err)

		v, _ := lp.AtIndex(0)
		assert.Equal(t, "Hello World", string(v))
	})
}

func TestListPack_ReplaceAt(t *testing.T) {
	t.Run("Replace equal size element", func(t *testing.T) {
		lp := NewListPack(100, []byte("first"), []byte("second"), []byte("third"))

		err := lp.ReplaceAt(1, []byte("SECOND"))
		assert.NoError(t, err)
		assertContainsExactly(t, []string{"first", "SECOND", "third"}, lp)

		err = lp.ReplaceAt(0, []byte("FIRST"))
		assert.NoError(t, err)
		assertContainsExactly(t, []string{"FIRST", "SECOND", "third"}, lp)

		err = lp.ReplaceAt(2, []byte("THIRD"))
		assert.NoError(t, err)
		assertContainsExactly(t, []string{"FIRST", "SECOND", "THIRD"}, lp)
	})
	t.Run("Replace with small size element", func(t *testing.T) {
		lp := NewListPack(100, []byte("first"), []byte("second"), []byte("third"))

		err := lp.ReplaceAt(1, []byte("SEC"))
		assert.NoError(t, err)
		assertContainsExactly(t, []string{"first", "SEC", "third"}, lp)

		err = lp.ReplaceAt(0, []byte("FIR"))
		assert.NoError(t, err)
		assertContainsExactly(t, []string{"FIR", "SEC", "third"}, lp)

		err = lp.ReplaceAt(2, []byte("THIR"))
		assert.NoError(t, err)
		assertContainsExactly(t, []string{"FIR", "SEC", "THIR"}, lp)
	})
	t.Run("Replace with large size element", func(t *testing.T) {
		lp := NewListPack(100, []byte("fir"), []byte("SEC"), []byte("thi"))

		err := lp.ReplaceAt(1, []byte("second"))
		assert.NoError(t, err)
		assertContainsExactly(t, []string{"fir", "second", "thi"}, lp)

		err = lp.ReplaceAt(0, []byte("first"))
		assert.NoError(t, err)
		assertContainsExactly(t, []string{"first", "second", "thi"}, lp)

		err = lp.ReplaceAt(2, []byte("third"))
		assert.NoError(t, err)
		assertContainsExactly(t, []string{"first", "second", "third"}, lp)
	})
}

func TestListPack_IndexOf(t *testing.T) {
	t.Run("find index of existing element without skipping odds", func(t *testing.T) {
		lp := NewListPack(200, []byte("V1"), []byte("V2"), []byte("V2"), []byte("V3"))
		index, found := lp.IndexOf("V2", false)
		assert.True(t, found)
		assert.Equal(t, 1, index)
	})
	t.Run("find index of existing element with skipping odds", func(t *testing.T) {
		lp := NewListPack(200, []byte("V1"), []byte("V2"), []byte("V2"), []byte("V3"))
		index, found := lp.IndexOf("V2", true)
		assert.True(t, found)
		assert.Equal(t, 2, index)
	})
	t.Run("find index of non existing element without skipping odds", func(t *testing.T) {
		lp := NewListPack(200, []byte("V1"), []byte("V2"), []byte("V2"), []byte("V3"))
		index, found := lp.IndexOf("V4", false)
		assert.False(t, found)
		assert.Equal(t, -1, index)
	})
	t.Run("find index of non existing element with skipping odds", func(t *testing.T) {
		lp := NewListPack(200, []byte("V1"), []byte("V2"), []byte("V2"), []byte("V3"))
		index, found := lp.IndexOf("V3", true)
		assert.False(t, found)
		assert.Equal(t, -1, index)
	})
}

func TestListPack_PushAllOrNone(t *testing.T) {
	t.Run("all entries fit, all are pushed", func(t *testing.T) {
		lp := NewEmptyListPack(21)
		count, err := lp.PushAllOrNone([]byte("hello"), []byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Equal(t, 2, lp.Length())
		v, _ := lp.AtIndex(0)
		assert.Equal(t, "hello", string(v))
		v, _ = lp.AtIndex(1)
		assert.Equal(t, "hello", string(v))
	})

	t.Run("not all entries fit, none are pushed (all-or-none)", func(t *testing.T) {
		lp := NewEmptyListPack(21)
		count, err := lp.PushAllOrNone([]byte("hello"), []byte("hello"), []byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, -1, count)
		assert.Equal(t, 0, lp.Length()) // nothing pushed
	})

	t.Run("single entry too large for empty listpack", func(t *testing.T) {
		lp := NewEmptyListPack(10)
		count, err := lp.PushAllOrNone([]byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, -1, count)
		assert.Equal(t, 0, lp.Length())
	})

	t.Run("listpack unchanged after failed push", func(t *testing.T) {
		lp := NewEmptyListPack(21)
		_, _ = lp.Push([]byte("hello"))

		count, err := lp.PushAllOrNone([]byte("hello"), []byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, -1, count)
		assert.Equal(t, 1, lp.Length())
		v, _ := lp.AtIndex(0)
		assert.Equal(t, "hello", string(v))
	})

	t.Run("push into non-empty listpack succeeds when space is available", func(t *testing.T) {
		lp := NewEmptyListPack(23)
		_, _ = lp.Push([]byte("hi"))
		_, _ = lp.Push([]byte("hi"))

		count, err := lp.PushAllOrNone([]byte("hi"), []byte("hi"))
		assert.NoError(t, err)
		assert.Equal(t, 4, count)
		assert.Equal(t, 4, lp.Length())
	})

	t.Run("empty entries slice pushes nothing and succeeds", func(t *testing.T) {
		lp := NewEmptyListPack(64)
		count, err := lp.PushAllOrNone()
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, 0, lp.Length())
	})
}

func TestListPack_DeleteFromIndex(t *testing.T) {
	//lp := NewListPack(1024, []byte("Hello"), []byte("World"), []byte("First"), []byte("Second"))
	//lp.DeleteFromIndex(1, 1)
	//assertContainsExactly(t, []string{"Hello", "First", "Second"}, lp)
}

func assertContainsExactly(t *testing.T, expected []string, lp *ListPack) {
	assert.Equal(t, len(expected), lp.Length(), "Unequal length")
	for i, expectedElem := range expected {
		actual, _ := lp.AtIndex(i)
		assert.Equal(t, expectedElem, string(actual), fmt.Sprintf("Failed at index %d", i))
	}
}
