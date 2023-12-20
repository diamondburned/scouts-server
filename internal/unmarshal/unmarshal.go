package unmarshal

import (
	"encoding"
	"reflect"
)

// Text unmarshals text into a type that implements encoding.TextUnmarshaler.
func Text[T encoding.TextUnmarshaler](text string) (T, error) {
	var t T
	rtype := reflect.TypeOf(t)
	if rtype == nil {
		panic("unmarshal: cannot unmarshal into interface type")
	}
	rtype = rtype.Elem()
	t = reflect.New(rtype).Interface().(T)
	err := t.UnmarshalText([]byte(text))
	return t, err
}
