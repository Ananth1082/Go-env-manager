package env_manager

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"slices"
	"strings"
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
	Global bool // flag to access all env variables
	Silent bool // show verbose error
}

func NewEnvManager(file string) *EnvManager {
	if file == "" {
		panic("file name cannot be empty")
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		panic(fmt.Sprintf("file %s does not exist", file))
	}
	return &EnvManager{
		File: file,
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
	e.BindEnvWithPrefix(envStructPtr, "")
}

func (e *EnvManager) BindEnvWithPrefix(envStructPtr any, prefix string) {
	// the varaible provided must be a struct ptr
	varType := reflect.TypeOf(envStructPtr)
	if varType.Kind() != reflect.Pointer || varType.Elem().Kind() != reflect.Struct {
		panic("Invalid type of varaible: must be a pointer to a struct")
	}

	envStructType := varType.Elem()
	// loop on each field of struct
	for i := range envStructType.NumField() {
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
			continue
		}

		if field.Type.Kind() == reflect.Map {
			keys := field.Tag.Get(STRUCT_TAG_KEYS)
			if keys == "" {
				panic(fmt.Sprintf("error: map field %s must have env_keys tag", field.Name))
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
					panic(fmt.Sprintf("error: map field %s must have at least one key in env_keys tag", field.Name))
				}
			}

			mapValue := reflect.MakeMap(field.Type)
			for _, key := range keysList {
				if key == "" {
					panic(fmt.Sprintf("error: empty key in env_keys tag for field %s", field.Name))
				}
				key, val := getEnvValue(fieldPrefix, key)
				if val == "" {
					val = field.Tag.Get(STRUCT_TAG_DEFAULT_VALUE)
					if val == "" {
						panic(fmt.Sprintf("error: env variable %s not found", key))
					}
				}
				elemValue, err := castString(val, field.Type.Elem(), delim)
				if err != nil {
					log.Println(err)
					continue
				}
				mapValue.SetMapIndex(reflect.ValueOf(key), elemValue)
			}
			reflect.ValueOf(envStructPtr).Elem().Field(i).Set(mapValue)
			continue
		} else if fieldType.Kind() == reflect.Struct {
			structPtr := reflect.New(fieldType)
			e.BindEnvWithPrefix(structPtr.Interface(), fieldPrefix)
			reflect.ValueOf(envStructPtr).Elem().Field(i).Set(structPtr.Elem())
			continue
		} else if fieldType.Kind() == reflect.Pointer && fieldType.Elem().Kind() == reflect.Struct {
			structPtr := reflect.New(fieldType.Elem())
			e.BindEnvWithPrefix(structPtr.Interface(), fieldPrefix)
			reflect.ValueOf(envStructPtr).Elem().Field(i).Set(structPtr)
			continue
		}

		envVarName := getNameFromTag(envTag, field.Name)

		key, valStr := getEnvValue(fieldPrefix, envVarName)
		if valStr == "" {
			if valStr = envStructType.Field(i).Tag.Get(STRUCT_TAG_DEFAULT_VALUE); valStr == "" {
				if field.Type.Kind() == reflect.Pointer {
					// if the field is a pointer type, we can set it to nil
					reflect.ValueOf(envStructPtr).Elem().Field(i).Set(reflect.Zero(field.Type))
					continue
				} else {
					panic(fmt.Sprintf("error: env variable %s not found", key))
				}
			}
		}
		if isPrimitive(field.Type) {
			if value, err := castStringToPrimitive(valStr, field.Type); err != nil {
				log.Println(err)
			} else {
				reflect.ValueOf(envStructPtr).Elem().Field(i).Set(value)
			}
		} else if field.Type.Kind() == reflect.Slice && isPrimitive(field.Type.Elem()) {
			delim := getDelim(field)
			if value, err := castStringToSlice(valStr, field.Type.Elem(), delim); err != nil {
				log.Println(err)
			} else {
				reflect.ValueOf(envStructPtr).Elem().Field(i).Set(value)
			}
		} else if field.Type.Kind() == reflect.Pointer {
			if value, err := castString(valStr, field.Type.Elem(), ""); err != nil {
				log.Println(err)
			} else {
				ptrValue := reflect.New(field.Type.Elem())
				ptrValue.Elem().Set(value)
				reflect.ValueOf(envStructPtr).Elem().Field(i).Set(ptrValue)
			}
		} else {
			panic(fmt.Sprintf("error: type %s not supported", field.Type.Kind()))
		}
	}
}
