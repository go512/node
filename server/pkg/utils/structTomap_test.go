package utils

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type User struct {
	ID        int       `gorm:"primaryKey;column:id" json:"id"`
	UserName  string    `gorm:"column:user_name" json:"user_name"`
	Password  string    `gorm:"column:password" json:"password"`
	Age       int       `gorm:"column:age" json:"age"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func TestName(t *testing.T) {
	user := &User{
		ID:       1,
		UserName: "test",
		Password: "123456",
		Age:      0,
	}

	includeFunc := func(value reflect.Value, tag reflect.StructTag) bool {
		if tag.Get("gorm") == "column:password" || tag.Get("gorm") == "column:created_at" {
			return false
		}

		return !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
	}

	arr := StructToUpdateMap(user, includeFunc)
	fmt.Println(arr)
}

func TestIncludeFunc(t *testing.T) {
	var oldUser User
	//db.First(&oldUser, 123)
	oldUser = User{UserName: "李四", Age: 20}

	input := User{UserName: "张三", Age: 20, CreatedAt: time.Now()}

	//自定义函数： 比较新旧值,仅在新值不同非0值时才更新
	// 注意：此函数是一个闭包，捕获了oldUser的值
	createIncludeIfChangeFunc := func(old any) ShouldIncludeFunc {
		oldVal := reflect.ValueOf(old)
		if oldVal.Kind() == reflect.Ptr {
			oldVal = oldVal.Elem()
		}

		return func(value reflect.Value, tag reflect.StructTag) bool {
			if reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				return false
			}
			fmt.Println("abc", value.Interface(), value.Field(0).Interface())
			if oldVal.Interface() == value.Interface() {
				return false
			}
			return true
		}
	}

	updateMap := StructToUpdateMap(input, createIncludeIfChangeFunc(oldUser))
	fmt.Println(updateMap)
}
