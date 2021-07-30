package di

import (
	"reflect"
	"strings"
)

func IsPtr(o interface{}) bool {
	return reflect.TypeOf(o).Kind() == reflect.Ptr
}

func GetBeanName(o interface{}) (name string) {
	if t, ok := o.(reflect.Type); ok {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name = t.Name()
	} else {
		if IsPtr(o) {
			name = reflect.TypeOf(o).Elem().Name()
		} else {
			name = reflect.TypeOf(o).Name()
		}
	}
	// 简单粗暴将首字母小写
	name = strings.ToLower(name[:1]) + name[1:]
	return
}
