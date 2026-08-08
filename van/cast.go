package van

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func indirect(v any) any {
	value := reflect.Indirect(reflect.ValueOf(v))
	if val, ok := value.Interface().(reflect.Value); ok {
		return val.Interface()
	}
	return value.Interface()
}

func isMap(v any) bool {
	v = indirect(v)
	return reflect.ValueOf(v).Kind() == reflect.Map
}

// toString 将任意值转为字符串。
// 优先 Stringer 接口，其次基础类型，兜底 fmt.Sprint（不再静默丢值）。
func toString(v any) string {
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
	case time.Duration:
		return s.String()
	case fmt.Stringer:
		return s.String()
	default:
		// 整数/无符号统一走 reflect，避免 case 模板
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return strconv.FormatInt(rv.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return strconv.FormatUint(rv.Uint(), 10)
		default:
			return fmt.Sprint(v)
		}
	}
}

var typeDuration = reflect.TypeFor[time.Duration]()

// parseIntN 解析带符号整数，bits 为位宽。
func parseIntN(s string, bits int) (int64, error) {
	return strconv.ParseInt(s, 10, bits)
}

// parseUintN 解析无符号整数，bits 为位宽。
func parseUintN(s string, bits int) (uint64, error) {
	return strconv.ParseUint(s, 10, bits)
}

// Cast 将值转换为目标类型。
// 支持基础类型、time.Duration（纯数字按毫秒兜底）、Stringer、
// 以及 slice/array（元素逐个转换）。
func Cast(v any, typ reflect.Type) (to any, err error) {
	v = indirect(v)

	// slice/array 目标：元素逐个转换
	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		return castSlice(v, typ)
	}

	// 字符串目标：toString 直接转
	if typ.Kind() == reflect.String {
		return toString(v), nil
	}

	// 可直接转换（数值/布尔等类型相等或可转换）
	value := reflect.ValueOf(v)
	if value.IsValid() && value.Type().ConvertibleTo(typ) && typ != typeDuration {
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
		n, e := parseIntN(s, 0)
		if e != nil {
			return nil, e
		}
		to = int(n)
	case reflect.Int64:
		if typ == typeDuration {
			// Duration：优先带单位解析；纯数字按毫秒兜底（历史行为）
			to, err = time.ParseDuration(s)
			if err != nil && strings.HasPrefix(err.Error(), "time: missing unit in duration") {
				to, err = parseIntN(s, 64)
				if err != nil {
					return nil, err
				}
				to = time.Duration(to.(int64)) * time.Millisecond
			}
		} else {
			n, e := parseIntN(s, 64)
			if e != nil {
				return nil, e
			}
			to = n
		}
	case reflect.Int32:
		n, e := parseIntN(s, 32)
		if e != nil {
			return nil, e
		}
		to = int32(n)
	case reflect.Int16:
		n, e := parseIntN(s, 16)
		if e != nil {
			return nil, e
		}
		to = int16(n)
	case reflect.Int8:
		n, e := parseIntN(s, 8)
		if e != nil {
			return nil, e
		}
		to = int8(n)
	case reflect.Uint:
		n, e := parseUintN(s, 0)
		if e != nil {
			return nil, e
		}
		to = uint(n)
	case reflect.Uint64:
		n, e := parseUintN(s, 64)
		if e != nil {
			return nil, e
		}
		to = n
	case reflect.Uint32:
		n, e := parseUintN(s, 32)
		if e != nil {
			return nil, e
		}
		to = uint32(n)
	case reflect.Uint16:
		n, e := parseUintN(s, 16)
		if e != nil {
			return nil, e
		}
		to = uint16(n)
	case reflect.Uint8:
		n, e := parseUintN(s, 8)
		if e != nil {
			return nil, e
		}
		to = uint8(n)
	default:
		return nil, fmt.Errorf("van: unsupported cast to %s", typ.String())
	}
	return
}

// castSlice 将源值转为目标 slice/array 类型。
// 源为 slice/array 时逐个元素转换；源为 string 且目标是 []byte 时直接转。
func castSlice(v any, typ reflect.Type) (any, error) {
	// string → []byte
	if typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Uint8 {
		if s, ok := v.(string); ok {
			return []byte(s), nil
		}
	}
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		return nil, fmt.Errorf("van: cannot cast %T to %s", v, typ.String())
	}
	out := reflect.MakeSlice(typ, rv.Len(), rv.Len())
	elemType := typ.Elem()
	for i := 0; i < rv.Len(); i++ {
		elem, err := Cast(rv.Index(i).Interface(), elemType)
		if err != nil {
			return nil, fmt.Errorf("van: cast slice element %d to %s failed: %w", i, elemType.String(), err)
		}
		out.Index(i).Set(reflect.ValueOf(elem))
	}
	return out.Interface(), nil
}
