package env_manager

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Loads the env variables from an env file
// supports use of quotes, double quotes, backticks, and variable substituion
func LoadEnv(fileName string) {
	cnt := openFile(fileName)
	envParser(cnt)
}

// Binds a pointer varaible to env varaibles. The assignment is done based on the value provided in
// the field tag 'env'
// example: cat struct{foo string `env:"FOO"`} gets its field foo binded to the varaible 'FOO' 's value
func BindEnv(envStructPtr interface{}) {
	// the varaible provided must be a struct ptr

	varType := reflect.TypeOf(envStructPtr)
	if varType.Kind() != reflect.Pointer || varType.Elem().Kind() != reflect.Struct {
		panic("Invalid type of varaible")
	}

	envStruct := varType.Elem()
	// loop on each field of struct
	for i := range envStruct.NumField() {
		field := envStruct.Field(i)
		envVarName := field.Tag.Get("env")
		if envVarName != "" {
			valStr := os.Getenv(envVarName)
			if valStr != "" {
				switch envStruct.Field(i).Type.Kind() {
				case reflect.String:
					reflect.ValueOf(envStructPtr).Elem().Field(i).Set(reflect.ValueOf(valStr))
				case reflect.Int:
					valInt, err := strconv.Atoi(valStr)
					if err != nil {
						errMsg := fmt.Sprintf("error: env variable %s is not a integer", envVarName)
						panic(errMsg)
					}
					reflect.ValueOf(envStructPtr).Elem().Field(i).Set(reflect.ValueOf(valInt))
				}
			} else {
				errMsg := fmt.Sprintf("error: env variable %s not found", envVarName)
				panic(errMsg)
			}
		}
	}
}

func subValues(str string) string {
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
		val := os.Getenv(variable)
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

func envParser(file string) {
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
				os.Setenv(key.String(), value.String())
			} else {
				os.Setenv(key.String(), subValues(value.String()))
			}
		}
	}
}

func openFile(fileName string) string {
	content, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return string(content)
}
