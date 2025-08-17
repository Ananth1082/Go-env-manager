package env_manager

import (
	"fmt"
	"log"
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

// EnvManager is a struct that holds the file name and silent mode
// It is used to manage environment variables from a file
type EnvManager struct {
	File   string
	EnvMap map[string]string
	Logger log.Logger
}

func NewEnvManager(file string) *EnvManager {
	if file == "" {
		panic("file name cannot be empty")
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		panic(fmt.Sprintf("file %s does not exist", file))
	}
	return &EnvManager{
		File:   file,
		Logger: *log.Default(),
	}
}

func (e *EnvManager) GetEnvMap() map[string]string {
	content := openFile(e.File)
	return envParser(content)
}

// Loads the env variables from an env file
// supports use of quotes, double quotes, backticks, and variable substituion
func (e *EnvManager) LoadEnv() {
	e.EnvMap = e.GetEnvMap()
	loadEnvMap(e.EnvMap)
}

// Binds a pointer varaible to env varaibles. The assignment is done based on the value provided in
// the field tag 'env'
// example: cat struct{foo string `env:"FOO"`} gets its field foo binded to the varaible 'FOO' 's value
func (e *EnvManager) BindEnv(envStructPtr any) {
	e.Logger.Println("Binding environment variable")
	if err := e.bindEnvWithPrefix(envStructPtr, ""); err != nil {
		e.Logger.Println(err)
	}
}

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
	e.Logger.Println(field.Name, "type: ", fieldType)
	if fieldType.Kind() == reflect.Map {
		if mapValue, err := e.castMap(field, fieldPrefix); err != nil {
			return err
		} else {
			e.setField(i, field.Name, envStructPtr, mapValue)
		}
		return nil
	} else if checkType(fieldType, "time.Duration") {
		e.Logger.Println("Hello")
		key, valStr := e.getEnvValue(fieldPrefix, envVarName)
		if t, err := time.ParseDuration(valStr); err != nil {
			e.Logger.Println("Error")
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
		for key := range e.EnvMap {
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
	e.Logger.Println("SET", key)
	field.Set(value)
}

func (e *EnvManager) getEnvValue(prefix, key string) (string, string) {
	if prefix != "" {
		key = prefix + "_" + key
	}
	e.Logger.Println("Accessed environment varaible", key)
	return key, os.Getenv(key)
}
