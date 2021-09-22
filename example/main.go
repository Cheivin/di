package main

import (
	"context"
	"fmt"
	"github.com/cheivin/di"
	"log"
	"os/signal"
	"syscall"
	"time"
)

type (
	DB struct {
		Prefix string
	}

	UserDao struct {
		Db               *DB `aware:"db"`
		TableName        string
		DefaultAge       int           `value:"base.user.age"`
		DefaultName      string        `value:"base.user.name"`
		DefaultType      uint8         `value:"base.user.type"`
		DefaultCacheTime time.Duration `value:"base.user.cache"`
		DefaultExpire    time.Duration `value:"base.user.expire"`
	}

	WalletDao struct {
		Db        *DB `aware:"db"`
		TableName string
	}

	OrderRepository interface {
		TableName() string
	}

	OrderDao struct {
		Db *DB `aware:"db"`
	}

	UserService struct {
		UserDao  *UserDao        `aware:""`
		Wallet   *WalletDao      `aware:""`
		OrderDao OrderRepository `aware:"order"`
	}
)

func (o *OrderDao) TableName() string {
	return o.Db.Prefix + "order"
}

func (u UserService) PreInitialize() {
	fmt.Println("依赖注入", "UserService")
}

func (u *UserDao) BeanName() string {
	return "user"
}

func (u *UserDao) AfterPropertiesSet() {
	fmt.Println("装载完成", "UserDao")
	u.TableName = "user"
}

func (w *WalletDao) Initialized() {
	fmt.Println("加载完成", "WalletDao")
	w.TableName = "wallet"
}

func (o *OrderDao) BeanConstruct() {
	fmt.Println("构造实例", "OrderDao")
}

func (u *OrderDao) BeanName() string {
	return "order"
}

func (u *UserService) GetUserTable() string {
	return u.UserDao.Db.Prefix + u.UserDao.TableName
}

func (u *UserService) GetUserDefault() map[string]interface{} {
	return map[string]interface{}{
		"age":    u.UserDao.DefaultAge,
		"name":   u.UserDao.DefaultName,
		"type":   u.UserDao.DefaultType,
		"cache":  u.UserDao.DefaultCacheTime,
		"expire": u.UserDao.DefaultExpire,
	}
}

func (u *UserService) GetWalletTable() string {
	return u.Wallet.Db.Prefix + u.Wallet.TableName
}

func (u *UserService) GetOrderTable() string {
	return u.OrderDao.TableName()
}

func (u *UserService) Destroy() {
	fmt.Println("注销实例", "UserService")
}

func main() {
	di.RegisterNamedBean("db", &DB{Prefix: "test_"}).
		ProvideNamedBean("user", UserDao{}).
		Provide(WalletDao{}).
		Provide(OrderDao{}).
		Provide(UserService{}).
		SetDefaultProperty("base.user.name", "新用户").
		SetProperty("base.user.age", 25).
		SetProperty("base.user.name", "新注册用户").
		SetProperty("base.user.type", "8").
		SetProperty("base.user.cache", "30000").
		SetProperty("base.user.expire", "1h").
		Load()

	bean, ok := di.GetBean("userService")
	if ok {
		log.Println(bean.(*UserService).GetUserTable())
		log.Println(bean.(*UserService).GetWalletTable())
		log.Println(bean.(*UserService).GetOrderTable())
		log.Println(bean.(*UserService).GetUserDefault())
	}
	// 退出信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	di.Serve(ctx)
	fmt.Println("容器退出")
}
