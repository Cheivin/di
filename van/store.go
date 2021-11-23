package van

import (
	"reflect"
	"strings"
)

type store struct {
	separator string
	tree      map[string]interface{}
}

func newStore(separator string) *store {
	return &store{separator: separator, tree: make(map[string]interface{})}
}

func toCaseInsensitiveMap(value interface{}, separator string) map[string]interface{} {
	m := make(map[string]interface{})

	iter := reflect.ValueOf(value).MapRange()
	for iter.Next() {
		key := strings.ToLower(toString(iter.Key()))
		val := iter.Value()
		keyPath := strings.Split(key, separator)
		if len(keyPath) > 1 {
			tmpV := deepSearchIfAbsent(m, keyPath[0:len(keyPath)-1])
			lastKey := keyPath[len(keyPath)-1]
			if isMap(val) {
				tmpV[lastKey] = toCaseInsensitiveMap(val.Interface(), separator)
			} else {
				tmpV[lastKey] = val.Interface()
			}
		} else {
			if isMap(val) {
				m[key] = toCaseInsensitiveMap(val.Interface(), separator)
			} else {
				m[key] = val.Interface()
			}
		}
	}

	return m
}

func copyStringMap(origin map[string]interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(origin))
	iter := reflect.ValueOf(origin).MapRange()
	for iter.Next() {
		key := iter.Key().String()
		if isMap(iter.Value()) {
			m[key] = copyStringMap(iter.Value().Interface().(map[string]interface{}))
		} else {
			m[key] = iter.Value().Interface()
		}
	}
	return m
}

func mergeStringMap(source map[string]interface{}, target map[string]interface{}) {
	for sk, sv := range source {
		tv, ok := target[sk]
		if !ok {
			target[sk] = sv
		} else {
			tvm := isMap(tv)
			svm := isMap(sv)
			if tvm && svm {
				mergeStringMap(sv.(map[string]interface{}), tv.(map[string]interface{}))
			} else if !tvm && !svm {
				target[sk] = sv
			}
		}
	}
}

func deepSearchIfAbsent(tree map[string]interface{}, path []string) map[string]interface{} {
	if len(path) == 0 {
		return tree
	}
	key := path[0]
	subPath := path[1:]
	if sub, ok := tree[key]; !ok {
		// map不存在则创建新map
		emptyTree := make(map[string]interface{})
		tree[key] = emptyTree
		return deepSearchIfAbsent(emptyTree, subPath)
	} else {
		subTree, ok := sub.(map[string]interface{})
		if !ok {
			// 强转失败则用新map代替
			subTree = make(map[string]interface{})
			tree[key] = subTree
		}
		return deepSearchIfAbsent(subTree, subPath)
	}
}

func deepSearch(v interface{}, path []string) interface{} {
	if v == nil || len(path) == 0 {
		return v
	}
	if tree, ok := v.(map[string]interface{}); !ok {
		if len(path) == 1 {
			return v
		}
	} else {
		key := path[0]
		subPath := path[1:]
		return deepSearch(tree[key], subPath)
	}
	return nil
}

func (s *store) Set(key string, value interface{}) {
	key = strings.ToLower(key)
	if isMap(value) {
		value = toCaseInsensitiveMap(value, s.separator)
	}
	keyPath := strings.Split(key, s.separator)
	lastKey := keyPath[len(keyPath)-1]
	tree := deepSearchIfAbsent(s.tree, keyPath[0:len(keyPath)-1])

	if sub, ok := tree[lastKey]; !ok {
		tree[lastKey] = value
	} else {
		if isMap(sub) && isMap(value) {
			mergeStringMap(value.(map[string]interface{}), sub.(map[string]interface{}))
		} else {
			tree[lastKey] = value
		}
	}
}

func (s *store) Get(key string) interface{} {
	key = strings.ToLower(key)
	keyPath := strings.Split(key, s.separator)
	return deepSearch(s.tree, keyPath)
}

func (s *store) GetAll() map[string]interface{} {
	return s.tree
}
