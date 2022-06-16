package serde

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/sebarcode/codekit"
)

var (
	tagName    string
	dateFormat string
)

func TagName() string {
	if tagName == "" {
		tagName = codekit.TagName()
	}
	return tagName
}

func SetTagName(nm string) {
	tagName = nm
}

func DateFormat() string {
	if dateFormat == "" {
		dateFormat = time.RFC3339
	}
	return dateFormat
}

func SetDateFormat(fmt string) {
	dateFormat = fmt
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
