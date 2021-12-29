---
description: 每个由容器创建的bean都有其生命周期
---

# 生命周期

### 实例构造BeanConstruct

* 相关接口方法
  * BeanConstruct
  * BeanConstruct(DI)，DI为当前容器
* 在容器执行`Load()`后，会按以下逻辑执行：
  1. 所有注册的结构体，按注册的先后顺序反射创建对象。
  2. 注入由`value`标记的配置项。
  3. 触发该生命周期，构造bean。
* 该生命周期只对由容器实例化且实现相关接口的bean有效。

### 实例初始化PreInitialize

* 相关接口方法
  * PreInitialize()
  * PreInitialize(DI)，DI为当前容器
* 在所有由容器实例化的bean构造完成后，按构造的先后顺序，每个bean开始注入相关依赖。
* 在每个bean的依赖注入前触发该生命周期，此时通过容器可以获得其他bean的指针对象，但获取到的bean不一定完成了依赖注入，可根据代码的配置顺序确定。
* 该生命周期只对由容器实例化且实现相关接口的bean有效。

### 依赖注入完成AfterPropertiesSet

* 相关接口方法
  * AfterPropertiesSet()
  * AfterPropertiesSet(DI)，DI为当前容器
* 在每个bean的依赖注入完成时触发该生命周期，此时通过容器可以获得其他bean的指针对象，但获取到的bean不一定完成了依赖注入，可根据代码的配置顺序确定。
* 该生命周期只对由容器实例化且实现相关接口的bean有效。

### 容器加载完成

* 相关接口方法
  * Initialized()
  * Initialized(DI)，DI为当前容器
* 在所有bean完成依赖注入后，按注册的先后顺序，触发该生命周期。
* 该生命周期对所有实现相关接口的bean有效。

### 容器销毁

* 相关接口方法
  * Destroy()
  * Destroy(DI)，DI为当前容器
* 在容器销毁时，按注册顺序的逆序，触发该生命周期。
* 如果是手动实例化的bean，则在对象被GC回收时触发该生命周期。
* 该生命周期对所有实现相关接口的bean有效。
