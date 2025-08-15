# Go Env Manager

Is a simple, light weight package to manage your env varaibles in a go project. It uses reflection for bind the variables which is better than generating code in this case it doesnt clutter your codebase with generated code and since env is initialized once, doesnt cause performance issues. 

## Features

- Load env variable from env files
- Bind structs to the env varaibles to provide a schema for your env files.

## Methods

1. `func LoadEnv(fileName string)` loads the env variables from the env file.
2. `func BindEnv(envStructPtr interface{})` binds the pointer to a struct to respective env variables (the name provided in the struct tags will be used)

## Struct fields

```go
// example
type Config struct {
	ProjectName string   `env:"APP_NAME"`
	Difficulty  []string `env_delim:";"`
	Port        int      `env_def:"8080"`
	Email       *struct {
		Host      string
		Port      int
		User      string
		Pass      string
		Signature string
	} `env_prefix:"EMAIL"`
	UserKeys map[string]string `env_keys:"USER*"`
}

```

```env
APP_NAME="my-game"
DIFFICULTY="easy;medium;hard"
PORT=4545
EMAIL_HOST=smtp.mailtrap.io
EMAIL_PORT=2525
EMAIL_USER="user@example.com"
EMAIL_PASS="password123"
EMAIL_SIGNATURE="Thanks,
${APP_NAME} Team
Contact: support@gmail.com"
USER1="hello-user-1"
USER2="hello-user-2"
USER3="hello-user-3"
```

1. env to name the key linked with a field
2. env_def used to set default value for a field
3. env_delim used to specify delimeter for spliting the string for environment variable
4. env_prefix to specify prefix for a group of fields
5. env_keys to specify list of keys to be binded to the field.  Wilcard * can also be used to specify prefix.

## ENV Format

Go env manager can parse env files with quotes, double quotes, back ticks. It also supports substituion of varaibles in the env file.
