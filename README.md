# Go Env Manager

Is a simple, light weight package to manage your env varaibles in a go project.

## Features

- Load env variable from env files
- Bind structs to the env varaibles to provide a schema for your env files.

## Methods

1. `func LoadEnv(fileName string)` loads the env variables from the env file.
2. `func BindEnv(envStructPtr interface{})` binds the pointer to a struct to respective env variables (the name provided in the struct tags will be used)

## ENV Format

Go env manager can parse env files with quotes, double quotes, back ticks. It also supports substituion of varaibles in the env file.
