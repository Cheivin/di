package van

import (
	"reflect"
	"strings"
)

type store struct {
	separator string
	tree      map[string]any
}

func newStore(separator string) *store {
	return &store{separator: separator, tree: make(map[string]any)}
}

func toCaseInsensitiveMap(value any, separator string) map[string]any {
	m := make(map[string]any)

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

func copyStringMap(origin map[string]any) map[string]any {
	m := make(map[string]any, len(origin))
	for key, value := range origin {
		if sub, ok := value.(map[string]any); ok {
			m[key] = copyStringMap(sub)
		} else {
			m[key] = value
		}
	}
	return m
}

func mergeStringMap(source map[string]any, target map[string]any) {
	for sk, sv := range source {
		tv, ok := target[sk]
		if !ok {
			target[sk] = sv
			continue
		}
		tvm, tIsMap := tv.(map[string]any)
		svm, sIsMap := sv.(map[string]any)
		if tIsMap && sIsMap {
			mergeStringMap(svm, tvm)
		} else {
			// 类型冲突（如 string vs map）：新值覆盖旧值（原为静默丢弃）
			target[sk] = sv
		}
	}
}

func deepSearchIfAbsent(tree map[string]any, path []string) map[string]any {
	if len(path) == 0 {
		return tree
	}
	key := path[0]
	subPath := path[1:]
	if sub, ok := tree[key]; !ok {
		// map不存在则创建新map
		emptyTree := make(map[string]any)
		tree[key] = emptyTree
		return deepSearchIfAbsent(emptyTree, subPath)
	} else {
		subTree, ok := sub.(map[string]any)
		if !ok {
			// 强转失败则用新map代替
			subTree = make(map[string]any)
			tree[key] = subTree
		}
		return deepSearchIfAbsent(subTree, subPath)
	}
}

func deepSearch(v any, path []string) any {
	if v == nil || len(path) == 0 {
		return v
	}
	if tree, ok := v.(map[string]any); !ok {
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

func (s *store) Set(key string, value any) {
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
			mergeStringMap(value.(map[string]any), sub.(map[string]any))
		} else {
			tree[lastKey] = value
		}
	}
}

func (s *store) Get(key string) any {
	key = strings.ToLower(key)
	keyPath := strings.Split(key, s.separator)
	return deepSearch(s.tree, keyPath)
}

func (s *store) GetAll() map[string]any {
	return s.tree
}
