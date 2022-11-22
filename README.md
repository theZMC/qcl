# QCL: Quick Config Loader
> **GO 1.18+ ONLY** This library makes use of generics, which are only available in Go 1.18+

[![codecov](https://codecov.io/gh/theZMC/qcl/branch/main/graph/badge.svg?token=CPKRYCUKSU)](https://codecov.io/gh/theZMC/qcl)
[![CI](https://github.com/theZMC/qcl/actions/workflows/ci.yml/badge.svg)](https://github.com/theZMC/qcl/actions/workflows/ci.yml)

`qcl` is a lightweight library for loading configuration values at runtime. It is designed to have a simple API, robust test suite, zero external dependencies, and be easy to integrate into existing projects. If you are looking for a more full-featured configuration library, check out [Viper](https://github.com/spf13/viper) or [Koanf](https://github.com/knadh/koanf). I've used both and they are great libraries, but I wanted something simpler for my use cases. I currently have no plans to support loading configuration from files, so if your use-case requires that, this library is not for you.

> **BE ADVISED** This library is still under active development, but the API is stable and will not change before 1.0.0. The test suite is pretty robust, but I'm sure there are edge cases I haven't thought of. If you find a bug, please open an issue.

## Installation
```shell
go get github.com/thezmc/qcl
```

## Simple Example
Make sure the environment variables you want to use are set. For example:
```shell
export HOST="localhost"
export PORT="8080"
```
You can also use command line arguments. For example:
```shell
go run main.go --port 8081
```
```go
type Config struct {
  Host string
  Port int
  SSL  bool
}

defaultConfig := Config{
  Host: "anotherhost",
  Port: 9090,
  SSL:  false,
}

conf, _ := qcl.Load(&defaultConfig)

fmt.Printf("Host: %s\n", conf.Host) // "Host: localhost" from environment
fmt.Printf("Port: %d\n", conf.Port) // "Port: 8081" from command line, overrides environment by default
fmt.Printf("SSL: %t\n", conf.SSL)   // "SSL: false" from the "defaultConfig" struct
```
## Default Behavior
By default, the library will use the field names as the environment variable / command-line argument names. The environment variables will be loaded first, followed by the command line arguments. If a value is found in both the environment and command line arguments, the command line argument will take precedence [**except** in the case of slice and map values.](#slice-and-map-values)

### Environment Variables
By default, the library will look for environment variables with the same name as the struct fields, but split along word boundaries:
```go
type Config struct {
  Host   string // "HOST" environment variable
  DBHost string // "DB_HOST" environment variable
  DBPort int    // "DB_PORT" environment variable
}
```
You can override the environment variable name by using the `env` tag:
```go
type Config struct {
  Host     string `env:"HoSt"` // "HOST" environment variable
  HTTPPort int    `env:"PORT"` // "PORT" environment variable
}
```
**NOTE:** The override is case-insensitive. The library will convert the tag value to uppercase before looking for the environment variable.

### Command Line Arguments
By default, the library will look for command line arguments with the same name as the struct fields, but split along word boundaries:
```go
type Config struct {
  Host   string // "--host" command line argument
  DBHost string // "--db-host" command line argument
  DBPort int    // "--db-port" command line argument
}
```
You can override the command line argument name by using the `flag` tag:
```go
type Config struct {
  Host     string                // "--host" command line argument
  HTTPPort int    `flag:"port"`  // "--port" command line argument
}
```

### Slice and Map Values
Slices and maps are special cases when it comes to overrides. If a slice or map value is found in the environment or command-line, it will be appended to the slice or map from the default config. For example:
```shell
export HOSTS="localhost,otherhost" # separate iterable values with a comma
go run main.go --hosts "yetanotherhost"
```
```go
type Config struct {
  Hosts []string
}

conf, _ := qcl.Load(&Config{})

fmt.Printf("Hosts: %s\n", conf.Hosts) // "Hosts: [localhost otherhost yetanotherhost]"
```
Same idea for maps:
```shell
export HOSTS="localhost=8080,otherhost=9090" # separate key-value pairs with a comma, separate keys and values with an equals sign
go run main.go --hosts "yetanotherhost=1234"
```
```go
type Config struct {
  Hosts map[string]int
}

conf := qcl.Load(&Config{})
fmt.Printf("Hosts: %s\n", conf.Hosts) // "Hosts: map[localhost:8080 otherhost:9090 yetanotherhost:1234]"
```
### Nested Structs
Nested structs are also supported. The field name for the nested struct will be used as the prefix for the environment variables and command-line arguments. For example:
```go
type Config struct {
  Host string   // "HOST" environment variable; "--host" command line argument
  DB   struct {
    Host string // "DB_HOST" environment variable; "--db-host" command line argument
    Port int    // "DB_PORT" environment variable; "--db-port" command line argument
  }
}
```

### Embedded Structs
Embedded structs are also supported. The embedded struct will be flattened into the parent struct and so will not have a prefix. For example:
```go
type Config struct {
  User     string // "USER" environment variable; "--user" command line argument
  struct {
    Host string // "HOST" environment variable; "--host" command line argument
    Port int    // "PORT" environment variable; "--port" command line argument
  }
}
// is treated the same as
type Config struct {
  User string // "USER" environment variable; "--user" command line argument
  Host string // "HOST" environment variable; "--host" command line argument
  Port int    // "PORT" environment variable; "--port" command line argument
}
```

## Advanced Usage
### Custom Environment Variable Prefix
By default, the library doesn't expect any prefix on your environment variables. You can set a custom environment variable prefix by using the `qcl.WithEnvPrefix` functional option:
```go
type Config struct {
  Host string // "MYAPP_HOST" environment variable
}

qcl.Load(&Config{}, qcl.UseEnv(qcl.WithEnvPrefix("MYAPP_"))) // the _ on the end is optional. It will be added automatically if not included.
```

### Custom Environment Variable Struct Tag
By default, the `env` struct tag is used to override environment variable names. You can set a custom environment variable struct tag by using the `qcl.WithEnvStructTag` functional option:
```go
type Config struct {
  HTTPHost string `envvar:"HOST"` // "HOST" environment variable
}

qcl.Load(&Config{}, qcl.UseEnv(qcl.WithEnvStructTag("envvar")))
```
**NOTE:** The override is case-insensitive. The library will convert the tag value to uppercase before looking for the environment variable.

### Custom Environment Variable Iterable Separator
By default, iterables are separated by a comma. You can set a custom environment variable iterable separator by using the `qcl.WithEnvSeparator` functional option:
```shell
export HOSTS="localhost|otherhost" # separate iterable values with a pipe
```
```go
type Config struct {
  Hosts []string // "HOSTS" environment variable
}

qcl.Load(&Config{}, qcl.UseEnv(qcl.WithEnvSeparator("|")))
```

### Custom Load Order
By default, the library will load environment variables first, followed by command line arguments. You can set a custom load order by using the `qcl.InThisOrder` functional option:
```go
qcl.Load(&Config{}, qcl.InThisOrder(qcl.Flag, qcl.Environment))
```

## Extending the Library
### Custom Loaders
You can create your own loaders
```go
const JSON qcl.Source = "json" // define a new source. qcl.Source is just a type alias for string

func UseJSON(path string) LoadOption {
	return func(lc *qcl.LoadConfig) {                   // define a new functional option which
		lc.Sources = append(lc.Sources, JSON)       // add the new source to the list of sources. Be aware of the order!
		lc.Loaders[JSON] = func(config any) error { // define a new loader for the new source that implements qcl.Loader
                        // do something with the path...
		}
	}
}
```
> **NOTE:** The order of the sources is important. The library will load the values from the sources in the order they
> are defined. If a value is found in multiple sources, the value from the last source will be used. You can use the
> `qcl.InThisOrder` functional option to change the order of the sources if needed.

## License
[MIT](LICENSE)

## Contributing
Bug reports and pull requests are welcome at the [issues page](https://github.com/theZMC/qcl/issues). For major changes, please open an issue first to discuss what you would like to change. Any new feature PRs must include adequate test coverage and documentation. Any features importing packages that aren't in the standard library will not be accepted.

## Author
[Zach Callahan](https://zmc.dev)
