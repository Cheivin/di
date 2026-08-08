// 示例：生命周期
//
// 演示 bean 的完整生命周期回调顺序，以及 WithContainer 变体。
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

// 生命周期顺序：BeanConstruct → PreInitialize → 注入 → AfterPropertiesSet → Initialized → Destroy(倒序)

type DB struct{}

type Service struct {
	DB *DB `aware:""`
}

// BeanConstruct 实例创建时（注入前）
func (s *Service) BeanConstruct() {
	fmt.Println("[BeanConstruct] 依赖尚未注入，s.DB =", s.DB)
}

// PreInitialize 注入前
func (s *Service) PreInitialize() {
	fmt.Println("[PreInitialize] 即将开始注入")
}

// AfterPropertiesSet 注入完成后
func (s *Service) AfterPropertiesSet() {
	fmt.Printf("[AfterPropertiesSet] 注入完成，s.DB = %T\n", s.DB)
}

// Initialized 所有 bean 加载完毕
func (s *Service) Initialized() {
	fmt.Println("[Initialized] 容器加载完成")
}

// Destroy 容器销毁时（倒序）
func (s *Service) Destroy() {
	fmt.Println("[Destroy] bean 销毁")
}

// 带 WithContainer 的变体：可获取容器引用
type AwareService struct {
	di.DI `aware:""`
}

func (a *AwareService) AfterPropertiesSet(container di.DI) {
	// 在回调里通过容器获取其他 bean
	db, ok := container.GetBean("dB")
	fmt.Printf("[AfterPropertiesSetWithContainer] 通过容器获取 db: %v\n", ok && db != nil)
}

func main() {
	di.RegisterBean(&DB{})
	di.Provide(Service{})
	di.Provide(AwareService{})
	di.Load()

	fmt.Println("\n=== 容器加载完成 ===")
	// 直接调 destroyBeans 演示销毁（Serve 会阻塞，这里跳过）
	// 实际应用中由 di.Serve(ctx) 在收到信号时触发
}
