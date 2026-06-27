package resolver

import (
	"reflect"
	"strings"
	"unicode"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

func convertFilterStructs[T any](filters []*T) []entity.FilterNode {
	if len(filters) == 0 {
		return nil
	}
	var nodes []entity.FilterNode
	for _, filter := range filters {
		if filter == nil {
			continue
		}
		nodes = append(nodes, convertSingleFilter(reflect.ValueOf(filter).Elem())...)
	}
	if len(nodes) == 0 {
		return nil
	}
	return nodes
}

func convertSingleFilter(filterVal reflect.Value) []entity.FilterNode {
	filterType := filterVal.Type()
	var nodes []entity.FilterNode

	for idx := 0; idx < filterType.NumField(); idx++ {
		fieldInfo := filterType.Field(idx)
		fieldVal := filterVal.Field(idx)

		if fieldVal.IsNil() {
			continue
		}

		fieldName := fieldInfo.Name
		switch fieldName {
		case "And":
			children := convertFilterSlice(fieldVal)
			if len(children) > 0 {
				nodes = append(nodes, entity.FilterNode{And: children})
			}
		case "Or":
			children := convertFilterSlice(fieldVal)
			if len(children) > 0 {
				nodes = append(nodes, entity.FilterNode{Or: children})
			}
		case "Not":
			notVal := fieldVal.Elem()
			children := convertSingleFilter(notVal)
			for _, child := range children {
				childCopy := child
				nodes = append(nodes, entity.FilterNode{Not: &childCopy})
			}
		default:
			fieldFilter := extractOperators(pascalToCamelCase(fieldName), fieldVal.Elem())
			nodes = append(nodes, fieldFilter...)
		}
	}

	return nodes
}

func convertFilterSlice(sliceVal reflect.Value) []entity.FilterNode {
	var nodes []entity.FilterNode
	for idx := 0; idx < sliceVal.Len(); idx++ {
		elem := sliceVal.Index(idx)
		if elem.IsNil() {
			continue
		}
		nodes = append(nodes, convertSingleFilter(elem.Elem())...)
	}
	return nodes
}

func extractOperators(fieldName string, operatorStruct reflect.Value) []entity.FilterNode {
	opType := operatorStruct.Type()
	var nodes []entity.FilterNode

	for idx := 0; idx < opType.NumField(); idx++ {
		opField := opType.Field(idx)
		opVal := operatorStruct.Field(idx)

		if opVal.IsNil() {
			continue
		}

		var operator entity.FilterOperator
		switch opField.Name {
		case "Eq":
			operator = entity.FilterOpEq
		case "Ne":
			operator = entity.FilterOpNe
		case "In":
			operator = entity.FilterOpIn
		case "NotIn":
			operator = entity.FilterOpNotIn
		default:
			continue
		}

		var value interface{}
		if operator == entity.FilterOpIn || operator == entity.FilterOpNotIn {
			sliceVal := opVal
			if sliceVal.Kind() == reflect.Ptr {
				sliceVal = sliceVal.Elem()
			}
			arr := make([]interface{}, sliceVal.Len())
			for arrIdx := 0; arrIdx < sliceVal.Len(); arrIdx++ {
				arr[arrIdx] = sliceVal.Index(arrIdx).Interface()
			}
			value = arr
		} else {
			if opVal.Kind() == reflect.Ptr {
				value = opVal.Elem().Interface()
			} else {
				value = opVal.Interface()
			}
		}

		nodes = append(nodes, entity.FilterNode{
			Field: &entity.FieldFilter{
				Field:    fieldName,
				Operator: operator,
				Value:    value,
			},
		})
	}

	return nodes
}

func extractOrderBy(orderByArg interface{}) (string, int) {
	if orderByArg == nil {
		return "createdAt", -1
	}

	val := reflect.ValueOf(orderByArg)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return "createdAt", -1
		}
		val = val.Elem()
	}

	valType := val.Type()
	for idx := 0; idx < valType.NumField(); idx++ {
		fieldVal := val.Field(idx)
		if fieldVal.IsNil() {
			continue
		}

		fieldName := pascalToCamelCase(valType.Field(idx).Name)
		sortOrder := fieldVal.Elem().String()

		direction := -1
		if strings.EqualFold(sortOrder, "ASC") {
			direction = 1
		}

		return fieldName, direction
	}

	return "createdAt", -1
}

func pascalToCamelCase(name string) string {
	if len(name) == 0 {
		return name
	}
	runes := []rune(name)
	for idx := 0; idx < len(runes); idx++ {
		if !unicode.IsUpper(runes[idx]) {
			break
		}
		if idx+1 < len(runes) && unicode.IsLower(runes[idx+1]) {
			runes[idx] = unicode.ToLower(runes[idx])
			break
		}
		runes[idx] = unicode.ToLower(runes[idx])
	}
	return string(runes)
}
