package main

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// step-1 parse key value pairs
// parser should take care of single quotes

func envParser(file string) map[string]string {
	envVars := make(map[string]string)
	var key, value strings.Builder

	for _, line := range strings.Split(file, "\n") {
		key.Reset()
		value.Reset()
		isKey := true
		isEnd := false
		isWithinQuotes := false

		for i, ch := range line {
			switch ch {
			case '\'':
				if isWithinQuotes {
					isWithinQuotes = false
				} else {
					isWithinQuotes = true
				}
			case '#':
				if !isWithinQuotes {
					isEnd = true
				} else {
					if isKey {
						key.WriteRune(ch)
					} else {
						value.WriteRune(ch)
					}
				}
			case '=':
				if !isWithinQuotes {
					if key.String() != "" && isKey {
						isKey = false
					} else {
						fmt.Println("Invalid file format, error in line ", i)
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
						key.WriteRune(ch)
					} else {
						value.WriteRune(ch)
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
		if key.String() != "" {
			envVars[key.String()] = value.String()
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
func runTests(testNum int) {
	dir := "test"
	for i := range testNum {
		file := fmt.Sprintf(".env_%d", i+1)
		envFile := openFile(path.Join(dir, file))
		fmt.Printf("Test %d results: %s\n", i+1, envParser(envFile))
	}
}

func main() {
	runTests(1)
}
