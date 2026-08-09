package di

import (
	"reflect"
	"strings"
)

// IsPtr 判断值是否为指针类型。对 nil（无类型 interface{}）返回 false。
func IsPtr(o any) bool {
	t := reflect.TypeOf(o)
	return t != nil && t.Kind() == reflect.Pointer
}

// GetBeanName 推断 bean 名称：取类型名并将首字母小写。
// 例如 *UserService → "userService"，*DB → "dB"。
// o 可接受实例值或 reflect.Type；传入 nil 会 panic。
func GetBeanName(o any) (name string) {
	if t, ok := o.(reflect.Type); ok {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		name = t.Name()
	} else {
		name = reflect.Indirect(reflect.ValueOf(o)).Type().Name()
	}
	// 简单粗暴将首字母小写
	name = strings.ToLower(name[:1]) + name[1:]
	return
}

// hasPrefix 判断 prefix 是否以 array 中任一字符串为前缀。
// 返回 (是否命中, 命中的前缀串)；array 为空时视为无条件命中。
func hasPrefix(prefix string, array []string) (bool, string) {
	if len(array) == 0 {
		return true, ""
	}
	for i := range array {
		if strings.HasPrefix(prefix, array[i]) {
			return true, array[i]
		}
	}
	return false, ""
}
