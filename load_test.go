package grocery

import (
	"testing"
	"time"
)

type StringAlias string

type LoadTestModel struct {
	Base

	StringVal  string
	IntVal     int
	UInt16Val  uint16
	BoolVal    bool
	Float32Val float32
	Float64Val float64

	StringAliasVal StringAlias
	MapVal         *Map
	SetVal         *Set
}

func TestLoad(t *testing.T) {
	mapData := map[string]string{"a": "b", "c": "d"}
	setData := []string{"a", "b"}

	model := &LoadTestModel{
		StringVal:      "hello world",
		IntVal:         4,
		UInt16Val:      3,
		BoolVal:        true,
		Float32Val:     3.5,
		Float64Val:     3.9,
		StringAliasVal: "asdf",
		MapVal:         newMap(mapData),
		SetVal:         newSet(setData),
	}

	id, err := Store(model)
	storeTime := time.Now()

	if err != nil {
		t.Error(err)
	}

	loadedModel := new(LoadTestModel)
	err = Load(id, loadedModel)

	if err != nil {
		t.Error(err)
	}

	if loadedModel.ID != id {
		t.Errorf("id FAILED, expected %s but got %s", id, model.ID)
	}

	if time.Since(loadedModel.CreatedAt) < time.Since(storeTime) {
		t.Errorf("createdAt FAILED")
	}

	if time.Since(loadedModel.UpdatedAt) < time.Since(storeTime) {
		t.Errorf("updatedAt FAILED")
	}

	if model.StringVal != loadedModel.StringVal {
		t.Errorf("string FAILED, expected %s but got %s", model.StringVal, loadedModel.StringVal)
	}

	if model.IntVal != loadedModel.IntVal {
		t.Errorf("int FAILED, expected %d but got %d", model.IntVal, loadedModel.IntVal)
	}

	if model.UInt16Val != loadedModel.UInt16Val {
		t.Errorf("uint16 FAILED, expected %d but got %d", model.UInt16Val, loadedModel.UInt16Val)
	}

	if model.BoolVal != loadedModel.BoolVal {
		t.Errorf("bool FAILED, expected %t but got %t", model.BoolVal, loadedModel.BoolVal)
	}

	if model.Float32Val != loadedModel.Float32Val {
		t.Errorf("float32 FAILED, expected %f but got %f", model.Float32Val, loadedModel.Float32Val)
	}

	if model.Float64Val != loadedModel.Float64Val {
		t.Errorf("float64 FAILED, expected %f but got %f", model.Float32Val, loadedModel.Float32Val)
	}

	if model.StringAliasVal != loadedModel.StringAliasVal {
		t.Errorf("string alias FAILED, expected %s but got %s", model.StringAliasVal, loadedModel.StringAliasVal)
	}

	for k, v := range mapData {
		loadedValue, ok := loadedModel.MapVal.Load(k)

		if !ok {
			t.Errorf("map value FAILED, key %s does not exist", k)
			continue
		}

		if v != loadedValue.(string) {
			t.Errorf("map value FAILED, expected m[%s]=%s but got m[%s]=%s", k, v, k, loadedValue.(string))
		}
	}

	for _, k := range setData {
		if !loadedModel.SetVal.Contains(k) {
			t.Errorf("set value FAILED, key %s does not exist", k)
		}
	}
}

func TestLoadMissing(t *testing.T) {
	model := new(LoadTestModel)
	err := Load("asdf", model)

	if err == nil {
		t.Error("Should throw error for missing models but did not")
	}
}

func TestLoadAll(t *testing.T) {
	model := &LoadTestModel{
		StringVal:      "hello world",
		IntVal:         4,
		BoolVal:        true,
		StringAliasVal: "asdf",
		MapVal:         newMap(map[string]string{"a": "b", "c": "d"}),
	}

	id, err := Store(model)

	if err != nil {
		t.Error(err)
	}

	loadedModels := make([]LoadTestModel, 1)
	err = LoadAll([]string{id}, &loadedModels)

	if err != nil {
		t.Error(err)
	}

	if loadedModels[0].ID != id {
		t.Errorf("id FAILED, expected %s but got %s", id, model.ID)
	}
}
