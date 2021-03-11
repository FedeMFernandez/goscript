package goscript

import (
	"errors"
	"fmt"
	"reflect"
)

type conector string

//Conectors constants
const (
	ConectorAND conector = "AND"
	ConectorOR  conector = "OR"
)

var (
	errUnprrocessableGivenInterface = errors.New("unprocessable given interface")
	errInvalidOrNil                 = errors.New("interface is invalid or nil")
	errNotSliceMap                  = errors.New("Interface is not map or slice interface")
	errOneConector                  = errors.New("Only one conector can be passed")
	errInvalidParams                = errors.New("some params are invalid")
	errCantSet                      = errors.New("can't set value maybe nil pointer")
)

// FindStruct returns the value of the first element and his index in the provided slice that satisfies
// the provided values in fieldValue param, if the conector is OR will return the first ocurrence that is true.
// If no values satisfies the conditions will return nil.
//
// * Usage
//
// m := make(map[string]interface{})
//
// m["FieldOfStruct"] := Value
//
// * Avalaible logical:
//
// ConectorAND, ConectorOR
func FindStruct(slice interface{}, fieldValue map[string]interface{}, conector conector) (response interface{}, err error) {
	if slice == nil || fieldValue == nil || len(fieldValue) <= 0 {
		err = errInvalidParams
		return
	}
	v, err := valueParser(slice, reflect.Slice)
	if err != nil {
		return
	}
	sliceLen := v.Len()
	for i := 0; i < sliceLen; i++ {
		k := 0
		numField := v.Index(i).NumField()
		for j := 0; j < numField; j++ {
			for c, vv := range fieldValue {
				if v.Index(i).Type().Field(j).Name == c && v.Index(i).Field(j).Interface() == vv {
					switch conector {
					case ConectorOR:
						response = v.Index(i).Interface()
						return
					case ConectorAND:
						k++
						if k == len(fieldValue) {
							response = v.Index(i).Interface()
							return
						}
					}
				}
			}
		}
	}
	return
}

// IndexOf returns the index of an element in a slice, if element is equal to some in the slice
// but is not his reference or if not exists it will return -1 (no-index)
func IndexOf(slice interface{}, element interface{}) (index int, err error) {
	if slice == nil || element == nil {
		err = errInvalidParams
		return
	}
	v, err := valueParser(slice, reflect.Slice)
	if err != nil {
		return
	}
	sliceLen := v.Len()
	for index = 0; index < sliceLen; index++ {
		if reflect.DeepEqual(v.Index(index).Interface(), element) {
			return
		}
	}
	index = -1
	return
}

func typeParser(i interface{}, kind reflect.Kind) (t reflect.Type, err error) {
	t = reflect.TypeOf(i)
	if t.Kind() == reflect.Invalid {
		err = errInvalidOrNil
	}
	for t.Kind() != kind {
		t = t.Elem()
	}
	return
}

func valueParser(i interface{}, kind reflect.Kind) (v reflect.Value, err error) {
	v = reflect.ValueOf(i)
	if v.Kind() == reflect.Invalid {
		err = errInvalidOrNil
	}
	for v.Kind() != kind {
		v = v.Elem()
	}
	return
}

//StructToMap ...
func StructToMap(t reflect.Type, v reflect.Value) (m map[string]interface{}) {
	m = make(map[string]interface{})
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		m[t.Field(i).Name] = v.Field(i).Interface()
	}
	return
}

//MapToStruct ...
func MapToStruct(m map[interface{}]interface{}, structPtr interface{}) (err error) {
	t, err := typeParser(structPtr, reflect.Struct)
	if err != nil {
		return
	}
	v, err := valueParser(structPtr, reflect.Struct)
	if err != nil {
		return
	}
	numField := t.NumField()
	for key, val := range m {
		for i := 0; i < numField; i++ {
			if key == t.Field(i).Name {
				mVal := reflect.ValueOf(val)
				v.Field(i).Set(mVal)
				break
			}
		}
	}
	return
}

func initReflection(v reflect.Value) (value reflect.Value, err error) {
	numField := v.NumField()
	for i := 0; i < numField; i++ {
		if v.Field(i).Kind() == reflect.Ptr {
			v.Field(i).Set(reflect.New(v.Field(i).Type().Elem()))

		}
	}
	value = v
	return
}

func parsePointer(i interface{}) (v reflect.Value, err error) {
	v = reflect.ValueOf(i)
	var oldV reflect.Value
	for v.Kind() == reflect.Ptr {
		oldV = v
		v = v.Elem()
		if !v.IsValid() {
			if oldV.Kind() == reflect.Ptr {
				oldV.Set(reflect.New(oldV.Type().Elem()))
				v = oldV
				continue
			}
		}
	}
	return
}

// Map maps into ptrTo the result of the mapping
// the params should be the same kind, slice or struct
func Map(from, ptrTo interface{}) (err error) {
	v, err := parsePointer(from)
	if err != nil {
		return
	}
	vv, err := parsePointer(ptrTo)
	if err != nil {
		return
	}
	if v.Kind() != vv.Kind() {
		return errInvalidParams
	}
	switch v.Kind() {
	case reflect.Struct:
		ptrTo = mapStruct(v, vv)
		return
	case reflect.Slice:
		ptrTo, err = mapSlice(v, vv)
		return
	default:
		err = errInvalidParams
	}
	return
}

func mapStruct(v, vv reflect.Value) (to interface{}) {
	numFieldFrom := v.NumField()
	numFieldTo := vv.NumField()
	for i := 0; i < numFieldFrom; i++ {
		for j := 0; j < numFieldTo; j++ {
			if v.Type().Field(i).Name != vv.Type().Field(j).Name {
				continue
			}
			if vv.CanSet() {
				if v.Field(i).Kind() == reflect.Ptr {
					t := reflect.TypeOf(v.Field(i).Interface()).Elem()
					if t.Kind() == v.Field(j).Kind() {
						vv.Field(j).Set(v.Field(i).Elem())
						fmt.Println(vv.Field(j).Interface())
					}
				} else if vv.Field(j).Kind() == reflect.Ptr {
					t := reflect.TypeOf(vv.Field(j).Interface()).Elem()
					if t.Kind() == v.Field(i).Kind() {
						vv.Field(j).Set(reflect.New(v.Field(i).Type()))
						vv.Field(j).Elem().Set(v.Field(i))
					}
				} else if v.Field(i).Kind() == vv.Field(j).Kind() {
					vv.Field(j).Set(v.Field(i))
					fmt.Println(vv.Field(j).Interface())
				}
			}
		}
	}
	to = vv.Interface()
	return
}

func mapSlice(sliceValue, sliceTo reflect.Value) (to interface{}, err error) {
	numFieldFrom := sliceValue.Len()
	elem := reflect.New(sliceTo.Type().Elem())
	for i := 0; i < numFieldFrom; i++ {
		var sliceElem reflect.Value
		sliceElem, err = parsePointer(sliceValue.Index(i).Interface())
		if err != nil {
			return
		}
		elem, err = parsePointer(elem.Interface())
		if err != nil {
			return
		}
		var str interface{}
		fmt.Println(str)
		str = mapStruct(sliceElem, elem)
		fmt.Println(str)
		val := reflect.ValueOf(str)
		if sliceTo.Type().Elem().Kind() == reflect.Ptr {
			val = reflect.New(val.Type())
		}
		sliceTo = reflect.Append(sliceTo, val)
	}
	fmt.Println(sliceTo.Interface())
	to = sliceTo.Interface()
	fmt.Println(to)
	return
}
