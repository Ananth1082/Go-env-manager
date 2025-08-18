package env_manager

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
)

func (e *envParser) getEnv(key string) (string, bool) {
	if value, ok := e.env[key]; ok {
		return value, true
	} else {
		return os.LookupEnv(key)
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
	fallback := pascalToSnakeCase(fieldName)
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

// Converts PascalCase or camelCase to SNAKE_CASE
// Keeps abbreviations intact: TLSCert -> TLS_CERT
func pascalToSnakeCase(str string) string {
	var result strings.Builder
	runes := []rune(str)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			// Add underscore if not first character and previous rune is lower
			if i > 0 && (unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
				result.WriteRune('_')
			}
		}
		result.WriteRune(unicode.ToUpper(r))
	}
	return result.String()
}

// IsPrimitive checks whether the type is a Go primitive type.
func isPrimitiveKind(t reflect.Type) bool {
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

func checkType(typ reflect.Type, fullTypeName string) bool {
	return typ.PkgPath()+"."+typ.Name() == fullTypeName
}

func getDefaultValue(field reflect.StructField) *string {
	value, exists := field.Tag.Lookup(STRUCT_TAG_DEFAULT_VALUE)
	if exists {
		return &value
	} else {
		return nil
	}
}
