package SortedSet_test

import (
	Error "github.com/go-composites/error/src"
	Result "github.com/go-composites/result/src"
	SortedSet "github.com/go-composites/sortedset/src"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// lessInt is the comparator used throughout: ascending integer order.
func lessInt(a, b interface{}) bool { return a.(int) < b.(int) }

// errResult builds a Result that reports HasError() == true, so Each
// short-circuits on it (HasError() is !error.IsNull()).
func errResult() Result.Interface {
	return Result.New(Result.WithError(Error.New("sentinel")))
}

// ints collects a SortedSet's int items into a Go slice in iteration order.
// Because iteration is sorted, no extra sorting is needed — the assertions
// exercise the ordering directly.
func ints(s SortedSet.Interface) []int {
	out := []int{}
	s.Each(func(item interface{}) Result.Interface {
		out = append(out, item.(int))
		return Result.New()
	})
	return out
}

var _ = ginkgo.Describe("SortedSet", func() {
	ginkgo.Describe("New", func() {
		ginkgo.It("returns a non-nil, non-null, empty SortedSet", func() {
			s := SortedSet.New(lessInt)
			gomega.Expect(s).NotTo(gomega.BeNil())
			gomega.Expect(s.IsNull()).To(gomega.BeFalse())
			gomega.Expect(s.Len()).To(gomega.Equal(0))
			gomega.Expect(s.IsEmpty()).To(gomega.BeTrue())
		})

		ginkgo.It("seeds variadic items, dedup and sorted by comparator", func() {
			s := SortedSet.New(lessInt, 3, 1, 3, 2, 1, 2)
			gomega.Expect(s.Len()).To(gomega.Equal(3))
			gomega.Expect(ints(s)).To(gomega.Equal([]int{1, 2, 3}))
			gomega.Expect(s.IsEmpty()).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("Add", func() {
		ginkgo.It("inserts items in sorted position and chains", func() {
			s := SortedSet.New(lessInt)
			ret := s.Add(2).Add(5).Add(1).Add(3)
			gomega.Expect(ret).To(gomega.BeIdenticalTo(s))
			gomega.Expect(s.Len()).To(gomega.Equal(4))
			gomega.Expect(ints(s)).To(gomega.Equal([]int{1, 2, 3, 5}))
		})

		ginkgo.It("is idempotent and keeps the set sorted", func() {
			s := SortedSet.New(lessInt, 1, 2, 3).Add(2).Add(1)
			gomega.Expect(s.Len()).To(gomega.Equal(3))
			gomega.Expect(ints(s)).To(gomega.Equal([]int{1, 2, 3}))
		})

		ginkgo.It("inserts a new minimum and a new maximum in place", func() {
			s := SortedSet.New(lessInt, 5).Add(9).Add(1)
			gomega.Expect(ints(s)).To(gomega.Equal([]int{1, 5, 9}))
		})
	})

	ginkgo.Describe("Delete", func() {
		ginkgo.It("removes an item, keeping the rest sorted", func() {
			s := SortedSet.New(lessInt, 1, 2, 3, 4)
			ret := s.Delete(2)
			gomega.Expect(ret).To(gomega.BeIdenticalTo(s))
			gomega.Expect(s.Has(2)).To(gomega.BeFalse())
			gomega.Expect(s.Len()).To(gomega.Equal(3))
			gomega.Expect(ints(s)).To(gomega.Equal([]int{1, 3, 4}))
		})

		ginkgo.It("is a no-op for an absent item", func() {
			s := SortedSet.New(lessInt, 1)
			s.Delete(99)
			gomega.Expect(s.Len()).To(gomega.Equal(1))
			gomega.Expect(ints(s)).To(gomega.Equal([]int{1}))
		})
	})

	ginkgo.Describe("Has", func() {
		ginkgo.It("is true for a present item", func() {
			s := SortedSet.New(lessInt, 1)
			gomega.Expect(s.Has(1)).To(gomega.BeTrue())
		})

		ginkgo.It("is false for an absent item", func() {
			s := SortedSet.New(lessInt)
			gomega.Expect(s.Has(1)).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("Each", func() {
		ginkgo.It("visits every item in sorted order, clean Result", func() {
			s := SortedSet.New(lessInt, 3, 1, 2)
			out := []int{}
			res := s.Each(func(item interface{}) Result.Interface {
				out = append(out, item.(int))
				return Result.New()
			})
			gomega.Expect(out).To(gomega.Equal([]int{1, 2, 3}))
			gomega.Expect(res).NotTo(gomega.BeNil())
			gomega.Expect(res.HasError()).To(gomega.BeFalse())
		})

		ginkgo.It("short-circuits on the first error Result", func() {
			s := SortedSet.New(lessInt, 1, 2, 3)
			count := 0
			res := s.Each(func(item interface{}) Result.Interface {
				count++
				return errResult()
			})
			gomega.Expect(count).To(gomega.Equal(1))
			gomega.Expect(res.HasError()).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("ToArray", func() {
		ginkgo.It("materialises the items into an Array, sorted", func() {
			s := SortedSet.New(lessInt, 3, 1, 2)
			arr := s.ToArray()
			gomega.Expect(arr).NotTo(gomega.BeNil())
			gomega.Expect(arr.Len()).To(gomega.Equal(3))

			out := []int{}
			arr.Each(func(_ int, item interface{}) Result.Interface {
				out = append(out, item.(int))
				return Result.New()
			})
			gomega.Expect(out).To(gomega.Equal([]int{1, 2, 3}))
		})
	})

	ginkgo.Describe("First", func() {
		ginkgo.It("carries the minimum item as payload", func() {
			res := SortedSet.New(lessInt, 3, 1, 2).First()
			gomega.Expect(res.HasError()).To(gomega.BeFalse())
			gomega.Expect(res.Payload()).To(gomega.Equal(1))
		})

		ginkgo.It("is an error Result for an empty set", func() {
			res := SortedSet.New(lessInt).First()
			gomega.Expect(res.HasError()).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("Last", func() {
		ginkgo.It("carries the maximum item as payload", func() {
			res := SortedSet.New(lessInt, 3, 1, 2).Last()
			gomega.Expect(res.HasError()).To(gomega.BeFalse())
			gomega.Expect(res.Payload()).To(gomega.Equal(3))
		})

		ginkgo.It("is an error Result for an empty set", func() {
			res := SortedSet.New(lessInt).Last()
			gomega.Expect(res.HasError()).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("Union", func() {
		ginkgo.It("keeps the union sorted, both directions", func() {
			a := SortedSet.New(lessInt, 1, 2, 3)
			b := SortedSet.New(lessInt, 3, 4, 5)
			gomega.Expect(ints(a.Union(b))).To(
				gomega.Equal([]int{1, 2, 3, 4, 5}))
			gomega.Expect(ints(b.Union(a))).To(
				gomega.Equal([]int{1, 2, 3, 4, 5}))
		})
	})

	ginkgo.Describe("Intersection", func() {
		ginkgo.It("keeps the common items sorted, both directions", func() {
			a := SortedSet.New(lessInt, 1, 2, 3, 4)
			b := SortedSet.New(lessInt, 4, 2)
			gomega.Expect(ints(a.Intersection(b))).To(
				gomega.Equal([]int{2, 4}))
			gomega.Expect(ints(b.Intersection(a))).To(
				gomega.Equal([]int{2, 4}))
		})
	})

	ginkgo.Describe("Difference", func() {
		ginkgo.It("keeps receiver items not in other, sorted", func() {
			a := SortedSet.New(lessInt, 1, 2, 3)
			b := SortedSet.New(lessInt, 3, 4, 5)
			gomega.Expect(ints(a.Difference(b))).To(
				gomega.Equal([]int{1, 2}))
			gomega.Expect(ints(b.Difference(a))).To(
				gomega.Equal([]int{4, 5}))
		})
	})

	ginkgo.Describe("IsSubset", func() {
		ginkgo.It("is true when every item is in the other set", func() {
			gomega.Expect(SortedSet.New(lessInt, 1, 2).IsSubset(
				SortedSet.New(lessInt, 1, 2, 3))).To(gomega.BeTrue())
		})

		ginkgo.It("is false when an item is missing from the other set", func() {
			gomega.Expect(SortedSet.New(lessInt, 1, 9).IsSubset(
				SortedSet.New(lessInt, 1, 2, 3))).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("Equal", func() {
		ginkgo.It("is order-INsensitive: same members", func() {
			gomega.Expect(SortedSet.New(lessInt, 1, 2, 3).Equal(
				SortedSet.New(lessInt, 3, 2, 1))).To(gomega.BeTrue())
		})

		ginkgo.It("is false for sets of different sizes", func() {
			gomega.Expect(SortedSet.New(lessInt, 1, 2).Equal(
				SortedSet.New(lessInt, 1, 2, 3))).To(gomega.BeFalse())
		})

		ginkgo.It("is false for same-size sets with different items", func() {
			gomega.Expect(SortedSet.New(lessInt, 1, 2, 3).Equal(
				SortedSet.New(lessInt, 1, 2, 9))).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("Null", func() {
		ginkgo.It("is a Null-Object: IsNull true and inert", func() {
			n := SortedSet.Null()
			gomega.Expect(n).NotTo(gomega.BeNil())
			gomega.Expect(n.IsNull()).To(gomega.BeTrue())
			gomega.Expect(n.Len()).To(gomega.Equal(0))
			gomega.Expect(n.IsEmpty()).To(gomega.BeTrue())

			// Mutators are no-ops that return the receiver.
			gomega.Expect(n.Add(1)).To(gomega.BeIdenticalTo(n))
			gomega.Expect(n.Delete(1)).To(gomega.BeIdenticalTo(n))
			gomega.Expect(n.Len()).To(gomega.Equal(0))

			// Membership always misses.
			gomega.Expect(n.Has(1)).To(gomega.BeFalse())

			// ToArray is empty.
			gomega.Expect(n.ToArray().Len()).To(gomega.Equal(0))

			// First/Last report an error Result.
			gomega.Expect(n.First().HasError()).To(gomega.BeTrue())
			gomega.Expect(n.Last().HasError()).To(gomega.BeTrue())

			// Set algebra returns the (inert) null set.
			gomega.Expect(n.Union(SortedSet.New(lessInt, 1)).IsNull()).To(
				gomega.BeTrue())
			gomega.Expect(n.Intersection(SortedSet.New(lessInt, 1)).IsNull()).To(
				gomega.BeTrue())
			gomega.Expect(n.Difference(SortedSet.New(lessInt, 1)).IsNull()).To(
				gomega.BeTrue())

			// IsSubset of anything is true; Equal holds only for empty sets.
			gomega.Expect(n.IsSubset(SortedSet.New(lessInt, 1))).To(
				gomega.BeTrue())
			gomega.Expect(n.Equal(SortedSet.New(lessInt))).To(gomega.BeTrue())
			gomega.Expect(n.Equal(SortedSet.New(lessInt, 1))).To(
				gomega.BeFalse())

			// Each returns a clean Result without invoking fn.
			called := false
			res := n.Each(func(item interface{}) Result.Interface {
				called = true
				return errResult()
			})
			gomega.Expect(called).To(gomega.BeFalse())
			gomega.Expect(res.HasError()).To(gomega.BeFalse())
		})
	})
})
