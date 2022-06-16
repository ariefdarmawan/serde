package serde_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ariefdarmawan/serde"
	"github.com/eaciit/toolkit"
)

func TestSerdeObjToObj(t *testing.T) {
	s1 := Struct1{
		ID1:   "ID1_X",
		Data1: "Data1_X",
		Data2: "Data2_X",
		Int2:  804,
		SubPtr: &Sub1{
			Sub2:     809,
			Generic1: int(2022),
		},
	}

	s2 := &Struct2{}

	e := serde.Serde(s1, s2)
	if e != nil {
		t.Fatalf("fail to serde. %s", e.Error())
	}
	if s2.Data1 != s1.Data1 {
		t.Errorf("string parsing error")
	}
	if s2.Int2 != s1.Int2 {
		t.Errorf("int parsing error")
	}
	if s2.SubPtr.Sub2 != s1.SubPtr.Sub2 {
		t.Errorf("Sub ptr error")
	}
	if s2.SubPtr.Generic1 != s1.SubPtr.Generic1 {
		t.Errorf("Generic  error")
	}
	fmt.Println(toolkit.JsonString(s1))
	fmt.Println(toolkit.JsonString(s2))
}

func TestSerdeMapToObj(t *testing.T) {
	s1 := map[string]interface{}{
		"Data1": "Data1_X",
		"Data2": "Data2_X",
		"Int2":  804,
		"Date2": time.Now(),
		"SubPtr": map[string]interface{}{
			"Sub2":     809,
			"Generic1": int(2020),
		},
	}

	s2 := &Struct2{}

	e := serde.Serde(s1, s2)
	if e != nil {
		t.Fatalf("fail to serde. %s", e.Error())
	}
	if s2.Data1 != s1["Data1"].(string) {
		t.Errorf("string parsing error")
	}
	if int(s2.Int2) != s1["Int2"].(int) {
		t.Errorf("int parsing error")
	}
	if s2.SubPtr == nil || s2.SubPtr.Sub2 != s1["SubPtr"].(map[string]interface{})["Sub2"].(int) {
		t.Errorf("Sub ptr error")
	}

	if s2.Date2 == nil || s2.Date2.String() != s1["Date2"].(time.Time).String() {
		t.Errorf("Date error")
	}
	fmt.Println(toolkit.JsonString(s1))
	fmt.Println(toolkit.JsonString(s2))
}

func TestSerdeObjToMap(t *testing.T) {
	date2 := time.Now()
	s1 := Struct1{
		ID1:   "ID1_X",
		Data1: "Data1_X",
		Data2: "Data2_X",
		Int2:  804,
		Date2: &date2,
		SubPtr: &Sub1{
			Sub2:     809,
			Generic1: int(2022),
		},
	}

	s2 := map[string]interface{}{}

	e := serde.Serde(s1, &s2)
	if e != nil {
		t.Fatalf("fail to serde. %s", e.Error())
	}
	if s2["Data1"].(string) != s1.Data1 {
		t.Errorf("string parsing error")
	}
	if s2["Int2"].(int32) != s1.Int2 {
		t.Errorf("int parsing error")
	}
	if s2["SubPtr"].(*Sub1).Sub2 != s1.SubPtr.Sub2 {
		t.Errorf("Sub ptr error")
	}
	if s2["SubPtr"].(*Sub1).Generic1 != s1.SubPtr.Generic1 {
		t.Errorf("Generic  error")
	}
	fmt.Println(toolkit.JsonString(s1))
	fmt.Println(toolkit.JsonString(s2))
}

func TestSerdeMapToMap(t *testing.T) {
	date2 := time.Now()
	s1 := map[string]interface{}{
		"Data1": "Data1_X",
		"Data2": "Data2_X",
		"Int2":  804,
		"Date2": date2,
		"SubPtr": map[string]interface{}{
			"Sub2":     809,
			"Generic1": int(2020),
		},
	}

	s2 := map[string]interface{}{}

	e := serde.Serde(s1, &s2)
	if e != nil {
		t.Fatalf("fail to serde. %s", e.Error())
	}
	if s2["Data1"].(string) != s1["Data1"].(string) {
		t.Errorf("string parsing error")
	}
	if s2["Int2"].(int) != s1["Int2"].(int) {
		t.Errorf("int parsing error")
	}
	if s2["SubPtr"].(map[string]interface{})["Generic1"] != s2["SubPtr"].(map[string]interface{})["Generic1"] {
		t.Errorf("Generic error")
	}
	if s2["Date2"].(time.Time).String() != s1["Date2"].(time.Time).String() {
		t.Errorf("Date error")
	}
	fmt.Println(toolkit.JsonString(s1))
	fmt.Println(toolkit.JsonString(s2))
}

func TestSliceOfMapToSliceOfObj(t *testing.T) {
	ms := []map[string]interface{}{}
	objs := []Struct2{}

	for i := 0; i < 10; i++ {
		ms = append(ms, map[string]interface{}{
			"Date2": time.Now(),
			"SubPtr": map[string]interface{}{
				"Generic1": int32(i * 100),
			},
		})
	}

	e := serde.Serde(ms, &objs)
	if e != nil {
		t.Fatalf("fail to serde. %s", e.Error())
	}

	if len(objs) != len(ms) {
		t.Fatalf("len error")
	}

	if ms[7]["SubPtr"].(map[string]interface{})["Generic1"] != objs[7].SubPtr.Generic1 {
		t.Errorf("Generic error")
	}
	if ms[7]["Date2"].(time.Time).String() != objs[7].Date2.String() {
		t.Errorf("Date error")
	}
}

func TestSliceOfMapToSliceOfPtrObj(t *testing.T) {
	ms := []map[string]interface{}{}
	objs := []*Struct2{}

	for i := 0; i < 10; i++ {
		ms = append(ms, map[string]interface{}{
			"Date2": time.Now(),
			"SubPtr": map[string]interface{}{
				"Generic1": int32(i * 100),
			},
		})
	}

	e := serde.Serde(ms, &objs)
	if e != nil {
		t.Fatalf("fail to serde. %s", e.Error())
	}

	if len(objs) != len(ms) {
		t.Fatalf("len error")
	}

	if ms[7]["SubPtr"].(map[string]interface{})["Generic1"] != objs[7].SubPtr.Generic1 {
		t.Errorf("Generic error")
	}
	if ms[7]["Date2"].(time.Time).String() != objs[7].Date2.String() {
		t.Errorf("Date error")
	}
}

func TestSliceOfObjToSliceOfMap(t *testing.T) {
	ms := []map[string]interface{}{}
	objs := []*Struct2{}

	date2 := time.Now()
	for i := 0; i < 10; i++ {
		objs = append(objs, &Struct2{
			Date2: &date2,
			SubPtr: &Sub1{
				Generic1: int32(i * 100),
			},
		})
	}

	e := serde.Serde(objs, &ms)
	if e != nil {
		t.Fatalf("fail to serde. %s", e.Error())
	}

	if len(objs) != len(ms) {
		t.Fatalf("len error")
	}

	if objs[7].SubPtr.Generic1 != ms[7]["SubPtr"].(*Sub1).Generic1 {
		t.Errorf("Generic error")
	}
	if objs[7].Date2.String() != ms[7]["Date2"].(*time.Time).String() {
		t.Errorf("Date error")
	}
}

type Sub1 struct {
	Sub1     string
	Sub2     int
	Generic1 interface{}

	privateData int
}

type Struct1 struct {
	ID1    string
	Data1  string
	Data2  string
	Int1   int
	Int2   int32
	Int3   int
	Date1  time.Time
	Date2  *time.Time
	Map    map[string]int
	Sub    Sub1
	SubPtr *Sub1
}

type Struct2 struct {
	ID1    string
	Data1  string
	Int1   int
	Int2   int32
	Date2  *time.Time
	Map    map[string]int
	SubPtr *Sub1
}
