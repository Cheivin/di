package di

import (
	"reflect"
	"sort"
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

func in(target string, array []string) bool {
	sort.Strings(array)
	index := sort.SearchStrings(array, target)
	if index < len(array) && array[index] == target {
		return true
	}
	return false
}

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
