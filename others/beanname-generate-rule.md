# beanName生成策略

## **B**eanName接口

`BeanName`接口在注册实例、结构体和依赖注入字段时触发，用于生成目标的beanName。

接口定义如下：

```
type BeanName interface {
	// BeanName 返回beanName
	BeanName() string
}
```

### 生成规则

* 如果实例实现了`BeanName`接口，则使用接口方法返回值作为beanName。
* 否则将根据结构体名称，把第一个字母转换小写，得到beanName。
  * eg：
    * `AService` => `aService`
    * `DB` => `DB`
    * `api` => `api`
