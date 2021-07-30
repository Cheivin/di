package di

import (
	"fmt"
	"reflect"
)

type (
	// bean定义
	definition struct {
		Name     string
		Type     reflect.Type
		awareMap map[string]aware // fieldName:aware
	}

	// 需要注入的bean信息
	aware struct {
		beanName string
		beanType reflect.Type
		isPtr    bool
	}
)

func newDefinition(beanName string, prototype reflect.Type) definition {
	def := definition{Name: beanName, Type: prototype}
	awareMap := map[string]aware{}
	for i := 0; i < prototype.NumField(); i++ {
		field := prototype.Field(i)
		// 忽略匿名字段
		if field.Anonymous {
			continue
		}
		switch field.Type.Kind() {
		case reflect.Struct, reflect.Ptr:
			if awareName, ok := field.Tag.Lookup("aware"); ok {
				// 取类型名称为注入的beanName
				if awareName == "" {
					awareName = GetBeanName(field.Type)
					fmt.Println("awareName", awareName)
				}
				// 注册aware信息
				awareMap[field.Name] = aware{
					beanName: awareName,
					beanType: field.Type,
					isPtr:    field.Type.Kind() == reflect.Ptr,
				}
			}
		default:
			// ignore其他类型
		}
	}
	def.awareMap = awareMap
	return def
}
