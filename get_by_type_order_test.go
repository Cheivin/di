package di

import (
	"testing"
)

type orderDep interface{ Id() int }

type orderImplA struct{}

func (orderImplA) Id() int { return 1 }

type orderImplB struct{}

func (orderImplB) Id() int { return 2 }

type orderImplC struct{}

func (orderImplC) Id() int { return 3 }

// TestGetByTypeAllOrder 验证 GetByTypeAll 按注册顺序返回（接口契约：非 map 随机序）。
func TestGetByTypeAllOrder(t *testing.T) {
	container := New()
	// 乱序注册：C、A、B
	container.Provide(orderImplC{})
	container.Provide(orderImplA{})
	container.Provide(orderImplB{})
	container.Load()

	beans := container.GetByTypeAll((*orderDep)(nil))
	if len(beans) != 3 {
		t.Fatalf("GetByTypeAll = %d beans, want 3", len(beans))
	}
	wantIds := []int{3, 1, 2} // 按注册顺序：C、A、B
	for i, want := range wantIds {
		if got := beans[i].Bean.(orderDep).Id(); got != want {
			t.Fatalf("GetByTypeAll order[%d] = %d, want %d (registration order)", i, got, want)
		}
	}
}

// TestGetByTypeOrder 验证 GetByType 返回第一个注册的匹配 bean。
func TestGetByTypeOrder(t *testing.T) {
	container := New()
	container.Provide(orderImplB{})
	container.Provide(orderImplA{})
	container.Load()

	bean, ok := container.GetByType((*orderDep)(nil))
	if !ok {
		t.Fatal("GetByType should find orderDep")
	}
	if got := bean.(orderDep).Id(); got != 2 {
		t.Fatalf("GetByType = %d, want 2 (first registered)", got)
	}
}
