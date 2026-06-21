package SortedSet

import (
	"sort"

	Array "github.com/go-composites/array/src"
	Error "github.com/go-composites/error/src"
	Result "github.com/go-composites/result/src"
)

// Interface is the public contract of a SortedSet composite — a collection of
// unique items kept permanently sorted by a caller-supplied comparator
// (TreeSet-like). It is the comparator-ordered sibling of Set (unordered) and
// OrderedSet (first-insertion order), and shares the Each/ToArray grammar with
// Array: iteration, ToArray and the set-algebra results are all emitted in
// ascending comparator order rather than Go's unspecified map order.
//
// Items must satisfy TWO constraints, which the caller must guarantee:
//   - comparable, because the SortedSet is backed by a Go map for O(1)
//     membership and dedup (exactly like a Dictionary key); and
//   - orderable by the less comparator passed to New, which defines a strict
//     weak ordering used to keep items in sorted position.
//
// Membership tests return a plain bool, fallible iteration returns a Result,
// First/Last return a Result (the minimum/maximum by the comparator, or an
// error Result when the set is empty), and every method honours the Null-Object
// invariant (never nil).
type Interface interface {
	Add(item interface{}) Interface
	Delete(item interface{}) Interface
	Has(item interface{}) bool
	Len() int
	IsEmpty() bool
	Each(fn func(item interface{}) Result.Interface) Result.Interface
	ToArray() Array.Interface
	First() Result.Interface
	Last() Result.Interface
	Union(other Interface) Interface
	Intersection(other Interface) Interface
	Difference(other Interface) Interface
	IsSubset(other Interface) bool
	Equal(other Interface) bool
	IsNull() bool
}

// data backs the SortedSet with BOTH a slice (kept in ascending comparator
// order) and a map (giving O(1) membership and dedup). The two are kept in
// lock-step: an item lives in order iff it is a key of member, and order is
// always sorted by less.
type data struct {
	less   func(a, b interface{}) bool
	order  []interface{}
	member map[interface{}]struct{}
}

// New creates a SortedSet ordered by the less comparator and seeded with the
// given items, deduplicated and inserted in sorted position. Items must be
// comparable (they back a Go map for membership) AND orderable by less.
func New(less func(a, b interface{}) bool, items ...interface{}) Interface {
	d := &data{
		less:   less,
		order:  []interface{}{},
		member: make(map[interface{}]struct{}),
	}
	for _, item := range items {
		d.Add(item)
	}
	return d
}

// Add inserts item in its sorted position when it is new (a no-op for an
// already-present item, which keeps the set sorted), and returns the receiver so
// calls chain. The insertion index is found with sort.Search over the existing
// order.
func (d *data) Add(item interface{}) Interface {
	if _, ok := d.member[item]; ok {
		return d
	}
	d.member[item] = struct{}{}
	i := sort.Search(len(d.order), func(i int) bool {
		return !d.less(d.order[i], item)
	})
	d.order = append(d.order, nil)
	copy(d.order[i+1:], d.order[i:])
	d.order[i] = item
	return d
}

// Delete removes item from both the slice and the map (a no-op when absent),
// keeping the remaining items sorted, and returns the receiver so calls chain.
func (d *data) Delete(item interface{}) Interface {
	if _, ok := d.member[item]; !ok {
		return d
	}
	delete(d.member, item)
	for i, existing := range d.order {
		if existing == item {
			d.order = append(d.order[:i], d.order[i+1:]...)
			break
		}
	}
	return d
}

// Has reports whether item is a member of the SortedSet.
func (d *data) Has(item interface{}) bool {
	_, ok := d.member[item]
	return ok
}

// Len returns the number of items in the SortedSet.
func (d *data) Len() int {
	return len(d.order)
}

// IsEmpty reports whether the SortedSet has no items.
func (d *data) IsEmpty() bool {
	return len(d.order) == 0
}

// Each iterates over the items in ascending comparator order, invoking fn for
// each. It short-circuits and returns the first Result for which HasError() is
// true; on a full pass it returns a fresh Result.New().
func (d *data) Each(
	fn func(item interface{}) Result.Interface,
) Result.Interface {
	for _, item := range d.order {
		if result := fn(item); result.HasError() {
			return result
		}
	}
	return Result.New()
}

// ToArray materialises the SortedSet into a go-composites Array, in ascending
// comparator order.
func (d *data) ToArray() Array.Interface {
	arr := Array.New()
	for _, item := range d.order {
		arr.Push(item)
	}
	return arr
}

// First returns a Result carrying the minimum item (by the comparator) as its
// payload, or an error Result when the set is empty.
func (d *data) First() Result.Interface {
	if len(d.order) == 0 {
		return Result.New(Result.WithError(Error.New("SortedSet is empty")))
	}
	return Result.New(Result.WithPayload(d.order[0]))
}

// Last returns a Result carrying the maximum item (by the comparator) as its
// payload, or an error Result when the set is empty.
func (d *data) Last() Result.Interface {
	if len(d.order) == 0 {
		return Result.New(Result.WithError(Error.New("SortedSet is empty")))
	}
	return Result.New(Result.WithPayload(d.order[len(d.order)-1]))
}

// Union returns a new SortedSet containing every item present in this set or in
// other (Ruby's `|`). The result uses the RECEIVER's comparator and stays
// sorted.
func (d *data) Union(other Interface) Interface {
	result := New(d.less)
	for _, item := range d.order {
		result.Add(item)
	}
	other.Each(func(item interface{}) Result.Interface {
		result.Add(item)
		return Result.New()
	})
	return result
}

// Intersection returns a new SortedSet containing only the items present in both
// this set and other (Ruby's `&`). The result uses the RECEIVER's comparator and
// stays sorted.
func (d *data) Intersection(other Interface) Interface {
	result := New(d.less)
	for _, item := range d.order {
		if other.Has(item) {
			result.Add(item)
		}
	}
	return result
}

// Difference returns a new SortedSet containing the items present in this set but
// not in other (Ruby's `-`). The result uses the RECEIVER's comparator and stays
// sorted.
func (d *data) Difference(other Interface) Interface {
	result := New(d.less)
	for _, item := range d.order {
		if !other.Has(item) {
			result.Add(item)
		}
	}
	return result
}

// IsSubset reports whether every item of this set is also in other.
func (d *data) IsSubset(other Interface) bool {
	for _, item := range d.order {
		if !other.Has(item) {
			return false
		}
	}
	return true
}

// Equal reports whether this set and other contain exactly the same members.
// Equality is order-INsensitive: two SortedSets are equal when they hold the
// same items regardless of their comparators, exactly like a mathematical set.
func (d *data) Equal(other Interface) bool {
	return d.Len() == other.Len() && d.IsSubset(other)
}

// IsNull reports that this is a real (non-null) SortedSet.
func (d *data) IsNull() bool {
	return false
}

// null is the Null-Object variant of a SortedSet: an empty, immutable
// placeholder that honours the full Interface without ever being nil. Mutating
// methods are no-ops that return the receiver; queries are empty/false/zero, and
// First/Last report an error Result.
type null struct{}

// Null returns the Null-Object SortedSet.
func Null() Interface {
	return &null{}
}

func (n *null) Add(item interface{}) Interface { return n }

func (n *null) Delete(item interface{}) Interface { return n }

func (n *null) Has(item interface{}) bool { return false }

func (n *null) Len() int { return 0 }

func (n *null) IsEmpty() bool { return true }

func (n *null) Each(
	fn func(item interface{}) Result.Interface,
) Result.Interface {
	return Result.New()
}

func (n *null) ToArray() Array.Interface { return Array.New() }

func (n *null) First() Result.Interface {
	return Result.New(Result.WithError(Error.New("SortedSet is empty")))
}

func (n *null) Last() Result.Interface {
	return Result.New(Result.WithError(Error.New("SortedSet is empty")))
}

func (n *null) Union(other Interface) Interface { return n }

func (n *null) Intersection(other Interface) Interface { return n }

func (n *null) Difference(other Interface) Interface { return n }

func (n *null) IsSubset(other Interface) bool { return true }

func (n *null) Equal(other Interface) bool { return other.IsEmpty() }

// IsNull reports that this is the null SortedSet.
func (n *null) IsNull() bool { return true }
