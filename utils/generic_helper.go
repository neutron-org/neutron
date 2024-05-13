package utils

import (
	"path"
	"reflect"
	"runtime"
)

// MakeNew makes a new instance of generic T.
// if T is a pointer, makes a new instance of the underlying struct via reflection,
// and then a pointer to it.
func MakeNew[T any]() T {
	var v T
	if typ := reflect.TypeOf(v); typ.Kind() == reflect.Ptr {
		elem := typ.Elem()
		//nolint:forcetypeassert
		return reflect.New(elem).Interface().(T) // must use reflect
	}

	return *new(T) // v is not ptr, alloc with new
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0) //nolint:dogsled
	d := path.Join(path.Dir(b), "..")
	return d
}
