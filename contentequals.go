// Licensed under the MIT license, see LICENSE file for details.

package quicktest

import (
	"hash/fnv"
	"math"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kr/pretty"
)

// ContentEquals is like DeepEquals but any slices in the compared values will
// be compared regardless of the order of their elements: two slices are
// considered equal when each element of one can be paired with a distinct,
// deeply equal element of the other. This also applies to nested slices.
var ContentEquals Checker = &contentEqualsChecker{
	argNames: []string{"got", "want"},
}

type contentEqualsChecker struct {
	argNames
}

// Check implements Checker.Check by checking that got == args[0] regardless
// of the order of slice elements.
func (c *contentEqualsChecker) Check(got interface{}, args []interface{}, note func(key string, value interface{})) (err error) {
	defer func() {
		// See cmpEqualsChecker.Check.
		if r := recover(); r != nil {
			err = BadCheckf("%s", r)
		}
	}()
	want := args[0]
	if cmp.Equal(want, got, unorderedSlices()) {
		return nil
	}
	// Sort slices by their printed representation to produce a readable
	// diff. This is only used for output: the equality check above does not
	// rely on printed representations.
	diff := cmp.Diff(want, got, sortSlicesForDiff)
	if diff == "" {
		diff = cmp.Diff(want, got, unorderedSlices())
	}
	note("error", Unquoted("values are not deep equal"))
	note("diff (-want +got)", Unquoted(diff))
	note("got", SuppressedIfLong{got})
	note("want", SuppressedIfLong{want})
	return ErrSilent
}

// sortSlicesForDiff sorts slices by their printed representation.
var sortSlicesForDiff = cmpopts.SortSlices(func(x, y interface{}) bool {
	return pretty.Sprint(x) < pretty.Sprint(y)
})

// unorderedSlices returns a cmp option comparing slices as multisets.
func unorderedSlices() cmp.Option {
	var opt cmp.Option
	opt = cmp.FilterValues(func(x, y interface{}) bool {
		vx, vy := reflect.ValueOf(x), reflect.ValueOf(y)
		return vx.IsValid() && vy.IsValid() && vx.Kind() == reflect.Slice && vx.Type() == vy.Type()
	}, cmp.Comparer(func(x, y interface{}) bool {
		return equalUnordered(reflect.ValueOf(x), reflect.ValueOf(y), opt)
	}))
	return opt
}

// equalUnordered reports whether the slices x and y hold the same elements,
// in any order, comparing elements with cmp.Equal and the given option.
//
// Since deep equality is an equivalence relation, greedily pairing each
// element of x with any equal unpaired element of y is enough. To avoid
// comparing every pair of elements, candidates are grouped by a structural
// hash that is guaranteed to be the same for equal values.
func equalUnordered(x, y reflect.Value, opt cmp.Option) bool {
	if x.IsNil() != y.IsNil() || x.Len() != y.Len() {
		return false
	}
	n := x.Len()
	var pending []int
	// Fast path: pair elements that are already in the same position.
	for i := 0; i < n; i++ {
		if !cmp.Equal(x.Index(i).Interface(), y.Index(i).Interface(), opt) {
			pending = append(pending, i)
		}
	}
	if len(pending) == 0 {
		return true
	}
	buckets := make(map[uint64][]int, len(pending))
	for _, j := range pending {
		h := contentHash(y.Index(j), 0)
		buckets[h] = append(buckets[h], j)
	}
outer:
	for _, i := range pending {
		xi := x.Index(i)
		h := contentHash(xi, 0)
		js := buckets[h]
		xv := xi.Interface()
		for k, j := range js {
			if cmp.Equal(xv, y.Index(j).Interface(), opt) {
				// Remove j from the bucket, preserving order.
				buckets[h] = append(js[:k:k], js[k+1:]...)
				continue outer
			}
		}
		return false
	}
	return true
}

// maxContentHashDepth bounds the recursion of contentHash, which also
// protects against cyclic data structures.
const maxContentHashDepth = 10

// contentHash returns a hash of v such that values that are equal according
// to cmp.Equal with the unorderedSlices option always have the same hash.
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
// values of type t.
func hasEqualMethod(t reflect.Type) bool {
	if _, ok := t.MethodByName("Equal"); ok {
		return true
	}
	if t.Kind() != reflect.Ptr && t.Kind() != reflect.Interface {
		if _, ok := reflect.PtrTo(t).MethodByName("Equal"); ok {
			return true
		}
	}
	return false
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
