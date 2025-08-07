package env_manager

import (
	"slices"
	"testing"
)

func assertEqual[T comparable](t *testing.T, actual, expected T, msg string) {
	if actual != expected {
		t.Errorf("Assertion error\nExpected %v got %v\n%s", expected, actual, msg)
	}
}

func assertCondition(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Errorf("Assertion error\nCondition failed: %s", msg)
	}
}

// Testing the parsing logic for simple env file
func TestGetEnvMapForSimpleFile(t *testing.T) {
	envManger := NewEnvManager("../test_data/simple.env")
	envMap := envManger.GetEnvMap()

	assertEqual(t, len(envMap), 6, "Invalid number of env variables parsed")
	assertEqual(t, envMap["APP_NAME"], "MyCoolApp", "Invalid App name from env")
	assertEqual(t, envMap["APP_PORT"], "8080", "Invalid port from env")
	t.Log("Env variables: ", len(envMap))
}

// Testing parsing logic for complex features like multi-line strings, varaibles
func TestGetEnvMapForComplexFile(t *testing.T) {
	envManger := NewEnvManager("../test_data/complex.env")
	envMap := envManger.GetEnvMap()

	assertEqual(t, len(envMap), 22, "Invalid number of env variables parsed")
	assertEqual(t, envMap["APP_NAME"], "MultiLineApp", "Invalid value for variable APP_NAME from env")

	assertEqual(t, envMap["WELCOME_MESSAGE"], `Welcome to $APP_NAME!
Environment: production
Running at http://127.0.0.1:8080`, "Invalid value for variable WELCOME_MESSAGE from env")

	assertEqual(t, envMap["INIT_SQL"], `CREATE TABLE users (
id SERIAL PRIMARY KEY,
name TEXT NOT NULL,
email TEXT UNIQUE NOT NULL
);`, "Invalid value for variable INIT_SQL from env")

	t.Log("Env variables: ", len(envMap))
}

type EnvData struct {
	AppName  string            `env:"APP_NAME"`
	Version  string            `env:"VERSION"`
	Options  []string          `env:"OPTIONS" env_delim:","`
	AppCount int               `env:"APP_COUNT" env_def:"69"`
	EnvKeys  map[string]string `env_keys:"APP_NAME,VERSION,OPTIONS" env_delim:","`
}

func TestBindEnvForSimpleStruct(t *testing.T) {

	envBinder := new(EnvData)

	envManger := NewEnvManager("../test_data/complex.env")
	envManger.LoadEnv()

	envManger.BindEnv(envBinder)
	assertEqual(t, envBinder.AppName, "MultiLineApp", "Invalid AppName")
	assertEqual(t, envBinder.Version, "1.0.0", "Invalid Version")
	assertCondition(t, slices.Equal(envBinder.Options, []string{"min", "med", "max"}), "Invalid Options")
	assertEqual(t, envBinder.AppCount, 69, "Invalid AppCount")
	t.Log(envBinder)

}
