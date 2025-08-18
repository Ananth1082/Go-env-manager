package env_manager

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	MAX_SUB_DEPTH = 10
)

type envParser struct {
	file    string
	content string
	env     map[string]string
	visited map[string]bool
}

func newEnvParser(file string, env map[string]string) (*envParser, error) {
	content, err := openFile(file)
	if err != nil {
		return nil, err
	}
	p := &envParser{
		file:    file,
		content: content,
		visited: make(map[string]bool),
	}
	if env == nil {
		p.env = make(map[string]string)
	} else {
		p.env = env
	}
	return p, nil
}

func (e *envParser) setEnv(key, value string) {
	e.env[strings.TrimSpace(key)] = strings.TrimSpace(value)
}

// TODO: create a better method to check if string contains substitution for the cached value
func (e *envParser) subValues(str string, depth int) (string, error) {
	if depth > MAX_SUB_DEPTH {
		return "", fmt.Errorf("maximum substitution depth %d exceeded", MAX_SUB_DEPTH)
	}
	start := 0
	for {
		open := strings.Index(str[start:], "${")
		if open == -1 {
			break
		}
		open += start + 2
		close := strings.Index(str[open:], "}")
		if close == -1 {
			break
		}
		close += open
		varName := str[open:close]

		if e.visited[varName] {
			return "", newConfigError(fmt.Errorf("circular reference detected for variable %s", varName))
		}

		if cached, ok := e.getEnv(varName); ok && !strings.Contains(cached, "${") {
			str = str[:open-2] + cached + str[close+1:]
			start = open - 2 + len(cached)
			continue
		}

		e.visited[varName] = true
		val, ok := e.getEnv(varName)
		if !ok {
			return "", newConfigError(fmt.Errorf("variable %s not found", varName))
		}

		subVal, err := e.subValues(val, depth+1)
		if err != nil {
			return "", err
		}

		e.setEnv(varName, subVal)

		str = str[:open-2] + subVal + str[close+1:]
		delete(e.visited, varName)

		start = open - 2 + len(subVal)
	}
	return str, nil
}

func (e *envParser) parse() error {
	var key, value strings.Builder
	isWithinQuotes := false
	isQuoteEnd := false
	quoteRune := rune(-1)
	isKey := true
	isEnd := false

	for lineNum, line := range strings.SplitAfter(e.content, "\n") {

		if !isWithinQuotes {
			isWithinQuotes = false
			isQuoteEnd = false
			key.Reset()
			value.Reset()
			isKey = true
			isEnd = false
		}

		for chNum, ch := range line {
			switch ch {
			case '\'', '"', '`':
				if !isWithinQuotes {
					quoteRune = ch
					isWithinQuotes = true
				} else if quoteRune == ch {
					isWithinQuotes = false
					isQuoteEnd = true
				} else {
					if isKey {
						return newParserError(e.file, lineNum+1, chNum+1, "No quotes allowed in key")
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
						return newParserError(e.file, lineNum+1, chNum+1, "Keys cannot be empty")
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
						return newParserError(e.file, lineNum+1, chNum+1, "Keys cannot have new line, expected '='")
					} else {
						value.WriteRune('\n')
					}
				}
			default:
				if isKey {
					key.WriteRune(ch)
				} else if isQuoteEnd && !unicode.IsSpace(ch) {
					return newParserError(e.file, lineNum+1, chNum+1, "Only white space charecters or comments allowed after end of quote")
				} else {
					value.WriteRune(ch)
				}
			}
			if isEnd {
				break
			}
		}
		if !isWithinQuotes && key.String() != "" {
			e.setEnv(key.String(), value.String())
		}
	}

	// substitute all variable values
	for k, v := range e.env {
		if subValue, err := e.subValues(v, 0); err != nil {
			return err
		} else {
			e.env[k] = subValue
		}
	}
	return nil
}
