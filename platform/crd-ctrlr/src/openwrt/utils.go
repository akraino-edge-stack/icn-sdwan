package openwrt

import (
	"reflect"
)

// util function to check whether items contains item
func IsContained(items interface{}, item interface{}) bool {
	switch reflect.TypeOf(items).Kind() {
	case reflect.Slice:
		v := reflect.ValueOf(items)
		for i := 0; i < v.Len(); i++ {
			if reflect.DeepEqual(item, v.Index(i).Interface()) {
				return true
			}
		}
	default:
		return false
	}

	return false
}
