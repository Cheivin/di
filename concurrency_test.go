package di

import (
	"sync"
	"testing"
)

// 并发测试用的 bean 类型
type concDB struct {
	Name string
}

type concService struct {
	DB *concDB `aware:""`
}

func (concDB) BeanName() string { return "concDB" }
func (*concDB) BeanConstruct()  {}

// TestConcurrent_GetBean Load 后多 goroutine 并发读，-race 下不应报错
func TestConcurrent_GetBean(t *testing.T) {
	c := New()
	c.RegisterBean(&concDB{Name: "read"})
	c.Load()

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			bean, ok := c.GetBean("concDB")
			if !ok {
				t.Error("expected concDB found")
			}
			if bean == nil {
				t.Error("expected non-nil bean")
			}
		})
	}
	wg.Wait()
}

// TestConcurrent_GetByType 并发按类型查询（range beanMap 路径）
func TestConcurrent_GetByType(t *testing.T) {
	c := New()
	c.RegisterBean(&concDB{Name: "type"})
	c.Load()

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			_, ok := c.GetByType(&concDB{})
			if !ok {
				t.Error("expected GetByType ok")
			}
		})
	}
	wg.Wait()
}

// TestConcurrent_RegisterAfterLoad Load 后并发动态注册 + 并发读
// 这是运行期 race 的核心场景：getAllByType 的 range 与 RegisterNamedBean 的写并发
func TestConcurrent_RegisterAfterLoad(t *testing.T) {
	c := New()
	c.RegisterBean(&concDB{Name: "init"})
	c.Load()

	var wg sync.WaitGroup
	// 并发写：动态注册新 bean
	for i := range 20 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := "dyn" + itoa(n)
			c.RegisterNamedBean(name, &concDB{Name: name})
		}(i)
	}
	// 并发读：range beanMap
	for range 20 {
		wg.Go(func() {
			_ = c.GetByTypeAll(&concDB{})
		})
	}
	// 并发读：按名查
	for range 20 {
		wg.Go(func() {
			_, _ = c.GetBean("init")
		})
	}
	wg.Wait()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
