# Go Env Manager

**Go Env Manager** is a simple, lightweight package to manage environment variables in Go projects.
It uses **reflection** to bind environment variables to struct fields, avoiding the need for code generation.
This keeps your codebase clean, while the one-time initialization ensures no noticeable performance cost.

## Features

* **Load** environment variables from `.env` files.
* **Bind** structs to environment variables to provide a schema and validation for your env files.
* **Supports**:

  * Default values
  * Field-specific delimiters
  * Nested structs with prefixes
  * Map binding from multiple env keys
  * Wildcard key matching

---

## Methods

1. **`func LoadEnv(fileName string)`**
   Loads environment variables from the given `.env` file.

2. **`func BindEnv(envStructPtr interface{})`**
   Binds a pointer to a struct to the respective environment variables.
   The struct tags define the mapping.

---

## Struct Field Tags

| Tag          | Description                                                               |
| ------------ | ------------------------------------------------------------------------- |
| `env`        | The env variable name linked to the field.                                |
| `env_def`    | Default value if the variable is missing.                                 |
| `env_delim`  | Delimiter for splitting values into slices.                               |
| `env_prefix` | Prefix for all env variables in a nested struct.                          |
| `env_keys`   | List of env keys for maps. Supports `*` wildcard to match keys by prefix. |

---

### Example

```go
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

---

### Example `.env` File

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

---

## ENV File Parsing

Go Env Manager supports:

* Quotes: `'single'`, `"double"`, `` `backtick` ``
* Variable substitution inside values (`${VAR_NAME}`)
* Flexible delimiters for lists and maps
* Multi-line values
