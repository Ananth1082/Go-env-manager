package main

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

type ENVMap map[string]string

var EnvVars ENVMap = ENVMap{"user": "admin", "password": "admin123"}

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
			reflect.ValueOf(envStructPtr).Elem().Field(i).Set(reflect.ValueOf(EnvVars[envVarName]))
		} else {
			errMsg := fmt.Sprintf("error: env variable %s not found", envVarName)
			panic(errMsg)
		}
	}
}

func subValues(str string, env ENVMap) string {
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
		val := env[variable]
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

func envParser(file string) map[string]string {
	envVars := make(map[string]string)
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
						fmt.Println("State: ", envVars)
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
						fmt.Println("State: ", envVars)
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
						fmt.Println("State: ", envVars)
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
				envVars[key.String()] = value.String()
			} else {
				envVars[key.String()] = subValues(value.String(), envVars)
			}
		}
	}
	return envVars
}

func openFile(fileName string) string {
	content, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return string(content)
}

func tostring(envmap map[string]string) {
	for key, val := range envmap {
		fmt.Printf("\t%s ...... %s\n", key, val)
	}
}

func runTests(testNum int) {
	dir := "test"
	for i := range testNum {
		file := fmt.Sprintf(".env_%d", i+1)
		envFile := openFile(path.Join(dir, file))
		fmt.Printf("***************************************\nTest %d results:\n", i+1)
		tostring(envParser(envFile))
		fmt.Print("**************************************\n\n")
	}
}

func main() {
	numStr := os.Args[1]
	num, _ := strconv.Atoi(numStr)
	fmt.Println("num: ", num)
	// runTests(num)
	a := new(struct {
		User     string `env:"user"`
		Password string `env:"password"`
	})
	BindEnv(a)
	fmt.Println(a)
}
