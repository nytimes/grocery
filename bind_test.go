package grocery

import (
	"strconv"
	"testing"
	"time"
)

type CustomString string
type CustomInt int
type CustomFloat float64

const (
	customStringVal CustomString = "asdf"
	customIntVal    CustomInt    = 5
	customFloatVal  CustomFloat  = 3.2
)

type BindTestEmbeddedStruct struct {
	ES string `grocery:"es"`
}

type bindTestStruct struct {
	// Embedded struct
	BindTestEmbeddedStruct

	// Primitives
	S   string  `grocery:"s"`
	P   *string `grocery:"p"`
	I   int     `grocery:"i"`
	I8  int8    `grocery:"i8"`
	I16 int16   `grocery:"i16"`
	I32 int32   `grocery:"i32"`
	I64 int64   `grocery:"i64"`
	U   uint    `grocery:"u"`
	U8  uint8   `grocery:"u8"`
	U16 uint16  `grocery:"u16"`
	U32 uint32  `grocery:"u32"`
	U64 uint64  `grocery:"u64"`
	F32 float32 `grocery:"f32"`
	F64 float64 `grocery:"f64"`
	B   bool    `grocery:"b"`

	// Custom types
	CS CustomString `grocery:"cs"`
	CI CustomInt    `grocery:"ci"`
	CF CustomFloat  `grocery:"cf"`

	// Timestamp
	T time.Time `grocery:"t"`

	// Don't store
	NoStore     string
	NoStoreDash string `grocery:"-"`

	// Can't set unexported values
	unexported string `grocery:"us"`
}

func TestEmbedding(t *testing.T) {
	data := map[string]string{
		"es": "asdf",
	}

	ptr := new(bindTestStruct)
	err := bind("", "", data, ptr)

	if err != nil {
		t.Errorf("embedding FAILED, got error %v", err)
	}

	if data["es"] == ptr.ES {
		t.Log("embedded string PASSED")
	} else {
		t.Errorf("embedded string FAILED, expected %v but got %v", data["es"], ptr.ES)
	}
}

func TestPrimitives(t *testing.T) {
	data := map[string]string{
		"s":   "asdf",
		"p":   "asdfg",
		"i":   "-4",
		"i8":  "-34",
		"i16": "-5432",
		"i32": "-235873",
		"i64": "-4300000000",
		"u":   "4",
		"u8":  "34",
		"u16": "65432",
		"u32": "235873",
		"u64": "4300000000",
		"f32": "329.14",
		"f64": "-2190.3895",
		"b":   "true",
	}

	ptr := new(bindTestStruct)
	ptr.P = new(string)

	err := bind("", "", data, ptr)

	if err != nil {
		t.Errorf("primitives FAILED, got error %v", err)
	}

	if data["s"] != ptr.S {
		t.Errorf("string FAILED, expected %v but got %v", data["s"], ptr.S)
	}

	if data["p"] != *ptr.P {
		t.Errorf("string pointer FAILED, expected %v but got %v", data["p"], *ptr.P)
	}

	iVal, _ := strconv.ParseInt(data["i"], 10, 64)
	if int(iVal) != ptr.I {
		t.Errorf("int FAILED, expected %v but got %v", data["i"], ptr.I)
	}

	i8Val, _ := strconv.ParseInt(data["i8"], 10, 8)
	if int8(i8Val) != ptr.I8 {
		t.Errorf("int8 FAILED, expected %v but got %v", data["i8"], ptr.I8)
	}

	i16Val, _ := strconv.ParseInt(data["i16"], 10, 16)
	if int16(i16Val) != ptr.I16 {
		t.Errorf("int16 FAILED, expected %v but got %v", data["i16"], ptr.I16)
	}

	i32Val, _ := strconv.ParseInt(data["i32"], 10, 32)
	if int32(i32Val) != ptr.I32 {
		t.Errorf("int32 FAILED, expected %v but got %v", data["i32"], ptr.I32)
	}

	i64Val, _ := strconv.ParseInt(data["i64"], 10, 64)
	if int64(i64Val) != ptr.I64 {
		t.Errorf("int64 FAILED, expected %v but got %v", data["i64"], ptr.I64)
	}

	uVal, _ := strconv.ParseUint(data["u"], 10, 64)
	if uint(uVal) != ptr.U {
		t.Errorf("uint FAILED, expected %v but got %v", data["u"], ptr.U)
	}

	u8Val, _ := strconv.ParseUint(data["u8"], 10, 8)
	if uint8(u8Val) != ptr.U8 {
		t.Errorf("uint8 FAILED, expected %v but got %v", data["u8"], ptr.U8)
	}

	u16Val, _ := strconv.ParseUint(data["u16"], 10, 16)
	if uint16(u16Val) != ptr.U16 {
		t.Errorf("uint16 FAILED, expected %v but got %v", data["u16"], ptr.U16)
	}

	u32Val, _ := strconv.ParseUint(data["u32"], 10, 32)
	if uint32(u32Val) != ptr.U32 {
		t.Errorf("uint32 FAILED, expected %v but got %v", data["u32"], ptr.U32)
	}

	u64Val, _ := strconv.ParseUint(data["u64"], 10, 64)
	if uint64(u64Val) != ptr.U64 {
		t.Errorf("uint64 FAILED, expected %v but got %v", data["u64"], ptr.U64)
	}

	f32Val, _ := strconv.ParseFloat(data["f32"], 32)
	if float32(f32Val) != ptr.F32 {
		t.Errorf("float32 FAILED, expected %v but got %v", data["f32"], ptr.F32)
	}

	f64Val, _ := strconv.ParseFloat(data["f64"], 64)
	if float64(f64Val) != ptr.F64 {
		t.Errorf("float64 FAILED, expected %v but got %v", data["f64"], ptr.F64)
	}

	bVal, _ := strconv.ParseBool(data["b"])
	if bVal != ptr.B {
		t.Errorf("bool FAILED, expected %v but got %v", data["b"], ptr.B)
	}
}

func TestCustomMapType(t *testing.T) {
	data := map[string]string{
		"cs": string(customStringVal),
		"ci": strconv.Itoa(int(customIntVal)),
		"cf": strconv.FormatFloat(float64(customFloatVal), 'f', 10, 64),
	}

	ptr := new(bindTestStruct)
	err := bind("", "", data, ptr)

	if err != nil {
		t.Errorf("custom map type FAILED, got error %v", err)
	}

	if data["cs"] == string(ptr.CS) {
		t.Log("custom string PASSED")
	} else {
		t.Errorf("custom string FAILED, expected %v but got %v", data["cs"], ptr.CS)
	}

	ciVal, _ := strconv.ParseInt(data["ci"], 10, 64)
	if int(ciVal) == int(ptr.CI) {
		t.Log("custom int PASSED")
	} else {
		t.Errorf("custom int FAILED, expected %v but got %v", data["ci"], ptr.CI)
	}

	cfVal, _ := strconv.ParseFloat(data["cf"], 64)
	if cfVal == float64(ptr.CF) {
		t.Log("custom float PASSED")
	} else {
		t.Errorf("custom float FAILED, expected %v but got %v", data["cf"], ptr.CF)
	}
}

func TestTimestamp(t *testing.T) {
	data := map[string]string{
		"t": "963210120",
	}

	ptr := new(bindTestStruct)
	err := bind("", "", data, ptr)

	if err != nil {
		t.Errorf("timestamp FAILED, got error %v", err)
	}

	tVal, _ := strconv.ParseInt(data["t"], 10, 64)
	if tVal == ptr.T.Unix() {
		t.Log("timestamp PASSED")
	} else {
		t.Errorf("timestamp FAILED, expected %v but got %v", data["t"], ptr.T)
	}
}

func TestInvalidValue(t *testing.T) {
	data := map[string]string{
		"i": "asdf",
	}

	ptr := new(bindTestStruct)
	err := bind("", "", data, ptr)

	if err != nil {
		t.Log("invalid value PASSED")
	} else {
		t.Error("invalid value FAILED, expecting error")
	}
}

func TestInvalidPointer(t *testing.T) {
	data := map[string]string{
		"s": "asdf",
	}

	err := bind("", "", data, nil)

	if err != nil {
		t.Log("nil pointer PASSED")
	} else {
		t.Error("nil pointer FAILED, expecting error")
	}

	err = bind("", "", data, 4)

	if err != nil {
		t.Log("nil pointer PASSED")
	} else {
		t.Error("nil pointer FAILED, expecting error")
	}
}

func TestEmptyData(t *testing.T) {
	emptyData := map[string]string{}
	ptr := new(bindTestStruct)
	err := bind("", "", emptyData, ptr)

	if err != nil {
		t.Log("empty data PASSED")
	} else {
		t.Error("empty data FAILED, expecting error")
	}
}
