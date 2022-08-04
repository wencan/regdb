package regdb

import (
	"fmt"
	"reflect"
)

type registered struct {
	Name  string
	Type  reflect.Type
	Value reflect.Value
}

// RegDB 对象注册和注入主体。
type RegDB struct {
	// registered 所有注册的对象
	registered []*registered
}

// RegisterObject 注册一个对象，注册名称为空。
func (regDB *RegDB) RegisterObject(obj interface{}) {
	regDB.RegisterObjectWithName("", obj)
}

// RegisterObject 注册一个对象，并指定名称。
func (regDB *RegDB) RegisterObjectWithName(name string, obj interface{}) {
	value := reflect.ValueOf(obj)
	item := &registered{Name: name, Type: value.Type(), Value: value}
	regDB.registered = append(regDB.registered, item)
}

// RegisterObjectFields 注册对象下全部带tag标签的字段，tag值为注册名称。
func (regDB *RegDB) RegisterObjectFields(obj interface{}, tagName string) {
	value := reflect.ValueOf(obj)
	for value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}

	typ := value.Type()
	for idx := 0; idx < value.NumField(); idx++ {
		field := typ.Field(idx)
		if field.PkgPath != "" {
			// 未导出的跳过
			continue
		}
		tagValue, tagged := field.Tag.Lookup(tagName)
		if !tagged {
			// 没打tag的跳过
			continue
		}
		fieldValue := value.Field(idx)
		regDB.RegisterObjectWithName(tagValue, fieldValue.Interface())
	}
}

// InjectObject 寻找类型匹配的注册对象，注入到目标对象。
// 如果注入失败，panic。
func (regDB *RegDB) InjectObject(targetObjectAddr interface{}) {
	regDB.InjectObjectByName("", targetObjectAddr)
}

// InjectObjectByName 寻找名称和类型匹配的注册对象，注入到目标对象。
// 如果name为空，将注入找到的第一个名称匹配的注册对象。
// 如果注入失败，panic。
func (regDB *RegDB) InjectObjectByName(name string, targetObjectAddr interface{}) {
	targetValue := reflect.Indirect(reflect.ValueOf(targetObjectAddr))
	targetType := targetValue.Type()

	// 如果对名称有要求，先检查名称匹配的
	if name != "" {
		for _, item := range regDB.registered {
			if name != item.Name {
				continue
			}
			if !item.Type.AssignableTo(targetType) {
				continue
			}
			targetValue.Set(item.Value)
			return
		}
	}

	// 对名称没要求
	for _, item := range regDB.registered {
		if name != "" && name != item.Name {
			continue
		}
		if !item.Type.AssignableTo(targetType) {
			continue
		}
		targetValue.Set(item.Value)
		return
	}

	panic(fmt.Sprintf("not found assignable object by name [%s]", name))
}

// InjectObjectFields 寻找匹配的已注册对象，注入到目标对象的字段。
// 注册的对象 必须能够赋值给注入目标。
// 如果目标对象通过tag指定了注册名称，优先注入类型匹配且名称匹配的注册对象，再选择没有注册名称但类型匹配的注册对象。
// 如果有多个匹配的注册对象，将会注入找到的第一个。
func (regDB *RegDB) InjectObjectFields(targetObjAddr interface{}, tagName string) {
	targetValue := reflect.ValueOf(targetObjAddr)
	for targetValue.Kind() == reflect.Ptr {
		targetValue = reflect.Indirect(targetValue)
	}

	targetType := targetValue.Type()
	for idx := 0; idx < targetValue.NumField(); idx++ {
		field := targetType.Field(idx)
		if field.PkgPath != "" {
			// 未导出的跳过
			continue
		}
		tagValue, tagged := field.Tag.Lookup(tagName)
		if !tagged {
			// 没打tag的跳过
			continue
		}

		fieldValue := targetValue.Field(idx)

		// 如果是结构体对象或者结构体对象的指针
		for fieldValue.Kind() == reflect.Ptr {
			fieldValue = reflect.Indirect(fieldValue)
		}
		if fieldValue.Kind() == reflect.Struct {
			regDB.InjectObjectFields(fieldValue.Addr().Interface(), tagName)
		} else {
			regDB.InjectObjectByName(tagValue, fieldValue.Addr().Interface())
		}
	}
}
