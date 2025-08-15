package env_manager

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
)

func getEnv(envMap map[string]string, key string) string {
	if value, ok := envMap[key]; ok {
		return value
	} else {
		return os.Getenv(key)
	}
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

func getNameFromTag(tags []string, fieldName string) string {
	if len(tags) == 0 {
		return ""
	}
	for _, part := range tags {
		if part != "" && !isKeyWord(part) {
			return part
		}
	}

	//fallback to pascale to snake case if no env tag is provided
	fallback := pascaleToSSnakeCase(fieldName)
	return fallback
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

func getEnvValue(prefix, key string) (string, string) {
	if prefix != "" {
		key = prefix + "_" + key
	}
	return key, os.Getenv(key)
}
