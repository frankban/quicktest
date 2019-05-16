package quicktest

import (
	"fmt"
	"reflect"
)

type containerIter interface {
	next() bool
	key() string
	value() reflect.Value
}

func newIter(x interface{}) (containerIter, error) {
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Map:
		return newMapIter(v), nil
	case reflect.Slice, reflect.Array:
		return &sliceIter{
			index: -1,
			v:     v,
		}, nil
	default:
		return nil, fmt.Errorf("map, slice or array required")
	}
}

type sliceIter struct {
	v     reflect.Value
	index int
}

func (i *sliceIter) next() bool {
	i.index++
	return i.index < i.v.Len()
}

func (i *sliceIter) value() reflect.Value {
	return i.v.Index(i.index)
}

func (i *sliceIter) key() string {
	return fmt.Sprintf("index %d", i.index)
}
