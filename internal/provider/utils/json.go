package utils

import (
	"encoding/json"
	"reflect"
)

func IsJSONEqual(a, b string) bool {
	if a == b {
		return true
	}
	var aVal, bVal any
	if err := json.Unmarshal([]byte(a), &aVal); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &bVal); err != nil {
		return false
	}
	return reflect.DeepEqual(aVal, bVal)
}
