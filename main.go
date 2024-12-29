package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

// step-1 parse key value pairs
// parser should take care of single quotes, double quotes and ticks

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
	runTests(num)
}
