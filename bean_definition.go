package di

import (
	"fmt"
	"reflect"
	"strings"
)

type (
	// bean定义
	definition struct {
		Name     string
		Type     reflect.Type
		awareMap map[string]aware // fieldName:aware
		valueMap map[string]aware // fieldName:aware
	}

	// 需要注入的信息
	aware struct {
		Name        string
		Type        reflect.Type
		IsPtr       bool // 是否为结构指针
		IsInterface bool // 是否为接口
		Anonymous   bool // 是否为匿名字段
		Omitempty   bool // 不存在依赖时则忽略注入
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
							container.log.Fatal(fmt.Sprintf("%s: %s(%s) as anonymous field in %s(%s.%s) can not implements %s",
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

// checkAnonymousFieldBean 检查匿名字段不能实现的接口
func checkAnonymousFieldBean(awareBean interface{}) string {
	// 匿名字段不能实现BeanConstruct/PreInitialize/AfterPropertiesSet/Initialized/Disposable等生命周期接口
	switch awareBean.(type) {
	case BeanConstruct:
		return "BeanConstruct"
	case BeanConstructWithContainer:
		return "BeanConstructWithContainer"
	case PreInitialize:
		return "PreInitialize"
	case PreInitializeWithContainer:
		return "PreInitializeWithContainer"
	case AfterPropertiesSet:
		return "AfterPropertiesSet"
	case AfterPropertiesSetWithContainer:
		return "AfterPropertiesSetWithContainer"
	case Initialized:
		return "Initialized"
	case InitializedWithContainer:
		return "InitializedWithContainer"
	case Disposable:
		return "Disposable"
	case DisposableWithContainer:
		return "DisposableWithContainer"
	default:
		return ""
	}
}
