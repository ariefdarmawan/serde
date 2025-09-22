package serde

import (
	"errors"
	"fmt"
	"reflect"
)

func Serde(source, dest interface{}) error {
	vSource := reflect.Indirect(reflect.ValueOf(source))
	isSourceSlice := vSource.Kind() == reflect.Slice

	vDest := reflect.ValueOf(dest)
	vDestKind := vDest.Kind()
	if vDestKind != reflect.Ptr {
		return errors.New("destination should be a pointer")
	}
	vDestKind = vDest.Elem().Kind()

	isDestSlice := vDestKind == reflect.Slice
	if isSourceSlice && !isDestSlice {
		return errors.New("destination should be a slice")
	}

	if isSourceSlice {
		return SerdeSlice(vSource, vDest)
	}

	vDestElem := vDest.Elem()
	return CopyValue(vSource, vDestElem)
}

func SerdeSlice(source, dest reflect.Value) error {
	if dest.Kind() == reflect.Ptr {
		return SerdeSlice(source, dest.Elem())
	}

	sliceType := dest.Type()
	elemType := sliceType.Elem()
	elemIsPtr := elemType.Kind() == reflect.Ptr
	sourceLen := source.Len()
	destBuffer := reflect.MakeSlice(sliceType, sourceLen, sourceLen)
	for i := 0; i < sourceLen; i++ {
		sourceItem := source.Index(i)
		destItem := CreatePtrFromType(elemType)
		if e := CopyValue(sourceItem, destItem.Elem()); e != nil {
			return fmt.Errorf("errors processing index %d, %s", i, e.Error())
		}
		if elemIsPtr {
			destBuffer.Index(i).Set(destItem)
		} else {
			destBuffer.Index(i).Set(destItem.Elem())
		}
	}
	dest.Set(destBuffer)
	return nil
}

func getFieldFromRV(rv reflect.Value, name string, isMap bool) reflect.Value {
	if isMap {
		return rv.MapIndex(reflect.ValueOf(name))
	}
	return rv.FieldByName(name)
}
