package van

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_toCaseInsensitiveMap(t *testing.T) {
	m := toCaseInsensitiveMap(map[any]any{
		"A": 1,
		"B": "2",
		"c": map[string]int{
			"c1":  1,
			"c2":  2,
			"c3":  3,
			"d.e": 5,
		},
		"c.d.f": map[string]any{
			"x": 6,
			"y": 7,
		},
		"m": map[string]any{
			"d.f": map[string]any{
				"x": 9,
				"y": 8,
			},
		},
	}, ".")
	b, _ := json.MarshalIndent(m, "", "\t")
	t.Log(string(b))
}

func Test_copyStringMap(t *testing.T) {
	a := map[string]any{
		"a": 1,
		"b": 2,
		"c": 3,
		"m1": map[string]any{
			"x": 1,
			"y": 2,
		},
	}
	b := copyStringMap(a)
	fmt.Println(b)
	b["d"] = 4
	bm1 := b["m1"].(map[string]any)
	bm1["z"] = "z"
	fmt.Println(b)
	fmt.Println(a)
}

func Test_mergeStringMap(t *testing.T) {
	a := map[string]any{
		"a": 1,
		"b": 2,
		"c": 3,
		"m1": map[string]any{
			"x":  1,
			"y":  2,
			"z1": 4,
		},
		"m2": map[string]any{
			"x":  1,
			"y":  2,
			"z1": 4,
		},
		"y": "zxc",
	}
	b := map[string]any{
		"a": 2,
		"b": 3,
		"c": 4,
		"m1": map[string]any{
			"x":  1,
			"y":  4,
			"z3": 5,
		},
		"x": "asd",
	}
	target := copyStringMap(b)
	mergeStringMap(a, target)
	fmt.Println(target)
	fmt.Println(b)
}

func TestVan_Set(t *testing.T) {
	v := New()
	v.Set("A", 1)
	v.Set("a.b.c", 3)
	v.Set("b.c", map[string]int{
		"x": 1,
		"y": 2,
	})
	v.Set("b.c.d", "test")

	t.Log(v)

	t.Log("GET")
	t.Log("a", v.Get("a"))
	t.Log("a.b.c", v.Get("a.b.c").(int))
	t.Log("b.c", v.Get("b.c").(map[string]any))
	t.Log("b.c.x", v.Get("b.c.x"))
	t.Log("b.c.d", v.Get("b.c.d").(string))
	t.Log("b.c.e", v.Get("b.c.e"))
	t.Log("b.d.e", v.Get("b.d.e"))
}
