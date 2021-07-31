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
		Db        DB `aware:"db"`
		TableName string
	}

	WalletDao struct {
		Db DB `aware:"db"`
	}

	UserService struct {
		UserDao *UserDao   `aware:"user"`
		Wallet  *WalletDao `aware`
	}
)

func (u UserService) PreInitialize() {
	fmt.Println("依赖注入", "UserService")
}

func (u *UserDao) AfterPropertiesSet() {
	fmt.Println("装载完成", "UserDao")
	u.TableName = "user"
}

func (w WalletDao) BeanConstruct() {
	fmt.Println("构造实例", "WalletDao")
}

func (u *UserService) GetUserTable() string {
	return u.UserDao.Db.Prefix + u.UserDao.TableName
}

func main() {
	di.RegisterNamedBean("db", &DB{Prefix: "test_"}).
		Provide(WalletDao{}).
		ProvideWithBeanName("user", UserDao{}).
		Provide(UserService{}).
		Load()

	if bean, ok := di.GetBean("userService"); ok {
		log.Println(bean.(*UserService).GetUserTable())
	}
}
