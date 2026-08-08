package van

import (
	"encoding/json"
	"reflect"
)

func toStringMap(v any) map[string]any {
	v = indirect(v)
	value := reflect.ValueOf(v)
	var m = map[string]any{}
	switch value.Kind() {
	case reflect.String:
		_ = json.Unmarshal([]byte(value.String()), &m)
	case reflect.Map:
		iter := value.MapRange()
		for iter.Next() {
			m[toString(iter.Key())] = iter.Value().Interface()
		}
	default:
		return m
	}
	return m
}
