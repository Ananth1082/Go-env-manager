package env_manager

import (
	"fmt"
)

type ErrType int

const (
	KEY_NOT_FOUND_ERROR = iota
	INVALID_USAGE_ERROR
	TYPE_CAST_ERROR
	PARSER_ERROR
	CONFIG_ERROR
	UNEXPECTED_ERROR
)

const (
	KEY_NOT_FOUND_MSG    = "Key not found"
	INVALID_USAGE_MSG    = "Invalid usage"
	TYPE_CAST_ERROR_MSG  = "Invalid value for the type"
	PARSER_ERROR_MSG     = "Invalid env file syntax"
	CONFIG_ERROR_MSG     = "Configuration error"
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
	case CONFIG_ERROR:
		return CONFIG_ERROR_MSG
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

func newEnvError(kind ErrType, err error) *EnvError {
	return &EnvError{
		Type: kind,
		Err:  err,
	}
}

func newConfigError(err error) *EnvError {
	return newEnvError(
		CONFIG_ERROR,
		err)
}

func newKeyNotFoundErr(key string) *EnvError {
	return newEnvError(
		KEY_NOT_FOUND_ERROR,
		fmt.Errorf("key %s is not in enviroment varaibles", key))
}

func newInvalidUsageErr(field, use string) *EnvError {
	return newEnvError(
		INVALID_USAGE_ERROR,
		fmt.Errorf("%s for field %s", field, use))
}

func newUnSupportedTypeError(field, typeName string) *EnvError {
	return newInvalidUsageErr(
		field,
		fmt.Sprintf("%s is not supported in field %s", field, typeName))
}

func newTypeCastErr(value, castType string, err error) *EnvError {
	return newEnvError(
		TYPE_CAST_ERROR,
		fmt.Errorf("%s cannot be casted to type %s (%v)", value, castType, err))
}

func newNoKeysForMapErr(field string) *EnvError {
	return newInvalidUsageErr("empty key in env_keys tag", field)
}

func newParserError(file string, line, ch int, reason string) *EnvError {
	return newEnvError(
		PARSER_ERROR,
		fmt.Errorf("invalid sytax in %s:%d:%d reason: %s", file, line, ch, reason))
}
