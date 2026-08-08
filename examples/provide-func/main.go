// 示例：构造函数注入（ProvideFunc）
//
// 演示通过工厂函数注册 bean，容器按入参类型自动注入依赖。
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

type DB struct {
	Prefix string
}

type Cache struct{}

// UserDao 不使用 aware 标签，依赖通过工厂函数入参注入。
type UserDao struct {
	DB    *DB
	Cache *Cache
}

// newUserDao 是工厂函数：入参类型即依赖类型，返回值作为 bean。
func newUserDao(db *DB, cache *Cache) *UserDao {
	return &UserDao{DB: db, Cache: cache}
}

func main() {
	di.RegisterBean(&DB{Prefix: "tbl_"})
	di.RegisterBean(&Cache{})

	// ProvideFunc 注册工厂函数
	di.ProvideFunc(newUserDao)

	di.Load()

	dao, _ := di.GetBean("userDao")
	u := dao.(*UserDao)
	fmt.Printf("userDao.DB.Prefix = %q\n", u.DB.Prefix)
	fmt.Printf("userDao.Cache injected = %v\n", u.Cache != nil)
}
