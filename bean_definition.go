package di

import (
	"fmt"
	"reflect"
	"strings"
)

type (
	// bean定义
	definition struct {
		Name        string
		Type        reflect.Type
		awareMap    map[string]aware // fieldName:aware
		valueMap    map[string]aware // fieldName:aware
		factory     reflect.Value    // 工厂函数；非零值表示工厂模式，按入参类型注入
		factoryArgs []reflect.Type   // 工厂入参类型列表
	}

	// 需要注入的信息
	aware struct {
		Name        string
		Type        reflect.Type
		IsPtr       bool // 是否为结构指针
		IsInterface bool // 是否为接口
		Anonymous   bool // 是否为匿名字段
		Omitempty   bool // 不存在依赖时则忽略注入
		IsSlice     bool // 是否为 slice，收集所有可赋值给 ElemType 的 bean
		IsMap       bool // 是否为 map[string]T，以 beanName 为 key 收集
		ElemType    reflect.Type // slice/map 的元素类型
	}
)

func (container *di) newDefinition(beanName string, prototype reflect.Type) definition {
	def := definition{Name: beanName, Type: prototype}
	awareMap := map[string]aware{}
	valueMap := map[string]aware{}
	for i := 0; i < prototype.NumField(); i++ {
		field := prototype.Field(i)
		switch field.Type.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Struct:
			if awareName, ok := field.Tag.Lookup("aware"); ok {
				omitempty := false
				switch {
				case strings.EqualFold(awareName, "omitempty"):
					omitempty = true
					awareName = ""
				case strings.HasSuffix(awareName, ",omitempty"):
					omitempty = true
					awareName = strings.TrimSuffix(awareName, ",omitempty")
				}

				switch field.Type.Kind() {
				case reflect.Ptr:
					if reflect.Interface == field.Type.Elem().Kind() {
						panic(fmt.Errorf("%w: aware bean not accept interface pointer for %s.%s", ErrDefinition, prototype.String(), field.Name))
					}
					tmpBean := reflect.New(field.Type.Elem()).Interface()
					if awareName == "" {
						switch tmpBean.(type) {
						case BeanName: // 取接口返回值为注入的beanName
							if name := tmpBean.(BeanName).BeanName(); name != "" {
								awareName = name
							}
						}
					}
					if awareName == "" {
						// 取类型名称为注入的beanName
						awareName = GetBeanName(field.Type)
					}
					// 检查匿名类
					if field.Anonymous {
						errInterface := checkAnonymousFieldBean(tmpBean)
						if errInterface != "" {
							container.log.Fatal(fmt.Errorf("%w: %s(%s) as anonymous field in %s(%s.%s) can not implements %s",
								ErrBean, awareName, field.Type.String(),
								def.Name, def.Type.String(), field.Name,
								errInterface,
							))
						}
					}

					// 注册aware信息
					awareMap[field.Name] = aware{
						Name:      awareName,
						Type:      field.Type,
						IsPtr:     true,
						Anonymous: field.Anonymous,
						Omitempty: omitempty,
					}
				case reflect.Interface:
					// 取类型名称为注入的beanName
					if awareName == "" {
						awareName = GetBeanName(field.Type)
					}
					// 注册aware信息
					awareMap[field.Name] = aware{
						Name:        awareName,
						Type:        field.Type,
						IsPtr:       false,
						IsInterface: true,
						Anonymous:   field.Anonymous,
						Omitempty:   omitempty,
					}
			case reflect.Struct:
				panic(fmt.Errorf("%w: aware bean not accept struct for %s.%s", ErrDefinition, prototype.String(), field.Name))
			}
		}
	case reflect.Slice, reflect.Map:
		if awareName, ok := field.Tag.Lookup("aware"); ok {
			// 解析元素类型
			elemType := field.Type.Elem()
			// map 必须是 map[string]T
			isMap := field.Type.Kind() == reflect.Map
			if isMap && field.Type.Key().Kind() != reflect.String {
				panic(fmt.Errorf("%w: aware map key must be string for %s.%s", ErrDefinition, prototype.String(), field.Name))
			}
			// omitempty 解析（slice/map 的 awareName 不影响行为，靠类型收集）
			omitempty := false
			switch {
			case strings.EqualFold(awareName, "omitempty"):
				omitempty = true
			case strings.HasSuffix(awareName, ",omitempty"):
				omitempty = true
			}
			awareMap[field.Name] = aware{
				Type:      field.Type,
				ElemType:  elemType,
				IsSlice:   !isMap,
				IsMap:     isMap,
				Omitempty: omitempty,
			}
		}
	case reflect.String, reflect.Bool,
			reflect.Float64, reflect.Float32,
			reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			if property, ok := field.Tag.Lookup("value"); ok {
				if property != "" {
					valueMap[field.Name] = aware{
						Name: property,
						Type: field.Type,
					}
				}
			}
		default:
			// ignore其他类型
		}
	}
	def.awareMap = awareMap
	def.valueMap = valueMap
	return def
}

func (container *di) getValueDefinition(prototype reflect.Type) definition {
	def := definition{Name: prototype.Name(), Type: prototype}
	valueMap := map[string]aware{}
	for i := 0; i < prototype.NumField(); i++ {
		field := prototype.Field(i)
		switch field.Type.Kind() {
		case reflect.String, reflect.Bool,
			reflect.Float64, reflect.Float32,
			reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
			reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			if property, ok := field.Tag.Lookup("value"); ok {
				if property != "" {
					valueMap[field.Name] = aware{
						Name: property,
						Type: field.Type,
					}
				}
			}
		default:
			// ignore其他类型
		}
	}
	def.valueMap = valueMap
	return def
}

// 匿名结构体字段不能实现的生命周期接口（实现会导致方法被意外提升）
var anonymousForbiddenInterfaces = []reflect.Type{
	reflect.TypeOf((*BeanConstruct)(nil)).Elem(),
	reflect.TypeOf((*BeanConstructWithContainer)(nil)).Elem(),
	reflect.TypeOf((*PreInitialize)(nil)).Elem(),
	reflect.TypeOf((*PreInitializeWithContainer)(nil)).Elem(),
	reflect.TypeOf((*AfterPropertiesSet)(nil)).Elem(),
	reflect.TypeOf((*AfterPropertiesSetWithContainer)(nil)).Elem(),
	reflect.TypeOf((*Initialized)(nil)).Elem(),
	reflect.TypeOf((*InitializedWithContainer)(nil)).Elem(),
	reflect.TypeOf((*Disposable)(nil)).Elem(),
	reflect.TypeOf((*DisposableWithContainer)(nil)).Elem(),
}

// checkAnonymousFieldBean 检查匿名字段不能实现的接口
func checkAnonymousFieldBean(awareBean any) string {
	// 匿名字段不能实现BeanConstruct/PreInitialize/AfterPropertiesSet/Initialized/Disposable等生命周期接口
	for _, iface := range anonymousForbiddenInterfaces {
		if reflect.TypeOf(awareBean).Implements(iface) {
			return iface.Name()
		}
	}
	return ""
}
