<p align="center"><img src="https://raw.githubusercontent.com/go-composites/brand/main/social/go-composites.png" alt="go-composites/sortedset" width="720"></p>

# sortedset

[![ci](https://github.com/go-composites/sortedset/actions/workflows/ci.yml/badge.svg)](https://github.com/go-composites/sortedset/actions/workflows/ci.yml)

A **comparator-sorted** Set composite for Composition-Oriented Programming — a
collection of unique items kept permanently sorted by a caller-supplied
comparator (TreeSet-like). It is the comparator-ordered sibling of
[`set`](https://github.com/go-composites/set) (unordered) and
[`orderedset`](https://github.com/go-composites/orderedset) (first-insertion
order), and shares the `Each`/`ToArray` grammar with
[`array`](https://github.com/go-composites/array).

A `SortedSet` is backed by **both** a `[]interface{}` (kept in ascending
comparator order) and a `map[interface{}]struct{}` (giving O(1) membership and
dedup), kept in lock-step. The result: iteration, `ToArray` and the set-algebra
results are all emitted in **ascending comparator order** rather than Go's
unspecified map order — so consumers (and tests) never depend on map-iteration
flakiness. Because the set is always sorted, `First()` and `Last()` (the minimum
and maximum by the comparator) are natural, O(1) operations.

It follows the go-composites grammar:

- **Comparator-sorted iteration**: `New(less, items...)` takes a `less`
  comparator that defines the order; `Add` inserts each new item in its sorted
  position (found with `sort.Search`); `Delete` removes while keeping the rest
  sorted; `Each`/`ToArray` iterate in ascending order; `Union`, `Intersection`
  and `Difference` all yield sorted results (using the **receiver's**
  comparator).
- **`First` / `Last`**: return a [`Result`](https://github.com/go-composites/result)
  carrying the minimum / maximum item — or an error `Result` for an empty set.
- **Never nil / Null-Object**: every constructor and method returns a real
  object; `Null()` provides an inert variant and `IsNull()` distinguishes it.
- **Result-based errors**: fallible iteration returns a `Result` — `Each`
  short-circuits on the first `Result` whose `HasError()` is true. No panics, no
  bare nils.
- **Composite returns**: `ToArray()` materialises into an
  [`Array`](https://github.com/go-composites/array); set algebra returns fresh
  `SortedSet`s.

Items must be **comparable** (they back a Go map for membership) **AND
orderable** by the `less` comparator passed to `New`.

**`Equal` is order-INsensitive**: two SortedSets are equal when they hold the
same members regardless of their comparators, exactly like a mathematical set.

## Install

```sh
go get github.com/go-composites/sortedset@main
```

## Usage

```go
package main

import (
	"fmt"

	Result "github.com/go-composites/result/src"
	SortedSet "github.com/go-composites/sortedset/src"
)

func main() {
	// less defines ascending integer order; the set stays sorted by it.
	less := func(a, b interface{}) bool { return a.(int) < b.(int) }

	a := SortedSet.New(less, 3, 1, 2)
	a.Add(5).Add(4) // out of order — inserted in sorted position; chainable.

	fmt.Println(a.Len())     // 5
	fmt.Println(a.Has(2))    // true
	fmt.Println(a.IsEmpty()) // false

	// First/Last are the min/max by the comparator.
	fmt.Println(a.First().Payload()) // 1
	fmt.Println(a.Last().Payload())  // 5

	b := SortedSet.New(less, 3, 4, 5)
	_ = a.Union(b)        // {1,2,3,4,5} — sorted, receiver's comparator
	_ = a.Intersection(b) // {3,4,5}     — sorted
	_ = a.Difference(b)   // {1,2}       — sorted

	fmt.Println(SortedSet.New(less, 1, 2).IsSubset(a))    // true
	fmt.Println(a.Equal(SortedSet.New(less, 5, 4, 3, 2, 1))) // true (order-insensitive)

	// Each iterates in sorted order and short-circuits on the first
	// Result whose HasError() is true.
	a.Each(func(item interface{}) Result.Interface {
		fmt.Println(item)
		return Result.New()
	})

	// ToArray materialises into a go-composites Array, in sorted order.
	_ = a.ToArray()

	a.Delete(1) // removes 1, keeping {2,3,4,5} sorted
}
```

### API

| Method | Returns | Notes |
| --- | --- | --- |
| `New(less, items...)` | `SortedSet.Interface` | `less` comparator defines the order; variadic items deduplicated, inserted in sorted position |
| `Null()` | `SortedSet.Interface` | inert Null-Object; `IsNull()` is `true` |
| `Add(item)` | `SortedSet.Interface` | inserts in sorted position only if new (idempotent); chainable |
| `Delete(item)` | `SortedSet.Interface` | no-op when absent; keeps the rest sorted; chainable |
| `Has(item)` | `bool` | membership test (O(1)) |
| `Len()` | `int` | number of items |
| `IsEmpty()` | `bool` | `true` when there are no items |
| `Each(fn)` | `Result.Interface` | iterate in sorted order; short-circuit on `HasError()` |
| `ToArray()` | `Array.Interface` | materialise into an Array, in sorted order |
| `First()` | `Result.Interface` | minimum by the comparator; error `Result` when empty |
| `Last()` | `Result.Interface` | maximum by the comparator; error `Result` when empty |
| `Union(other)` | `SortedSet.Interface` | Ruby `\|` — sorted, receiver's comparator |
| `Intersection(other)` | `SortedSet.Interface` | Ruby `&` — common items, sorted |
| `Difference(other)` | `SortedSet.Interface` | Ruby `-` — receiver items not in other, sorted |
| `IsSubset(other)` | `bool` | every item is also in `other` |
| `Equal(other)` | `bool` | same members, **order-insensitive** |
| `IsNull()` | `bool` | `false` for a real SortedSet |

## License

BSD-3-Clause — see [LICENSE](LICENSE).
