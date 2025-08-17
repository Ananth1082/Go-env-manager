package env_manager

import (
	"fmt"
	"strings"
)

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

func envParser(file string, envMap map[string]string) map[string]string {
	if envMap == nil {
		envMap = make(map[string]string)
	}
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
