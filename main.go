package main

import (
	"fmt"
	"os"
	"strings"
)

// step-1 parse key value pairs

func openFile(fileName string) string {
	content, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return string(content)
}

func envParser(file string) map[string]string {
	envVars := make(map[string]string)
	var key, value strings.Builder
	isKey := true
	isEnd := false
	for _, line := range strings.Split(file, "\n") {
		for i, ch := range line {
			switch ch {
			case '#':
				isEnd = true
			case '=':
				if key.String() != "" && isKey {
					isKey = false
				} else {
					fmt.Println("Invalid file format, error in line ", i)
					fmt.Println("State: ", envVars)
					panic("Error")
				}
			case '\n':
				isEnd = true
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
		key.Reset()
		value.Reset()
		isKey = true
		isEnd = false
	}
	return envVars
}

func main() {
	cnt := openFile("test/.env")
	fmt.Println(envParser(cnt))
}
