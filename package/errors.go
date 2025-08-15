package env_manager

import (
	"fmt"
)

type ErrType int

const (
	KEY_NOT_FOUND_ERROR = iota
	INVALID_USAGE_ERROR
	TYPE_CAST_ERROR
	UNEXPECTED_ERROR
)

const (
	KEY_NOT_FOUND_MSG    = "Key not found"
	INVALID_USAGE_MSG    = "Invalid usage"
	TYPE_CAST_ERROR_MSG  = "Invalid value for the type"
	UNEXPECTED_ERROR_MSG = "Unexpected error"
)

func (err *ErrType) toString() string {
	switch int(*err) {
	case KEY_NOT_FOUND_ERROR:
		return KEY_NOT_FOUND_MSG
	case INVALID_USAGE_ERROR:
		return INVALID_USAGE_MSG
	case TYPE_CAST_ERROR:
		return TYPE_CAST_ERROR_MSG
	default:
		return UNEXPECTED_ERROR_MSG
	}
}

type EnvError struct {
	Type ErrType
	Err  error
}

func (e *EnvError) Error() string {
	return fmt.Sprintf("error occured: %s\n\t%v", e.Type.toString(), e.Err)
}

func NewEnvError(kind ErrType, err error) *EnvError {
	return &EnvError{
		Type: kind,
		Err:  err,
	}
}

func NewKeyNotFoundErr(key string) *EnvError {
	return NewEnvError(
		KEY_NOT_FOUND_ERROR,
		fmt.Errorf("key %s is not in enviroment varaibles", key))
}

func NewInvalidUsageErr(field, use string) *EnvError {
	return NewEnvError(
		INVALID_USAGE_ERROR,
		fmt.Errorf("%s for field %s", field, use))
}

func NewUnSupportedTypeError(field, typeName string) *EnvError {
	return NewInvalidUsageErr(
		field,
		fmt.Sprintf("%s is not supported in field %s", field, typeName))
}

func NewTypeCastErr(value, castType string, err error) *EnvError {
	return NewEnvError(
		TYPE_CAST_ERROR,
		fmt.Errorf("%s cannot be casted to type %s (%v)", value, castType, err))
}

func NewNoKeysForMapErr(field string) *EnvError {
	return NewInvalidUsageErr("empty key in env_keys tag", field)
}
