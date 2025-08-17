package env_manager

import (
	"fmt"
	"log"
	"os"
)

// EnvManager is a struct that holds the file name and silent mode
// It is used to manage environment variables from a file
type EnvManager struct {
	Files  []string
	EnvMap map[string]string
	Logger log.Logger
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
	return &EnvManager{
		Files:  files,
		Logger: *log.Default(),
	}
}

func (e *EnvManager) GetEnvMap() map[string]string {
	envMap := make(map[string]string)
	for _, file := range e.Files {
		content := openFile(file)
		envParser(content, envMap)
	}
	return envMap
}

// Loads the env variables from an env file
// supports use of quotes, double quotes, backticks, and variable substituion
func (e *EnvManager) LoadEnv() {
	e.EnvMap = e.GetEnvMap()
	loadEnvMap(e.EnvMap)
}

// Binds a pointer varaible to env varaibles. The assignment is done based on the value provided in
// the field tag 'env'
// example: cat struct{foo string `env:"FOO"`} gets its field foo binded to the varaible 'FOO' 's value
func (e *EnvManager) BindEnv(envStructPtr any) {
	e.Logger.Println("Binding environment variable")
	if err := e.bindEnvWithPrefix(envStructPtr, ""); err != nil {
		e.Logger.Println(err)
	}
}
