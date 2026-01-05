package utils

import (
	"reflect"
	"strings"
)

func StructToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	// 处理指针
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return result
	}

	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)

		// 跳过未导出字段
		if !fieldValue.CanInterface() {
			continue
		}

		// 读取 json tag
		tag := fieldType.Tag.Get("json")

		// json:"-" 直接跳过
		if tag == "-" {
			continue
		}

		key := tag
		if key == "" {
			// 没有 tag → 用字段名转小驼峰
			key = lowerCamel(fieldType.Name)
		} else {
			// json:"name,omitempty" → 取 name
			key = strings.Split(key, ",")[0]
		}

		result[key] = fieldValue.Interface()
	}

	return result
}
func lowerCamel(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
