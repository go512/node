package utils

import (
	"reflect"
	"strings"
	"unicode"
)

// 定义一个函数类型，允许自定义决定是否包含某个字段
type ShouldIncludeFunc func(value reflect.Value, tag reflect.StructTag) bool

func StructToUpdateMap(obj any, shouldInclude ...ShouldIncludeFunc) map[string]any {
	result := make(map[string]any)
	v := reflect.ValueOf(obj)
	t := v.Type()

	//处理指针，如果是指针，则获取指针指向的值
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	//如果不是结构体，则直接返回空map
	if v.Kind() != reflect.Struct {
		return result
	}

	//确定用于判断是否包含字段的函数
	var includeFunc ShouldIncludeFunc
	if len(shouldInclude) > 0 && shouldInclude[0] != nil {
		includeFunc = shouldInclude[0]
	} else {
		//默认包含非零值的字段
		includeFunc = func(value reflect.Value, tag reflect.StructTag) bool {
			return !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
		}
	}

	//遍历所有字段
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)
		fieldTag := structField.Tag

		//判断是否应包含此字段
		if !includeFunc(field, fieldTag) {
			continue
		}

		//获取该字段的列名
		columnName := getColumnName(structField)
		if columnName == "-" {
			continue
		}

		if columnName != "" {
			result[columnName] = field.Interface()
		}
	}
	return result
}

func getColumnName(field reflect.StructField) string {
	//1.解析 gorm 标签
	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		//简单解析 column 值，适用于大部分简单场景
		for _, part := range strings.Split(gormTag, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}

			// 如果标签是 `gorm:"-"`，直接返回忽略标识
			if part == "-" {
				return ""
			}
		}
	}

	//如果没有指定column 标签， 则将字段名转换为snake_case
	// 这里使用一个简单的转换函数，更严谨的实现可参考 GORM 源码
	return toSnakeCase(field.Name)
}

func toSnakeCase(s string) string {
	var output []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				output = append(output, '_')
			}
			output = append(output, unicode.ToLower(r))
		} else {
			output = append(output, r)
		}
	}
	return string(output)
}
