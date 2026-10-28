// Licensed under the MIT license, see LICENSE file for details.

package quicktest

import (
	"hash/fnv"
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/google/go-cmp/cmp"
)

// ContentEquals is like DeepEquals but any slices in the compared values will
// be compared regardless of the order of their elements, including nested
// slices.
//
// Before being compared, slices are put in a canonical order: elements are
// sorted by an order-insensitive structural hash, which is the same for
// deeply equal values. Slice types having an Equal method are compared using
// that method instead.
//
// Elements are paired by position after sorting. Values of types with a
// custom Equal method are ordered by their structure, as their Equal method
// cannot be used for ordering: if such values are equal according to their
// Equal method but structurally different (e.g. approximate comparisons, or
// comparisons ignoring some fields), permuted slices of them may not match,
// although slices of them that are already equal in order always match.
// Hashes only inspect values up to a bounded depth, so permuted slices of
// values differing only at a very deep level (e.g. in long linked lists)
// may not match either.
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
		hashes:    make(map[hashKey]uint64),
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

func newSliceID(v reflect.Value) sliceID {
	return sliceID{typ: v.Type(), ptr: v.Pointer(), len: v.Len()}
}

// hashKey identifies a reference value (pointer, slice or map) hashed at a
// given depth with a given hash mode.
type hashKey struct {
	id         sliceID
	depth      int
	structural bool
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
	// hashes memoizes the hashes of reference values, so that shared or
	// cyclic data is not hashed repeatedly.
	hashes map[hashKey]uint64
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
	if hasEqualMethod(vx.Type().Elem()) && equalByMethod(vx, vy) {
		// Elements are compared with their Equal method, which cannot be
		// used for ordering. If they are already equal in order, do not
		// reorder them, which could pair values that are equal according to
		// Equal but structurally different with the wrong elements.
		return false
	}
	// Do not transform slices that are already in canonical order: this also
	// prevents applying the transformation to its own output.
	return !s.isCanonical(vx) || !s.isCanonical(vy)
}

// equalByMethod reports whether the slices x and y, whose elements have an
// Equal method, are equal element by element according to that method.
func equalByMethod(x, y reflect.Value) (eq bool) {
	if x.Len() != y.Len() {
		return false
	}
	defer func() {
		// Equal methods can panic, e.g. when called on nil pointers: let
		// cmp report the problem.
		if recover() != nil {
			eq = false
		}
	}()
	for i := 0; i < x.Len(); i++ {
		if !x.Index(i).MethodByName("Equal").Call([]reflect.Value{y.Index(i)})[0].Bool() {
			return false
		}
	}
	return true
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
	ok = isIdentity(s.order(v))
	s.mu.Lock()
	s.canonical[id] = ok
	s.mu.Unlock()
	return ok
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
	order := s.order(src)
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

// order returns the indexes of the elements of the slice v in canonical
// order. The order is stable: applying it to a slice that is already in
// canonical order returns the identity permutation.
//
// Elements are sorted by their hash, which is the same for values that are
// equal according to cmp. Ties, which happen for hash collisions and for
// values whose Equal method prevents inspecting them, are broken using a
// structural hash which ignores Equal methods.
func (s *contentSorter) order(v reflect.Value) []int {
	n := v.Len()
	order := make([]int, n)
	hashes := make([]uint64, n)
	for i := 0; i < n; i++ {
		order[i] = i
		hashes[i] = s.hash(v.Index(i), 0, false)
	}
	sort.SliceStable(order, func(i, j int) bool {
		return hashes[order[i]] < hashes[order[j]]
	})
	for start := 0; start < n; {
		end := start + 1
		for end < n && hashes[order[end]] == hashes[order[start]] {
			end++
		}
		if end-start > 1 {
			run := order[start:end]
			structural := make(map[int]uint64, len(run))
			for _, i := range run {
				structural[i] = s.hash(v.Index(i), 0, true)
			}
			sort.SliceStable(run, func(i, j int) bool {
				return structural[run[i]] < structural[run[j]]
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

// maxContentHashDepth and maxStructuralHashDepth bound the recursion of
// hash, which also guarantees termination with cyclic data structures.
// Values differing only beyond these depths have the same hash. The
// structural hash, only used to break ties, inspects values more deeply.
// Thanks to memoization, the cost of hashing is linear in these limits.
const (
	maxContentHashDepth    = 32
	maxStructuralHashDepth = 128
)

// hash returns a hash of v such that values that are equal according to
// ContentEquals always have the same hash. Unequal values may also share the
// same hash. Slices are hashed in an order-insensitive way.
//
// When structural is false, values of types having an Equal method, which
// cmp uses for comparison, are not inspected. When structural is true, they
// are inspected as any other value.
//
// Hashes of reference values (pointers, slices and maps) are memoized by
// identity and depth, so that the cost of hashing shared or cyclic data is
// bounded by the number of distinct references times the maximum depth.
func (s *contentSorter) hash(v reflect.Value, depth int, structural bool) uint64 {
	if !v.IsValid() {
		return 0
	}
	t := v.Type()
	h := hashString(t.String())
	maxDepth := maxContentHashDepth
	if structural {
		maxDepth = maxStructuralHashDepth
	}
	if depth > maxDepth || (!structural && hasEqualMethod(t)) {
		return h
	}
	var key hashKey
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice:
		if v.IsNil() {
			return h
		}
		id := sliceID{typ: t, ptr: v.Pointer()}
		if v.Kind() == reflect.Slice {
			id.len = v.Len()
		}
		key = hashKey{id: id, depth: depth, structural: structural}
		s.mu.Lock()
		memo, ok := s.hashes[key]
		s.mu.Unlock()
		if ok {
			return memo
		}
		h = s.hashRef(v, h, depth, structural)
		s.mu.Lock()
		s.hashes[key] = h
		s.mu.Unlock()
		return h
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
	case reflect.Interface:
		if v.IsNil() {
			return h
		}
		return mix(h, s.hash(v.Elem(), depth+1, structural))
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			h = mix(h, s.hash(v.Index(i), depth+1, structural))
		}
		return h
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			h = mix(h, s.hash(v.Field(i), depth+1, structural))
		}
		return h
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		// These are compared by identity by cmp (non-nil functions are never
		// equal).
		return mix(h, uint64(v.Pointer()))
	}
	return h
}

// hashRef hashes the non-nil pointer, map or slice v, starting from h.
func (s *contentSorter) hashRef(v reflect.Value, h uint64, depth int, structural bool) uint64 {
	switch v.Kind() {
	case reflect.Ptr:
		return mix(h, s.hash(v.Elem(), depth+1, structural))
	case reflect.Slice:
		// Order-insensitive: combine element hashes with addition.
		var sum uint64
		for i := 0; i < v.Len(); i++ {
			sum += s.hash(v.Index(i), depth+1, structural)
		}
		return mix(mix(h, uint64(v.Len())), sum)
	default: // reflect.Map
		var sum uint64
		iter := v.MapRange()
		for iter.Next() {
			sum += mix(s.hash(iter.Key(), depth+1, structural), s.hash(iter.Value(), depth+1, structural))
		}
		return mix(mix(h, uint64(v.Len())), sum)
	}
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

// mix combines a hash with a value. Its result is fully avalanched, so that
// order-insensitive sums of hashes are unlikely to collide.
func mix(h, v uint64) uint64 {
	return avalanche(h*0x9e3779b97f4a7c15 ^ avalanche(v))
}

// avalanche is the splitmix64 finalizer.
func avalanche(z uint64) uint64 {
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}
