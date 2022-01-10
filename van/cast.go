package van

import (
	"reflect"
	"strconv"
	"strings"
	"time"
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
	case time.Duration:
		return s.String()
	default:
		return ""
	}
}

var typeDuration = reflect.TypeOf(time.Nanosecond)

func Cast(v interface{}, typ reflect.Type) (to interface{}, err error) {
	v = indirect(v)
	if typ.Kind() == reflect.String {
		return toString(v), nil
	}
	value := reflect.ValueOf(v)
	if value.Type().ConvertibleTo(typ) && typ != typeDuration {
		return value.Convert(typ).Interface(), nil
	}
	s := toString(v)
	switch typ.Kind() {
	case reflect.Bool:
		to, err = strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
	case reflect.Float64:
		to, err = strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
	case reflect.Float32:
		to, err = strconv.ParseFloat(s, 32)
		if err != nil {
			return nil, err
		}
		to = float32(to.(float64))
	case reflect.Int:
		to, err = strconv.ParseInt(s, 10, 0)
		if err != nil {
			return nil, err
		}
		to = int(to.(int64))
	case reflect.Int64:
		if typ == typeDuration {
			to, err = time.ParseDuration(s)
			if err != nil && strings.HasPrefix(err.Error(), "time: missing unit in duration") {
				to, err = strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				to = time.Duration(to.(int64)) * time.Millisecond
			}
		} else {
			to, err = strconv.ParseInt(s, 10, 64)
			if err != nil {
				return nil, err
			}
		}
	case reflect.Int32:
		to, err = strconv.ParseInt(s, 10, 32)
		if err != nil {
			return nil, err
		}
		to = int32(to.(int64))
	case reflect.Int16:
		to, err = strconv.ParseInt(s, 10, 16)
		if err != nil {
			return nil, err
		}
		to = int16(to.(int64))
	case reflect.Int8:
		to, err = strconv.ParseInt(s, 10, 8)
		if err != nil {
			return nil, err
		}
		to = int8(to.(int64))
	case reflect.Uint:
		to, err = strconv.ParseUint(s, 10, 0)
		if err != nil {
			return nil, err
		}
		to = uint(to.(uint64))
	case reflect.Uint64:
		to, err = strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
	case reflect.Uint32:
		to, err = strconv.ParseUint(s, 10, 32)
		if err != nil {
			return nil, err
		}
		to = uint32(to.(uint64))
	case reflect.Uint16:
		to, err = strconv.ParseUint(s, 10, 16)
		if err != nil {
			return nil, err
		}
		to = uint16(to.(uint64))
	case reflect.Uint8:
		to, err = strconv.ParseUint(s, 10, 8)
		if err != nil {
			return nil, err
		}
		to = uint8(to.(uint64))
	default:
	}
	return
}
