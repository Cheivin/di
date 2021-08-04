package van

import (
	"reflect"
	"strconv"
)

func indirect(v interface{}) interface{} {
	value := reflect.Indirect(reflect.ValueOf(v))
	if val, ok := value.Interface().(reflect.Value); ok {
		return val.Interface()
	} else {
		return value.Interface()
	}
}

func isMap(v interface{}) bool {
	v = indirect(v)
	return reflect.ValueOf(v).Kind() == reflect.Map
}

func toString(v interface{}) string {
	v = indirect(v)
	switch s := v.(type) {
	case string:
		return s
	case bool:
		return strconv.FormatBool(s)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32)
	case int:
		return strconv.Itoa(s)
	case int64:
		return strconv.FormatInt(s, 10)
	case int32:
		return strconv.Itoa(int(s))
	case int16:
		return strconv.FormatInt(int64(s), 10)
	case int8:
		return strconv.FormatInt(int64(s), 10)
	case uint:
		return strconv.FormatUint(uint64(s), 10)
	case uint64:
		return strconv.FormatUint(s, 10)
	case uint32:
		return strconv.FormatUint(uint64(s), 10)
	case uint16:
		return strconv.FormatUint(uint64(s), 10)
	case uint8:
		return strconv.FormatUint(uint64(s), 10)
	default:
		return ""
	}
}
