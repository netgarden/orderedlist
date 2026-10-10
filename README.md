# orderedlist

`github.com/netgarden/orderedlist` — a generic, concurrency-safe sorted
list. Elements are kept in ascending order (per a user-supplied
`Comparator`) as they're added and removed, so index `0` is always the
smallest element — useful as a simple priority queue (e.g. "next scheduled
run time" ordering).

```go
l := orderedlist.New(func(a, b int) int { return a - b })

l.Add(5)
l.Add(1)
l.Add(3)

v, _ := l.Peek() // 1, without removing it
l.RemoveAt(0)    // remove and return the smallest element
l.Remove(3)      // remove a specific value if present
```

## Ties and identity

Elements that compare equal (same `Comparator` result) are not necessarily
the same value — `Add` preserves their relative insertion order (stable,
FIFO among ties), and `Remove` only treats two elements as "the same" if
they're actually equal (`==`), not merely tied under the comparator. This
matters whenever `T` carries more state than the sort key: removing one of
several tied elements removes exactly that one, never a sibling that
happens to sort the same.

## Concurrency

`List` is safe for concurrent use by multiple goroutines.

## License

Apache License 2.0, see [LICENSE](LICENSE).
