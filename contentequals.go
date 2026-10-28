// Licensed under the MIT license, see LICENSE file for details.

package quicktest

import (
	"hash/fnv"
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/kr/pretty"
)

// ContentEquals is like DeepEquals but any slices in the compared values will
// be compared regardless of the order of their elements, including nested
// slices.
//
// Before being compared, slices are put in a canonical order: elements are
// sorted by an order-insensitive structural hash, which is the same for
// deeply equal values, and elements with the same hash are sorted by their
// printed representation. Slice types having an Equal method are compared
// using that method instead. Elements are paired by position after sorting,
// so slices containing values compared with custom, non-transitive Equal
// methods (e.g. approximate comparisons) may not be matched.
var ContentEquals Checker = &contentEqualsChecker{
	argNames: []string{"got", "want"},
}

type contentEqualsChecker struct {
	argNames
}

// Check implements Checker.Check by checking that got == args[0] after
// putting all slices in canonical order.
func (c *contentEqualsChecker) Check(got interface{}, args []interface{}, note func(key string, value interface{})) error {
	// A new sorter is used for each check, so that its memoization does not
	// outlive the compared values.
	s := &contentSorter{
		canonical: make(map[sliceID]bool),
		sorted:    make(map[sliceID]interface{}),
	}
	checker := &cmpEqualsChecker{
		opts: cmp.Options{
			cmp.FilterValues(s.filter, cmp.Transformer("quicktest.ContentEquals", s.sort)),
		},
	}
	return checker.Check(got, args, note)
}

// sliceID identifies a slice value.
type sliceID struct {
	typ reflect.Type
	ptr uintptr
	len int
}

// contentSorter puts slices in canonical order. It memoizes its results,
// so it must only be used for a single comparison.
type contentSorter struct {
	// mu protects the fields below: go-cmp may call transformers
	// concurrently to detect races.
	mu sync.Mutex
	// canonical records whether slices are already in canonical order.
	canonical map[sliceID]bool
	// sorted maps slices to their already computed sorted copies. Returning
	// the same copy for the same slice allows go-cmp to detect cycles in
	// the compared values.
	sorted map[sliceID]interface{}
}

// filter reports whether the slices x and y must be put in canonical order
// before being compared.
func (s *contentSorter) filter(x, y interface{}) bool {
	vx, vy := reflect.ValueOf(x), reflect.ValueOf(y)
	if !vx.IsValid() || !vy.IsValid() || vx.Type() != vy.Type() || vx.Kind() != reflect.Slice {
		return false
	}
	if vx.Len() <= 1 && vy.Len() <= 1 {
		return false
	}
	if hasEqualMethod(vx.Type()) {
		// Let cmp use the Equal method.
		return false
	}
	// Do not transform slices that are already in canonical order: this also
	// prevents applying the transformation to its own output.
	return !s.isCanonical(vx) || !s.isCanonical(vy)
}

// isCanonical reports whether the slice v is already in canonical order.
func (s *contentSorter) isCanonical(v reflect.Value) bool {
	id := newSliceID(v)
	s.mu.Lock()
	ok, found := s.canonical[id]
	s.mu.Unlock()
	if found {
		return ok
	}
	ok = isIdentity(contentOrder(v))
	s.mu.Lock()
	s.canonical[id] = ok
	s.mu.Unlock()
	return ok
}

func newSliceID(v reflect.Value) sliceID {
	return sliceID{typ: v.Type(), ptr: v.Pointer(), len: v.Len()}
}

// sort returns a copy of the slice x in canonical order.
func (s *contentSorter) sort(x interface{}) interface{} {
	src := reflect.ValueOf(x)
	id := newSliceID(src)
	s.mu.Lock()
	sorted, ok := s.sorted[id]
	s.mu.Unlock()
	if ok {
		return sorted
	}
	order := contentOrder(src)
	dst := reflect.MakeSlice(src.Type(), len(order), len(order))
	for i, j := range order {
		dst.Index(i).Set(src.Index(j))
	}
	sorted = dst.Interface()
	s.mu.Lock()
	if prev, ok := s.sorted[id]; ok {
		sorted = prev
	} else {
		s.sorted[id] = sorted
	}
	s.mu.Unlock()
	return sorted
}

// contentOrder returns the indexes of the elements of the slice v in
// canonical order. The order is stable: applying it to a slice that is
// already in canonical order returns the identity permutation.
func contentOrder(v reflect.Value) []int {
	n := v.Len()
	order := make([]int, n)
	hashes := make([]uint64, n)
	for i := 0; i < n; i++ {
		order[i] = i
		hashes[i] = contentHash(v.Index(i), 0)
	}
	sort.SliceStable(order, func(i, j int) bool {
		return hashes[order[i]] < hashes[order[j]]
	})
	// Sort runs of elements with the same hash by their printed
	// representation, computed only for those elements.
	for start := 0; start < n; {
		end := start + 1
		for end < n && hashes[order[end]] == hashes[order[start]] {
			end++
		}
		if end-start > 1 {
			run := order[start:end]
			printed := make(map[int]string, len(run))
			for _, i := range run {
				printed[i] = pretty.Sprint(v.Index(i).Interface())
			}
			sort.SliceStable(run, func(i, j int) bool {
				return printed[run[i]] < printed[run[j]]
			})
		}
		start = end
	}
	return order
}

func isIdentity(order []int) bool {
	for i, j := range order {
		if i != j {
			return false
		}
	}
	return true
}

// maxContentHashDepth bounds the recursion of contentHash, which also
// protects against cyclic data structures.
const maxContentHashDepth = 10

// contentHash returns a hash of v such that values that are equal according
// to ContentEquals always have the same hash.
// Unequal values may also share the same hash.
//
// Slices are hashed in an order-insensitive way. Values of types having an
// Equal method, which cmp uses for comparison, are not inspected.
func contentHash(v reflect.Value, depth int) uint64 {
	if !v.IsValid() {
		return 0
	}
	t := v.Type()
	h := hashString(t.String())
	if depth > maxContentHashDepth || hasEqualMethod(t) {
		return h
	}
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return mix(h, 1)
		}
		return mix(h, 2)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return mix(h, uint64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return mix(h, v.Uint())
	case reflect.Float32, reflect.Float64:
		return mix(h, hashFloat(v.Float()))
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		return mix(mix(h, hashFloat(real(c))), hashFloat(imag(c)))
	case reflect.String:
		return mix(h, hashString(v.String()))
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return h
		}
		return mix(h, contentHash(v.Elem(), depth+1))
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			h = mix(h, contentHash(v.Index(i), depth+1))
		}
		return h
	case reflect.Slice:
		// Order-insensitive: combine element hashes with addition.
		var sum uint64
		for i := 0; i < v.Len(); i++ {
			sum += contentHash(v.Index(i), depth+1)
		}
		return mix(mix(h, uint64(v.Len())), sum)
	case reflect.Map:
		var sum uint64
		iter := v.MapRange()
		for iter.Next() {
			sum += mix(contentHash(iter.Key(), depth+1), contentHash(iter.Value(), depth+1))
		}
		return mix(mix(h, uint64(v.Len())), sum)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			h = mix(h, contentHash(v.Field(i), depth+1))
		}
		return h
	}
	// Functions, channels and unsafe pointers.
	return h
}

// hasEqualMethod reports whether cmp would use an Equal method to compare
// values of type t. This mirrors the rule used by go-cmp: t must have a
// method of the form Equal(I) bool, where t is assignable to I.
func hasEqualMethod(t reflect.Type) bool {
	m, ok := t.MethodByName("Equal")
	if !ok {
		return false
	}
	// For non-interface types, the method type includes the receiver.
	// Methods of interface types never match, as in go-cmp.
	mt := m.Type
	return !mt.IsVariadic() &&
		mt.NumIn() == 2 && mt.NumOut() == 1 &&
		mt.In(0).AssignableTo(mt.In(1)) &&
		mt.Out(0) == boolType
}

func hashFloat(f float64) uint64 {
	if f == 0 {
		// Normalize negative zero, which compares equal to zero.
		f = 0
	}
	return math.Float64bits(f)
}

func hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// mix combines a hash with a value.
func mix(h, v uint64) uint64 {
	h ^= v + 0x9e3779b97f4a7c15 + (h << 6) + (h >> 2)
	return h * 0x100000001b3
}
