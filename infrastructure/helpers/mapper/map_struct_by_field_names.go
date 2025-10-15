package mapper

import (
	"reflect"
)

// MapStructByFieldNames maps fields from source struct to destination struct
// based on matching field names (case sensitive).
// Both src and dst should be pointers to structs.
func MapStructByFieldNames(src interface{}, dst interface{}) error {
	srcValue := reflect.ValueOf(src)
	dstValue := reflect.ValueOf(dst)

	// Ensure we have pointers
	if srcValue.Kind() != reflect.Ptr || dstValue.Kind() != reflect.Ptr {
		panic("both src and dst must be pointers")
	}

	// Get the underlying structs
	srcStruct := srcValue.Elem()
	dstStruct := dstValue.Elem()

	// Ensure we have structs
	if srcStruct.Kind() != reflect.Struct || dstStruct.Kind() != reflect.Struct {
		panic("both src and dst must point to structs")
	}

	srcType := srcStruct.Type()
	dstType := dstStruct.Type()

	// Create a map of destination field names for quick lookup
	dstFields := make(map[string]reflect.Value)
	for i := 0; i < dstStruct.NumField(); i++ {
		field := dstStruct.Field(i)
		fieldName := dstType.Field(i).Name
		if field.CanSet() {
			dstFields[fieldName] = field
		}
	}

	// Iterate through source fields and map to destination
	for i := 0; i < srcStruct.NumField(); i++ {
		srcField := srcStruct.Field(i)
		srcFieldName := srcType.Field(i).Name

		// Check if destination has a field with the same name
		if dstField, exists := dstFields[srcFieldName]; exists {
			// Check if types are compatible
			if srcField.Type().AssignableTo(dstField.Type()) {
				dstField.Set(srcField)
			} else if srcField.Type().ConvertibleTo(dstField.Type()) {
				dstField.Set(srcField.Convert(dstField.Type()))
			}
		}
	}

	return nil
}
