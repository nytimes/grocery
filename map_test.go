package grocery

import (
	"encoding/json"
	"testing"
)

func TestMap(t *testing.T) {
	mapData := map[string]string{"a": "b", "c": "d"}
	m := newMap(mapData)

	m.Map.Range(func(key, value interface{}) bool {
		loadedValue, ok := mapData[key.(string)]

		if !ok {
			t.Errorf("map value FAILED, key %s does not exist", key)
			return true
		}

		if value != loadedValue {
			t.Errorf("map value FAILED, expected m[%s]=%s but got m[%s]=%s", key, value, key, loadedValue)
		}

		return true
	})

	if len(mapData) != m.Count() {
		t.Errorf("map count FAILED, expected %d but got %d", len(mapData), m.Count())
	}

	expectedJSON, _ := json.Marshal(mapData)
	json, _ := json.Marshal(m)

	if string(json) != string(expectedJSON) {
		t.Errorf("marshaljson FAILED, expected %s but got %s", string(expectedJSON), string(json))
	}
}
