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
		Name      string
		Type      reflect.Type
		IsPtr     bool
		Anonymous bool
		Omitempty bool // 不存在依赖时则忽略注入
	}
)

func newDefinition(beanName string, prototype reflect.Type) definition {
	def := definition{Name: beanName, Type: prototype}
	awareMap := map[string]aware{}
	valueMap := map[string]aware{}
	for i := 0; i < prototype.NumField(); i++ {
		field := prototype.Field(i)
		//// 忽略匿名字段
		//if field.Anonymous {
		//	continue
		//}
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
					if awareName == "" {
						// 取接口返回值为注入的beanName
						if tmpBeanName, ok := (reflect.New(field.Type.Elem()).Interface()).(BeanName); ok {
							if name := tmpBeanName.BeanName(); name != "" {
								awareName = name
							}
						}
						if awareName == "" {
							// 取类型名称为注入的beanName
							awareName = GetBeanName(field.Type)
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
						Name:      awareName,
						Type:      field.Type,
						IsPtr:     false,
						Anonymous: field.Anonymous,
						Omitempty: omitempty,
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