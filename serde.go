package serde

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
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

	return CopyValue(vSource, vDest.Elem())
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

func CopyValue(source, dest reflect.Value) error {
	if source.Kind() == reflect.Ptr {
		return CopyValue(source.Elem(), dest)
	}

	if dest.Kind() == reflect.Ptr {
		if dest.IsNil() {
			return errors.New("destination is nil")
		}
		return CopyValue(source, dest.Elem())
	}

	sourceTypeName := source.Type().String()
	destTypeName := dest.Type().String()

	if sourceTypeName == destTypeName {
		dest.Set(source)
		return nil
	}

	sourceIsInterface := sourceTypeName == "interface {}"
	sourceIsMap := source.Kind() == reflect.Map

	destIsMap := dest.Kind() == reflect.Map
	destIsStruct := dest.Kind() == reflect.Struct

	if !sourceIsInterface {
		if destIsMap {
			return copyValueToMap(source, dest, sourceIsMap, false)
		} else if destIsStruct {
			return copyValueToStruct(source, dest, sourceIsMap, false)
		}
	} else if destIsMap || destIsStruct {
		sourceData := source.Interface()
		return CopyValue(reflect.ValueOf(sourceData), dest)
	}

	var e error
	func() {
		defer RecoverToError(&e)
		sourceType := source.Type().String()
		destType := dest.Type().String()

		if sourceType == destType {
			dest.Set(source)
			return
		}

		var data interface{}

		switch destType {
		case "float32":
			data = float32(source.Interface().(float64))
		case "int8":
			data = int8(source.Interface().(int))
		case "int16":
			data = int16(source.Interface().(int))
		case "int32":
			data = int32(source.Interface().(int))
		case "int64":
			data = int64(source.Interface().(int))
		default:
			data = source.Interface()
		}
		dest.Set(reflect.ValueOf(data))
	}()
	return e
}

func copyValueToStruct(source, dest reflect.Value, sourceIsMap, ignoreError bool) error {
	destType := dest.Type()
	fieldCount := destType.NumField()
	for i := 0; i < fieldCount; i++ {
		fieldMeta := destType.Field(i)
		sourceField := getFieldFromRV(source, fieldMeta.Name, sourceIsMap)
		if sourceField.IsValid() && !sourceField.IsZero() {
			if !fieldMeta.IsExported() {
				return fmt.Errorf("fail processing %s. it is an unexported field", fieldMeta.Name)
			}

			var eSet error
			destField := dest.FieldByName(fieldMeta.Name)
			if destField.Kind() == reflect.Ptr {
				bufferPtr := CreatePtrFromType(fieldMeta.Type)
				eSet = CopyValue(sourceField, bufferPtr)
				if eSet == nil {
					destField.Set(bufferPtr)
				}
			} else {
				eSet = CopyValue(sourceField, destField)
			}
			if eSet != nil {
				if !ignoreError {
					return fmt.Errorf("fail processing %s. %ss", fieldMeta.Name, eSet.Error())
				} else {
					fmt.Printf("fail processing %s. %ss", fieldMeta.Name, eSet.Error())
				}
			}
		}
	}
	return nil
}

func copyValueToMap(source, dest reflect.Value, sourceIsMap, ignoreError bool) error {
	keys := []reflect.Value{}
	sourceType := source.Type()
	fieldCount := 0
	if sourceIsMap {
		keys = source.MapKeys()
		fieldCount = len(keys)
	} else {
		fieldCount = source.NumField()
	}

	for i := 0; i < fieldCount; i++ {
		var (
			sourceField reflect.Value
			eSet        error
			fieldName   = ""
		)

		func() {
			defer RecoverToError(&eSet)
			var key reflect.Value
			if sourceIsMap {
				key = keys[i]
				fieldName = fmt.Sprintf("%v", key.Interface())
				sourceField = source.MapIndex(key)
			} else {
				fieldName = sourceType.Field(i).Name
				sourceField = source.Field(i)
				key = reflect.ValueOf(fieldName)
			}
			if sourceField.IsValid() && !sourceField.IsZero() {
				//fmt.Println("data:", sourceField.Interface())
				dest.SetMapIndex(key, sourceField)
			}
		}()

		if eSet != nil {
			if !ignoreError {
				return fmt.Errorf("fail processing %s. %ss", fieldName, eSet.Error())
			} else {
				fmt.Printf("fail processing %s. %ss", fieldName, eSet.Error())
			}
		}
	}

	return nil
}

func CreatePtrFromType(t reflect.Type) reflect.Value {
	isPtr := t.Kind() == reflect.Ptr
	elemType := t

	if isPtr {
		elemType = elemType.Elem()
	}

	if elemType.Kind() == reflect.Map {
		ptr := reflect.New(elemType)
		m := reflect.MakeMap(elemType)
		ptr.Elem().Set(m)
		return ptr
	}

	return reflect.New(elemType)
}

func getFieldFromRV(rv reflect.Value, name string, isMap bool) reflect.Value {
	if isMap {
		return rv.MapIndex(reflect.ValueOf(name))
	}
	return rv.FieldByName(name)
}

func RecoverToError(e *error) {
	if r := recover(); r != nil {
		switch r.(type) {
		case *reflect.ValueError:
			ve := r.(*reflect.ValueError)
			*e = errors.New(ve.Error() + " " + string(debug.Stack()))

		case string:
			*e = errors.New(r.(string) + " " + string(debug.Stack()))

		default:
			*e = errors.New(fmt.Sprintf("%v", r) + " " + string(debug.Stack()))
		}
	}
}
