package env_manager

import (
	"slices"
	"testing"
	"time"
)

// Testing the parsing logic for simple env file
func TestParsingForSimpleEnvFile(t *testing.T) {
	emap := make(map[string]string)

	if err := newEnvParser("../test_data/simple.env", emap).parse(); err != nil {
		t.Error(err)
	}

	assertEqual(t, len(emap), 6, "Invalid number of env variables parsed")
	assertEqual(t, emap["APP_NAME"], "MyCoolApp", "Invalid App name from env")
	assertEqual(t, emap["APP_PORT"], "8080", "Invalid port from env")
	t.Log("Env variables: ", len(emap))
}

// Testing parsing logic for complex features like multi-line strings, varaibles
func TestParsingForComplexFile(t *testing.T) {
	envManger := NewEnvManager("../test_data/complex.env")
	envMap := envManger.GetEnvMap()

	assertEqual(t, len(envMap), 25, "Invalid number of env variables parsed")
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

type TestBindEnvForDefaultKeyNamesStruct struct {
	APPName string
	APPEnv  string
	APPPort int
}

func TestBindEnvForDefaultKeyNames(t *testing.T) {
	envBinder := new(TestBindEnvForDefaultKeyNamesStruct)
	envManager := NewEnvManager("../test_data/simple.env")
	envManager.LoadEnv()

	envManager.BindEnv(envBinder)

	assertEqual(t, envBinder.APPName, "MyCoolApp", "APPName must access APP_NAME")
	assertEqual(t, envBinder.APPEnv, "production", "APPEnv must access APP_ENV")
	assertEqual(t, envBinder.APPPort, 8080, "APPPort must access APP_PORT")
}

type TestBindEnvForDefaultValueStruct struct {
	AppName    string
	AppEnv     string
	AppVersion string `env_def:"v0.0"`
	AppSeed    int    `env_def:"69"`
}

func TestBindEnvForDefaultValue(t *testing.T) {
	envBinder := new(TestBindEnvForDefaultValueStruct)
	envManager := NewEnvManager("../test_data/simple.env")
	envManager.LoadEnv()

	envManager.BindEnv(envBinder)

	assertEqual(t, envBinder.AppSeed, 69, "AppSeed should be defaulted to 69")
	assertEqual(t, envBinder.AppVersion, "v0.0", "AppVersion should be defauled to v0.0")
}

type TestNilPointerFieldStruct struct {
	AppName string
}

type TestBindEnvForComplexDataStruct struct {
	IgnoreField int `env:"ignore"`

	AppName  *string       `env:"APP_NAME"`
	Version  string        `env:"VERSION"`
	Options  []string      `env:"OPTIONS" env_delim:";"`
	Colors   []string      `env:"COLORS"`
	AppCount int           `env:"APP_COUNT" env_def:"69"`
	Expiry   time.Duration `env:"EXPIRY"`

	Email struct {
		Host      string
		Port      int
		User      string
		Pass      string
		Signature string
	} `env_prefix:"EMAIL"`
	TLS *struct {
		TLSCert string
		TLSKey  string
	}

	EnvKeys  map[string]string `env_keys:"*" env_delim:","`
	MetaKeys map[string]string `env_keys:"META_*" env_delim:","`
	AppKeys  map[string]string `env_keys:"APP_NAME,VERSION,OPTIONS"`
}

func TestBindEnvForComplexData(t *testing.T) {
	envBinder := new(TestBindEnvForComplexDataStruct)

	envManger := NewEnvManager("../test_data/complex.env")
	envManger.LoadEnv()

	envManger.BindEnv(envBinder)
	assertEqual(t, *envBinder.AppName, "MultiLineApp", "Invalid AppName")
	assertEqual(t, envBinder.Version, "1.0.0", "Invalid Version")
	assertCondition(t, slices.Equal(envBinder.Options, []string{"min", "med", "max"}), "Invalid Options")
	assertEqual(t, envBinder.AppCount, 69, "Invalid AppCount")
	t.Log(envBinder)
	t.Log(*envBinder.TLS)
}
