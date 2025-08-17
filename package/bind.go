package env_manager

import (
	"os"
	"reflect"
	"slices"
	"strings"
	"time"
)

const (
	STRUCT_TAG_ENV           = "env"
	STRUCT_TAG_DEFAULT_VALUE = "env_def"
	STRUCT_TAG_DELIMITER     = "env_delim"
	STRUCT_TAG_PREFIX        = "env_prefix"
	STRUCT_TAG_KEYS          = "env_keys"
)

const (
	STRUCT_KEYWORD_IGNORE = "ignore"
	STRUCT_KEYWORD_ALL    = "*"
)

func (e *EnvManager) bindEnvWithPrefix(envStructPtr any, prefix string) error {
	// the varaible provided must be a struct ptr
	varType := reflect.TypeOf(envStructPtr)
	if varType.Kind() != reflect.Pointer || varType.Elem().Kind() != reflect.Struct {
		return newInvalidUsageErr("binding varaible", "binding variable must be a pointer to a struct")
	}

	envStructType := varType.Elem()
	// loop on each field of struct
	for i := range envStructType.NumField() {
		if err := e.handleField(envStructPtr, envStructType, i, prefix); err != nil {
			return err
		}
	}
	return nil
}

func (e *EnvManager) handleField(envStructPtr any, envStructType reflect.Type, i int, prefix string) error {
	field := envStructType.Field(i)
	fieldType := field.Type
	envTag := strings.Split(field.Tag.Get(STRUCT_TAG_ENV), ",")

	fieldPrefix := prefix
	if field.Tag.Get(STRUCT_TAG_PREFIX) != "" {
		if fieldPrefix != "" {
			fieldPrefix += "_"
		}
		fieldPrefix += field.Tag.Get(STRUCT_TAG_PREFIX)
	}

	if slices.Contains(envTag, STRUCT_KEYWORD_IGNORE) {
		return nil
	}

	envVarName := getNameFromTag(envTag, field.Name)
	e.logger.Println(field.Name, "type: ", fieldType)
	if fieldType.Kind() == reflect.Map {
		if mapValue, err := e.castMap(field, fieldPrefix); err != nil {
			return err
		} else {
			e.setField(i, field.Name, envStructPtr, mapValue)
		}
		return nil
	} else if checkType(fieldType, "time.Duration") {
		key, valStr := e.getEnvValue(fieldPrefix, envVarName)
		if t, err := time.ParseDuration(valStr); err != nil {
			panic(err)
		} else {
			e.setField(i, key, envStructPtr, reflect.ValueOf(t))
			return nil
		}
	} else if fieldType.Kind() == reflect.Struct {
		structPtr := reflect.New(fieldType)
		if err := e.bindEnvWithPrefix(structPtr.Interface(), fieldPrefix); err != nil {
			return err
		} else {
			e.setField(i, field.Name, envStructPtr, structPtr.Elem())
			return nil
		}
	} else if fieldType.Kind() == reflect.Pointer && fieldType.Elem().Kind() == reflect.Struct {
		structPtr := reflect.New(fieldType.Elem())
		if err := e.bindEnvWithPrefix(structPtr.Interface(), fieldPrefix); err != nil {
			return err
		} else {
			e.setField(i, field.Name, envStructPtr, structPtr)
			return nil
		}
	}

	key, valStr := e.getEnvValue(fieldPrefix, envVarName)
	if valStr == "" {
		if valStr = envStructType.Field(i).Tag.Get(STRUCT_TAG_DEFAULT_VALUE); valStr == "" {
			if fieldType.Kind() == reflect.Pointer {
				// if the field is a pointer type, we can set it to nil
				e.setField(i, field.Name, envStructPtr, reflect.Zero(fieldType))
				return nil
			} else {
				return newKeyNotFoundErr(key)
			}
		}
	}
	if isPrimitiveKind(fieldType) {
		if value, err := castStringToPrimitive(valStr, fieldType); err != nil {
			return newTypeCastErr(valStr, fieldType.Name(), err)
		} else {
			e.setField(i, field.Name, envStructPtr, value)
		}
	} else if fieldType.Kind() == reflect.Slice && isPrimitiveKind(fieldType.Elem()) {
		delim := getDelim(field)
		if value, err := castStringToSlice(valStr, fieldType.Elem(), delim); err != nil {
			return newTypeCastErr(valStr, fieldType.Name(), err)
		} else {
			e.setField(i, field.Name, envStructPtr, value)
		}
	} else if fieldType.Kind() == reflect.Pointer {
		if value, err := castString(valStr, fieldType.Elem(), ""); err != nil {
			return newTypeCastErr(valStr, fieldType.Name(), err)
		} else {
			ptrValue := reflect.New(fieldType.Elem())
			ptrValue.Elem().Set(value)
			e.setField(i, field.Name, envStructPtr, ptrValue)
		}
	} else {
		return newUnSupportedTypeError(field.Name, fieldType.Name())
	}
	return nil
}

func (e *EnvManager) castMap(field reflect.StructField, fieldPrefix string) (reflect.Value, error) {
	emptyValue := reflect.Value{}
	keys := field.Tag.Get(STRUCT_TAG_KEYS)
	if keys == "" {
		return emptyValue, newNoKeysForMapErr(field.Name)
	}

	keysList := []string{}
	delim := getDelim(field)

	if strings.HasSuffix(keys, "*") {
		keyPrefix := strings.TrimSuffix(keys, "*")
		for key := range e.envMap {
			if strings.HasPrefix(key, keyPrefix) {
				keysList = append(keysList, key)
			}
		}
	} else {
		keysList = strings.Split(keys, delim)
		if len(keysList) == 0 {
			return emptyValue, newNoKeysForMapErr(field.Name)
		}
	}

	mapValue := reflect.MakeMap(field.Type)
	for _, key := range keysList {
		if key == "" {
			return emptyValue, newNoKeysForMapErr(field.Name)
		}
		key, val := e.getEnvValue(fieldPrefix, key)
		if val == "" {
			val = field.Tag.Get(STRUCT_TAG_DEFAULT_VALUE)
			if val == "" {
				return emptyValue, newInvalidUsageErr(field.Name, "no default value given")
			}
		}

		if elemValue, err := castString(val, field.Type.Elem(), delim); err != nil {
			return emptyValue, newTypeCastErr(val, field.Type.Name(), err)
		} else {
			mapValue.SetMapIndex(reflect.ValueOf(key), elemValue)
		}
	}
	return mapValue, nil
}

func (e *EnvManager) setField(i int, key string, ptr any, value reflect.Value) {
	field := reflect.ValueOf(ptr).Elem().Field(i)
	e.logger.Println("SET", key)
	field.Set(value)
}

func (e *EnvManager) getEnvValue(prefix, key string) (string, string) {
	if prefix != "" {
		key = prefix + "_" + key
	}
	e.logger.Println("Accessed environment varaible", key)
	return key, os.Getenv(key)
}
