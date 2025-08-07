package env_manager

import (
	"fmt"
	"log"
	"maps"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"
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
	e.EnvMap = e.GetEnvMap()
	loadEnvMap(e.EnvMap)
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
		envTag := strings.Split(field.Tag.Get(STRUCT_TAG_ENV), ",")

		if slices.Contains(envTag, STRUCT_KEYWORD_IGNORE) {
			continue
		}
		//TODO: fix the map implementation
		if field.Type.Kind() == reflect.Map {
			keys := field.Tag.Get(STRUCT_TAG_KEYS)
			if keys == "" {
				panic(fmt.Sprintf("error: map field %s must have env_keys tag", field.Name))
			}
			var keysList []string
			delim := getDelim(field)
			if keys == STRUCT_KEYWORD_ALL {
				keysList = slices.Collect(maps.Keys(e.EnvMap))
			} else {
				keysList = strings.Split(keys, delim)
				if len(keysList) == 0 {
					panic(fmt.Sprintf("error: map field %s must have at least one key in env_keys tag", field.Name))
				}
			}

			mapValue := reflect.MakeMap(field.Type)
			fmt.Println("keysList: ", keysList)
			for _, key := range keysList {
				key = strings.TrimSpace(key)
				if key == "" {
					panic(fmt.Sprintf("error: empty key in env_keys tag for field %s", field.Name))
				}
				val := os.Getenv(key)
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
		}

		envVarName := getNameFromTag(envTag)

		//fallback to pascale to snake case if no env tag is provided
		if envVarName == "" {
			envVarName = pascaleToSSnakeCase(field.Name)
		}

		valStr := os.Getenv(envVarName)
		if valStr == "" {
			if valStr = envStructType.Field(i).Tag.Get(STRUCT_TAG_DEFAULT_VALUE); valStr == "" {
				panic(fmt.Sprintf("error: env variable %s not found", envVarName))
			}
		}
		if isPrimitive(field.Type) {
			if value, err := castStringToPrimitive(valStr, field.Type); err != nil {
				log.Println(err)
			} else {
				reflect.ValueOf(envStructPtr).Elem().Field(i).Set(value)
			}
		} else if field.Type.Kind() == reflect.Slice {
			delim := getDelim(field)
			if value, err := castStringToSlice(valStr, field.Type.Elem(), delim); err != nil {
				log.Println(err)
			} else {
				reflect.ValueOf(envStructPtr).Elem().Field(i).Set(value)
			}
		} else {
			panic(fmt.Sprintf("error: type %s not supported", field.Type.Kind()))
		}
	}
}

func castString(value string, target reflect.Type, delim string) (reflect.Value, error) {
	var castValue reflect.Value
	var err error
	if isPrimitive(target) {
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

func getNameFromTag(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	for _, part := range tags {
		if part != "" && !isKeyWord(part) {
			return part
		}
	}
	return ""
}

func isKeyWord(tagEntry string) bool {
	switch tagEntry {
	case STRUCT_KEYWORD_IGNORE:
		return true
	default:
		return false
	}
}

func pascaleToSSnakeCase(str string) string {
	var result strings.Builder
	for i, ch := range str {
		if unicode.IsUpper(ch) && i > 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToUpper(ch))
	}
	return result.String()
}

// IsPrimitive checks whether the type is a Go primitive type.
func isPrimitive(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}

func getDelim(field reflect.StructField) string {
	delim := field.Tag.Get(STRUCT_TAG_DELIMITER)
	if delim == "" {
		delim = ","
	}
	return delim
}
