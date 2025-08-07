package env_manager

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	STRUCT_TAG_ENV           = "env"
	STRUCT_TAG_DEFAULT_VALUE = "env_def"
	STRUCT_TAG_DELIMITER     = "env_delim"
	STRUCT_TAG_PREFIX        = "env_prefix"
	STRUCT_TAG_IGNORE        = "env_ignore"
)

// EnvManager is a struct that holds the file name and silent mode
// It is used to manage environment variables from a file
type EnvManager struct {
	File   string
	Silent bool
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
	content := openFile(e.File)
	envMap := envParser(content)

	loadEnvMap(envMap)
}

// Binds a pointer varaible to env varaibles. The assignment is done based on the value provided in
// the field tag 'env'
// example: cat struct{foo string `env:"FOO"`} gets its field foo binded to the varaible 'FOO' 's value
func (e *EnvManager) BindEnv(envStructPtr any) {
	// the varaible provided must be a struct ptr

	varType := reflect.TypeOf(envStructPtr)
	if varType.Kind() != reflect.Pointer || varType.Elem().Kind() != reflect.Struct {
		panic("Invalid type of varaible")
	}

	envStructType := varType.Elem()
	// loop on each field of struct
	for i := range envStructType.NumField() {
		field := envStructType.Field(i)
		envVarName := field.Tag.Get(STRUCT_TAG_ENV)
		if envVarName != "" {
			valStr := os.Getenv(envVarName)
			if valStr == "" {
				if valStr = envStructType.Field(i).Tag.Get(STRUCT_TAG_DEFAULT_VALUE); valStr == "" {
					panic(fmt.Sprintf("error: env variable %s not found", envVarName))
				}
			}
			if value, err := castString(valStr, field.Type, field.Tag.Get(STRUCT_TAG_DELIMITER)); err != nil {
				log.Println(err)
			} else {
				reflect.ValueOf(envStructPtr).Elem().Field(i).Set(value)
			}
		}
	}
}

func castString(value string, target reflect.Type, delim string) (reflect.Value, error) {
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
	case reflect.Array, reflect.Slice:
		// delimter if not provided is ','
		if delim == "" {
			delim = ","
		}
		// split the string by the delimiter and convert each part to the target type
		parts := strings.Split(value, delim)
		slice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(value)), len(parts), len(parts))
		for i, part := range parts {
			part = strings.TrimSpace(part)
			elemValue, elemErr := castString(part, target.Elem(), delim)
			if elemErr != nil {
				return reflect.Value{}, elemErr
			}
			slice.Index(i).Set(elemValue)
		}
		return slice, nil
	case reflect.Chan, reflect.Func:
	// ignored
	default:
		err = fmt.Errorf("error: type %s not supported", target.Kind())
	}
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(castValue), nil
}

func subValues(envMap map[string]string, str string) string {
	start, open, close := 0, 0, 0
	variable := ""
	n := len(str)
	for start < n {
		open = strings.Index(str[start:], "${")
		if open == -1 {
			// no '${' found hence come out of loop
			break
		}
		open += start + 2
		close = strings.Index(str[open:], "}")
		if close == -1 {
			//no '}' found hence come out of loop
			break
		}
		close += open
		variable = str[open:close]
		val := getEnv(envMap, variable)
		if val == "" {
			errMsg := fmt.Sprintf("Error: undefined varaible %s", variable)
			panic(errMsg)
		} else {
			str = str[:open-2] + val + str[close+1:]
		}
		start = close + 1
	}
	return str
}

func getEnv(envMap map[string]string, key string) string {
	if value, ok := envMap[key]; ok {
		return value
	} else {
		return os.Getenv(key)
	}
}

func envParser(file string) map[string]string {
	envMap := make(map[string]string)
	var key, value strings.Builder
	isWithinQuotes := false
	quoteRune := rune(-1)
	isKey := true
	isEnd := false

	for _, line := range strings.SplitAfter(file, "\n") {

		if !isWithinQuotes {
			isWithinQuotes = false
			key.Reset()
			value.Reset()
			isKey = true
			isEnd = false
		}

		for _, ch := range line {
			switch ch {
			case '\'', '"', '`':
				if !isWithinQuotes {
					quoteRune = ch
					isWithinQuotes = true
				} else if quoteRune == ch {
					isWithinQuotes = false
				} else {
					if isKey {
						fmt.Println("Invalid file format, quotes in key")
						panic("Error")
					} else {
						value.WriteRune(ch)
					}
				}
			case '#':
				if isWithinQuotes {
					if isKey {
						key.WriteRune(ch)
					} else {
						value.WriteRune(ch)
					}
				} else {
					isEnd = true
				}
			case '=':
				if !isWithinQuotes {
					if key.String() != "" && isKey {
						isKey = false
					} else {
						fmt.Println("Invalid file format, error")
						panic("Error")
					}
				} else {
					if isKey {
						key.WriteRune(ch)
					} else {
						value.WriteRune(ch)
					}
				}
			case '\n':
				if !isWithinQuotes {
					isEnd = true
				} else {
					if isKey {
						fmt.Println("Invalid file format, error")
						panic("Error")
					} else {
						value.WriteRune('\n')
					}
				}
			default:
				if isKey {
					key.WriteRune(ch)
				} else {
					value.WriteRune(ch)
				}
			}
			if isEnd {
				break
			}
		}
		if !isWithinQuotes && key.String() != "" {
			if quoteRune == '\'' {
				envMap[key.String()] = value.String()
			} else {
				envMap[key.String()] = subValues(envMap, value.String())
			}
		}
	}
	return envMap
}

func openFile(fileName string) string {
	content, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return string(content)
}

func loadEnvMap(envMap map[string]string) {
	for key, value := range envMap {
		if err := os.Setenv(key, value); err != nil {
			panic(fmt.Sprintf("Error setting env variable %s: %v", key, err))
		}
	}

}
