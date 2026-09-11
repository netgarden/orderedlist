// Package orderedlist provides a generic, concurrency-safe sorted list:
// elements are kept in ascending order (as defined by a Comparator) as they
// are added and removed, making index 0 always the "smallest" element —
// useful for e.g. a priority queue of scheduled work ordered by run time.
package orderedlist

import (
	"sort"
	"sync"
)

// Comparator returns a negative number if a orders before b, a positive
// number if a orders after b, and 0 if they are equal for ordering
// purposes. Elements that compare equal may still be distinct values;
// List preserves their relative insertion order (stable) and Remove
// distinguishes between them by value equality, not just by comparing equal.
type Comparator[T any] func(a, b T) int

// New creates an empty List ordered by comparator.
func New[T comparable](comparator Comparator[T]) *List[T] {
	return &List[T]{
		data:       make([]T, 0),
		comparator: comparator,
	}
}

// List is a sorted list safe for concurrent use by multiple goroutines.
type List[T comparable] struct {
	mu         sync.Mutex
	data       []T
	comparator Comparator[T]
}

// Add inserts item in sorted order and returns the index it was inserted
// at. When one or more existing elements compare equal to item, item is
// inserted after all of them (stable, FIFO among ties).
func (l *List[T]) Add(item T) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	index := sort.Search(len(l.data), func(i int) bool {
		return l.comparator(l.data[i], item) > 0
	})

	l.data = append(l.data, item)
	copy(l.data[index+1:], l.data[index:])
	l.data[index] = item

	return index
}

// Remove removes the first element equal to item (Go equality, not just
// comparator equality) and reports whether an element was removed. When
// multiple elements compare equal under the Comparator, only the range of
// comparator-ties is scanned to find the exact match, so distinct elements
// that happen to sort together are never confused with one another.
func (l *List[T]) Remove(item T) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	start := sort.Search(len(l.data), func(i int) bool {
		return l.comparator(l.data[i], item) >= 0
	})

	for i := start; i < len(l.data) && l.comparator(l.data[i], item) == 0; i++ {
		if l.data[i] == item {
			l.removeAt(i)
			return true
		}
	}

	return false
}

// RemoveAt removes and returns the element at index i. It panics if i is
// out of range, consistent with slice indexing.
func (l *List[T]) RemoveAt(i int) T {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.removeAt(i)
}

func (l *List[T]) removeAt(i int) T {
	item := l.data[i]
	l.data = append(l.data[:i], l.data[i+1:]...)
	return item
}

// Get returns the element at index i. It panics if i is out of range,
// consistent with slice indexing.
func (l *List[T]) Get(i int) T {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.data[i]
}

// Peek returns the smallest element (index 0) without removing it. It
// returns false if the list is empty.
func (l *List[T]) Peek() (T, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.data) == 0 {
		var zero T
		return zero, false
	}

	return l.data[0], true
}

// Len returns the number of elements currently in the list.
func (l *List[T]) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.data)
}
