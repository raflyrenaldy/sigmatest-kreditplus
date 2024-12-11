package reflecty

import (
	"reflect"
)

func GetTagByFieldName[T, F any](st T, field F, tag string) string {
	fieldName := GetFieldName(st, field)
	if fieldName == "" {
		return ""
	}

	result, found := reflect.TypeOf(st).Elem().FieldByName(fieldName)
	if !found {
		return ""
	}

	return result.Tag.Get(tag)
}

func GetFieldName[T, F any](st T, field F) string {
	if reflect.ValueOf(field).Kind() != reflect.Ptr {
		return ""
	}

	structType := reflect.ValueOf(st).Elem()
	fieldType := reflect.ValueOf(field).Elem()

	for i := 0; i < structType.NumField(); i++ {
		if structType.Field(i).Addr().Interface() == fieldType.Addr().Interface() {
			return structType.Type().Field(i).Name
		}
	}

	return ""
}
