package main

import (
	"fmt"

	Result "github.com/go-composites/result/src"
	SortedSet "github.com/go-composites/sortedset/src"
)

func main() {
	// less defines ascending integer order; items are kept sorted by it.
	less := func(a, b interface{}) bool { return a.(int) < b.(int) }

	a := SortedSet.New(less, 3, 1, 2)
	a.Add(5).Add(4).Add(3) // out of order; 3 is idempotent; stays sorted.

	fmt.Printf("Len = %d\n", a.Len())
	fmt.Printf("Has(2) = %t\n", a.Has(2))
	fmt.Printf("IsEmpty = %t\n", a.IsEmpty())
	fmt.Printf("sorted = %v\n", sorted(a)) // [1 2 3 4 5] — by comparator.

	// First/Last are the min/max by the comparator — natural for a sorted set.
	fmt.Printf("First = %v\n", a.First().Payload())
	fmt.Printf("Last  = %v\n", a.Last().Payload())

	b := SortedSet.New(less, 3, 4, 5)
	fmt.Printf("Union        = %v\n", sorted(a.Union(b)))
	fmt.Printf("Intersection = %v\n", sorted(a.Intersection(b)))
	fmt.Printf("Difference   = %v\n", sorted(a.Difference(b)))
	fmt.Printf("IsSubset     = %t\n", SortedSet.New(less, 1, 2).IsSubset(a))
	fmt.Printf("Equal        = %t\n", a.Equal(SortedSet.New(less, 5, 4, 3, 2, 1)))

	a.Delete(1)
	fmt.Printf("after Delete(1), sorted = %v\n", sorted(a))
}

// sorted collects a SortedSet's int items in iteration order — which, by
// construction, is ascending comparator order, so no extra sorting is needed.
func sorted(s SortedSet.Interface) []int {
	out := []int{}
	s.Each(func(item interface{}) Result.Interface {
		out = append(out, item.(int))
		return Result.New()
	})
	return out
}
