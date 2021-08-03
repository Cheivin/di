package main

import (
	"fmt"
	"github.com/cheivin/di"
	"log"
)

type (
	DB struct {
		Prefix string
	}

	UserDao struct {
		Db        *DB `aware:"db"`
		TableName string
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
		UserDao  *UserDao        `aware:"user"`
		Wallet   *WalletDao      `aware:""`
		OrderDao OrderRepository `aware:"orderDao"`
	}
)

func (o *OrderDao) TableName() string {
	return o.Db.Prefix + "order"
}

func (u UserService) PreInitialize() {
	fmt.Println("依赖注入", "UserService")
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

func (u *UserService) GetUserTable() string {
	return u.UserDao.Db.Prefix + u.UserDao.TableName
}

func (u *UserService) GetWalletTable() string {
	return u.Wallet.Db.Prefix + u.Wallet.TableName
}

func (u *UserService) GetOrderTable() string {
	return u.OrderDao.TableName()
}

func main() {
	di.RegisterNamedBean("db", &DB{Prefix: "test_"}).
		ProvideWithBeanName("user", UserDao{}).
		Provide(WalletDao{}).
		Provide(OrderDao{}).
		Provide(UserService{}).
		Load()

	bean, ok := di.GetBean("userService")
	if ok {
		log.Println(bean.(*UserService).GetUserTable())
		log.Println(bean.(*UserService).GetWalletTable())
		log.Println(bean.(*UserService).GetOrderTable())
	}
}
