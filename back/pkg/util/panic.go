package util

import (
	"fmt"
	"reflect"
)

func PanicOnNil(val any) {
	if val == nil || reflect.ValueOf(val).IsNil() {
		panic(fmt.Sprintf("value cannot be nil (type: %T)", val))
	}
}
