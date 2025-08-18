package env_manager

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	DEBUG = iota + 1
	DEFAULT
	SILENT
)

// EnvManager is a struct that holds the file name and silent mode
// It is used to manage environment variables from a file
type EnvManager struct {
	files   []string
	envMap  map[string]string //contains all the
	logger  *log.Logger
	logMode int
}

func NewEnvManager(files ...string) (*EnvManager, error) {
	// if no files are provided then .env is checked
	if len(files) == 0 {
		files = []string{".env"}
	}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return nil, newConfigError(fmt.Errorf("file %s does not exist", file))
		}
	}

	l := log.Default()
	l.SetFlags(0)

	return &EnvManager{
		envMap:  make(map[string]string),
		files:   files,
		logger:  l,
		logMode: DEFAULT,
	}, nil
}

func (e *EnvManager) SetMode(mode int) *EnvManager {
	switch mode {
	case SILENT:
		e.logger = log.New(io.Discard, "", log.LstdFlags)
	case DEFAULT, DEBUG:
		e.logMode = mode
	default:
		e.logMode = DEFAULT
	}
	return e
}

func (e *EnvManager) SetLogger(l *log.Logger) *EnvManager {
	e.logger = l
	return e
}

func (e *EnvManager) GetEnvMap() map[string]string {
	e.parseEnv()
	return e.envMap
}

// Loads the env variables from an env file
// supports use of quotes, double quotes, backticks, and variable substituion
func (e *EnvManager) LoadEnv() {
	e.parseEnv()
	if err := loadEnvMap(e.envMap); err != nil {
		e.Log(HIGH, "Error loading environment variables: %v", err)
	}
}

// Binds a pointer varaible to env varaibles. The assignment is done based on the value provided in
// the field tag 'env'
// example: cat struct{foo string `env:"FOO"`} gets its field foo binded to the varaible 'FOO' 's value
func (e *EnvManager) BindEnv(envStructPtr any) {
	e.Log(MED, "Binding environment variables")
	if err := e.bindEnvWithPrefix(envStructPtr, ""); err != nil {
		e.Log(HIGH, "Error binding environment variable: %v", err)
	}
}

func (e *EnvManager) parseEnv() {
	for _, file := range e.files {
		if parser, err := newEnvParser(file, e.envMap); err == nil {
			if err := parser.parse(); err != nil {
				e.Log(HIGH, "Error parsing env file %s: %v", file, err)
			}
		} else {
			e.Log(HIGH, "Error creating env parser for file %s: %v", file, err)
		}
	}
}
