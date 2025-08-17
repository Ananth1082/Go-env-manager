package env_manager

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func castString(value string, target reflect.Type, delim string) (reflect.Value, error) {
	var castValue reflect.Value
	var err error
	if isPrimitiveKind(target) {
		castValue, err = castStringToPrimitive(value, target)
	} else if target.Kind() == reflect.Slice {
		castValue, err = castStringToSlice(value, target.Elem(), delim)
	} else {
		panic(fmt.Sprintf("error: type %s not supported", target.Kind()))
	}
	if err != nil {
		return reflect.Value{}, err
	} else {
		return castValue, nil
	}
}

func castStringToSlice(value string, targetElement reflect.Type, delim string) (reflect.Value, error) {
	parts := strings.Split(value, delim)
	slice := reflect.MakeSlice(reflect.SliceOf(targetElement), len(parts), len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		elemValue, elemErr := castStringToPrimitive(part, targetElement)
		if elemErr != nil {
			return reflect.Value{}, elemErr
		}
		slice.Index(i).Set(elemValue)
	}
	return slice, nil
}

func castStringToPrimitive(value string, target reflect.Type) (reflect.Value, error) {
	var err error
	var castValue any

	switch target.Kind() {
	case reflect.String:
		castValue = value
	case reflect.Int:
		castValue, err = strconv.Atoi(value)
	case reflect.Int8:
		castValue, err = strconv.ParseInt(value, 10, 8)
	case reflect.Int16:
		castValue, err = strconv.ParseInt(value, 10, 16)
	case reflect.Int32:
		castValue, err = strconv.ParseInt(value, 10, 32)
	case reflect.Int64:
		castValue, err = strconv.ParseInt(value, 10, 64)
	case reflect.Uint:
		castValue, err = strconv.ParseUint(value, 10, 0)
	case reflect.Uint8:
		castValue, err = strconv.ParseUint(value, 10, 8)
	case reflect.Uint16:
		castValue, err = strconv.ParseUint(value, 10, 16)
	case reflect.Uint32:
		castValue, err = strconv.ParseUint(value, 10, 32)
	case reflect.Uint64:
		castValue, err = strconv.ParseUint(value, 10, 64)
	case reflect.Uintptr:
		castValue, err = strconv.ParseUint(value, 10, 0)
	case reflect.Complex64:
		castValue, err = strconv.ParseComplex(value, 64)
	case reflect.Complex128:
		castValue, err = strconv.ParseComplex(value, 128)
	case reflect.Bool:
		castValue, err = strconv.ParseBool(value)
	case reflect.Float32:
		castValue, err = strconv.ParseFloat(value, 32)
	case reflect.Float64:
		castValue, err = strconv.ParseFloat(value, 64)
	default:
		err = fmt.Errorf("error: type %s not supported", target.Kind())
	}
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(castValue), nil
}
