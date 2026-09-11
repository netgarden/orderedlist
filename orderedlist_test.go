package orderedlist

import (
	"sort"
	"sync"
	"testing"
)

func intComparator(a, b int) int {
	return a - b
}

func TestAdd_MaintainsSortOrder(t *testing.T) {
	l := New(intComparator)

	values := []int{5, 1, 9, 3, 7, 2, 8, 4, 6, 0}
	for _, v := range values {
		l.Add(v)
	}

	got := drain(l)
	want := append([]int(nil), values...)
	sort.Ints(want)

	assertEqual(t, got, want)
}

func TestAdd_ManyDuplicates(t *testing.T) {
	// Regression test: the original gate/utils/orderedList implementation
	// failed to stop at the first valid insertion index, so it ended up
	// inserting at the position of the *last* larger element instead of
	// the first — this only surfaces reliably with 3+ distinct values.
	l := New(intComparator)

	for _, v := range []int{1, 9, 5, 3, 7} {
		l.Add(v)
	}

	assertEqual(t, drain(l), []int{1, 3, 5, 7, 9})
}

func TestAdd_StableAmongTies(t *testing.T) {
	type item struct {
		key int
		seq int
	}
	cmp := func(a, b item) int { return a.key - b.key }
	l := New(cmp)

	l.Add(item{key: 1, seq: 0})
	l.Add(item{key: 1, seq: 1})
	l.Add(item{key: 1, seq: 2})

	if l.Get(0).seq != 0 || l.Get(1).seq != 1 || l.Get(2).seq != 2 {
		t.Fatalf("expected insertion order preserved among ties, got seq order %d,%d,%d",
			l.Get(0).seq, l.Get(1).seq, l.Get(2).seq)
	}
}

func TestRemove_DistinguishesTiedValues(t *testing.T) {
	// Regression test: the original implementation matched by comparator
	// equality alone, so removing one of two elements with an equal sort
	// key could remove the wrong one.
	type item struct {
		key int
		id  string
	}
	cmp := func(a, b item) int { return a.key - b.key }
	l := New(cmp)

	a := item{key: 5, id: "a"}
	b := item{key: 5, id: "b"}
	c := item{key: 5, id: "c"}
	l.Add(a)
	l.Add(b)
	l.Add(c)

	if !l.Remove(b) {
		t.Fatal("expected Remove(b) to report success")
	}

	remaining := drain(l)
	if len(remaining) != 2 {
		t.Fatalf("expected 2 elements remaining, got %d: %v", len(remaining), remaining)
	}
	for _, it := range remaining {
		if it.id == "b" {
			t.Fatal("Remove(b) removed the wrong tied element")
		}
	}
}

func TestRemove_NotPresent(t *testing.T) {
	l := New(intComparator)
	l.Add(1)
	l.Add(2)

	if l.Remove(99) {
		t.Fatal("expected Remove of an absent value to report false")
	}
	if l.Len() != 2 {
		t.Fatalf("expected list untouched, got len %d", l.Len())
	}
}

func TestRemoveAt(t *testing.T) {
	l := New(intComparator)
	for _, v := range []int{1, 2, 3} {
		l.Add(v)
	}

	got := l.RemoveAt(1)
	if got != 2 {
		t.Fatalf("expected RemoveAt(1) to return 2, got %d", got)
	}
	assertEqual(t, drain(l), []int{1, 3})
}

func TestPeek(t *testing.T) {
	l := New(intComparator)

	if _, ok := l.Peek(); ok {
		t.Fatal("expected Peek on empty list to report false")
	}

	l.Add(5)
	l.Add(1)

	v, ok := l.Peek()
	if !ok || v != 1 {
		t.Fatalf("expected Peek to return (1, true), got (%d, %v)", v, ok)
	}
	if l.Len() != 2 {
		t.Fatal("Peek must not remove the element")
	}
}

func TestConcurrentAddRemove(t *testing.T) {
	l := New(intComparator)

	var wg sync.WaitGroup
	const n = 200

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			l.Add(v)
		}(i)
	}
	wg.Wait()

	if l.Len() != n {
		t.Fatalf("expected %d elements after concurrent adds, got %d", n, l.Len())
	}

	got := drain(l)
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Fatalf("list not sorted after concurrent adds: %v", got)
		}
	}
}

func drain[T comparable](l *List[T]) []T {
	out := make([]T, 0, l.Len())
	for l.Len() > 0 {
		out = append(out, l.RemoveAt(0))
	}
	return out
}

func assertEqual(t *testing.T, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
