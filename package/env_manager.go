package env_manager

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	SILENT = iota
	// Not implemented yet
	SEXY
	DEFAULT
)

// EnvManager is a struct that holds the file name and silent mode
// It is used to manage environment variables from a file
type EnvManager struct {
	files   []string
	envMap  map[string]string //contains all the
	logger  *log.Logger
	logMode int
}

func NewEnvManager(files ...string) *EnvManager {
	// if no files are provided then .env is checked
	if len(files) == 0 {
		files = []string{".env"}
	}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			panic(fmt.Sprintf("file %s does not exist", file))
		}
	}

	l := log.Default()
	l.SetFlags(0)

	return &EnvManager{
		envMap: make(map[string]string),
		files:  files,
		logger: l,
	}
}

func (e *EnvManager) SetMode(mode int) *EnvManager {
	switch mode {
	case SILENT:
		e.logger = log.New(io.Discard, "", log.LstdFlags)
	case DEFAULT, SEXY:
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
	loadEnvMap(e.envMap)
}

// Binds a pointer varaible to env varaibles. The assignment is done based on the value provided in
// the field tag 'env'
// example: cat struct{foo string `env:"FOO"`} gets its field foo binded to the varaible 'FOO' 's value
func (e *EnvManager) BindEnv(envStructPtr any) {
	e.logger.Println("Binding environment variable")
	if err := e.bindEnvWithPrefix(envStructPtr, ""); err != nil {
		e.logger.Fatalln(err)
	}
}

func (e *EnvManager) parseEnv() {
	for _, file := range e.files {
		content := openFile(file)
		e.logger.Println("parsing file", file)
		if _, err := envParser(content, e.envMap); err != nil {
			e.logger.Fatalln(err)
		}
	}
}
