package serde

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sebarcode/codekit"
)

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
	//sourceIsDate := source.Type().String() == "time.Time"
	destIsDate := dest.Type().String() == "time.Time"

	destIsMap := dest.Kind() == reflect.Map
	destIsStruct := dest.Kind() == reflect.Struct

	//--- not a time.Time
	if !destIsDate {
		//--- not interface
		if !sourceIsInterface {
			if destIsMap {
				return copyValueToMap(source, dest, sourceIsMap, false)
			} else if destIsStruct {
				return copyValueToStruct(source, dest, sourceIsMap, false)
			}

			// is not primitive type
		} else if destIsMap || destIsStruct {
			sourceData := source.Interface()
			return CopyValue(reflect.ValueOf(sourceData), dest)
		}
	}

	var e error
	func() {
		defer RecoverToError(&e)

		if sourceTypeName == destTypeName {
			dest.Set(source)
			return
		}

		var data interface{}
		switch destTypeName {
		case "time.Time":
			// will only go here if dest is a time.Time but source is not. Most likely will be []byte or string or int
			raw := source.Interface()
			switch raw.(type) {
			case int:
				data = time.UnixMicro(int64(raw.(int)))
			case int64:
				data = time.UnixMicro(raw.(int64))
			case []byte:
				dateStr := string(raw.([]byte))
				if dt, err := time.Parse(time.RFC3339, dateStr); err == nil {
					data = dt
				} else {
					data = codekit.String2Date(dateStr, DateFormat())
				}
			default:
				dateStr := strings.Replace(codekit.JsonString(raw), "\"", "", -1)
				//fmt.Println(dateStr)
				if dt, err := time.Parse(time.RFC3339, dateStr); err == nil {
					data = dt
				} else {
					data = codekit.String2Date(dateStr, DateFormat())
				}
			}
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
		case "int":
			data = codekit.ToInt(source.Interface(), codekit.RoundingAuto)
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
		if !fieldMeta.IsExported() {
			continue
			//return fmt.Errorf("fail processing %s. it is an unexported field", fieldMeta.Name)
		}

		var sourceField reflect.Value
		originalFieldName := fieldMeta.Name
		if sourceIsMap {
			alias := fieldMeta.Tag.Get(TagName())
			if alias == "" {
				alias = originalFieldName
			}
			sourceField = getFieldFromRV(source, alias, sourceIsMap)
		} else {
			sourceField = getFieldFromRV(source, originalFieldName, sourceIsMap)
		}
		if sourceField.IsValid() && !sourceField.IsZero() {
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
				meta := sourceType.Field(i)
				fieldName = meta.Name
				alias := meta.Tag.Get(TagName())
				if alias == "" {
					alias = fieldName
				}
				sourceField = source.Field(i)
				key = reflect.ValueOf(alias)
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
