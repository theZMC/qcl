package qcl

import (
	"errors"
	"os"
	"reflect"
	"strings"
)

const Environment Source = "env"

type envConfig struct {
	prefix    string
	structTag string
	separator string
}

var defaultEnvConfig = &envConfig{
	prefix:    "",
	structTag: "env",
	separator: ",",
}

type EnvOption func(*envConfig)

var (
	NotAMapError    = errors.New("not a map")
	NotASliceError  = errors.New("not a slice")
	ConfigTypeError = errors.New("config must be a pointer to a struct")
)

// UseEnv allows you to load configuration from environment variables. The environment variables are expected to be
// in all caps and separated by underscores. For example, a field named "FooBar" will be set by the environment
// variable "FOO_BAR".
//
// Example:
//
//	export FOO_BAR=baz
//
//	type Config struct {
//		FooBar string
//	}
//
//	var defaultConfig Config
//
//	ocl.Load(defaultConfig, ocl.UseEnv())
//
// will set the value of FooBar to the value of the environment variable "FOO_BAR".
func UseEnv(opts ...EnvOption) LoadOption {
	envConf := defaultEnvConfig

	for _, opt := range opts {
		opt(envConf)
	}
	return func(o *LoadConfig) {
		o.Sources = append(o.Sources, Environment)
		o.Loaders[Environment] = loadFromEnv(envConf)
	}
}

// WithEnvPrefix allows you to specify a prefix for environment variables. For example, if you specify "FOO" as the
// prefix, then an environment variable named "FOO_BAR" will be used to set a field named "Bar" in the config struct.
//
// Example:
//
//	type Config struct {
//		Bar string
//	}
//
//	WithEnvPrefix("FOO")
//
// will set the value of Bar to the value of the environment variable "FOO_BAR".
//
// The default is no prefix.
func WithEnvPrefix(prefix string) EnvOption {
	return func(c *envConfig) {
		c.prefix = prefix
	}
}

// WithEnvStructTag allows you to specify a custom struct tag to use for environment variable names. By default, the loader
// looks for the "env" struct tag, but if that's not found the field name itself is used as the environment variable
// name and in either case, it is split on word boundaries. For example, a field named "FooBar" will be set by the
// environment variable "FOO_BAR" by default.
//
// Example:
//
//	WithEnvStructTag("mytag")
//
//	type Config struct {
//		FooBar string `mytag:"FOO"` // FooBar will be set by the environment variable "FOO" instead of the default "FOO_BAR"
//	}
//
// By default, the environment loader looks for a struct tag "env" and in the absence of a struct tag, will use the field
// name itself.
func WithEnvStructTag(tag string) EnvOption {
	return func(c *envConfig) {
		c.structTag = tag
	}
}

// WithEnvSeparator allows you to specify a custom separator for environment variables that are setting iterables.
//
// Example:
//
//	WithEnvSeparator(";")
//
// will allow an environment variable like
//
//	export FOO="bar;baz"
//
// to set a field named "Foo" to a slice of strings with the values "bar" and "baz":
//
//	type Config struct {
//		Foo []string // foo will be set to []string{"bar", "baz"}
//	}
//
// This also works for maps where the key/value pairs are separated by the separator.
//
// Example:
//
//	export FOO="bar=baz;qux=quux"
//
// will set a field named "Foo" to a map with the key/value pairs bar=baz, and qux=quux.
//
//	type Config struct {
//		Foo map[string]string // foo will be set to map[string]string{"bar": "baz", "qux": "quux"}
//	}
//
// The default separator is a comma (,)
func WithEnvSeparator(separator string) EnvOption {
	return func(c *envConfig) {
		c.separator = separator
	}
}

func loadFromEnv(envConf *envConfig) Loader {
	if envConf == nil {
		envConf = defaultEnvConfig
	}
	if envConf.prefix != "" && !strings.HasSuffix(envConf.prefix, "_") {
		envConf.prefix += "_"
	}
	return func(config any) error {
		if reflect.TypeOf(config).Kind() != reflect.Ptr {
			return ConfigTypeError
		}
		val := reflect.ValueOf(config).Elem()
		typ := val.Type()
		return envSetFields(val, typ, envConf.prefix, envConf.structTag, envConf.separator)
	}
}

func envSetFields(val reflect.Value, typ reflect.Type, envPrefix, structTag, separator string) error {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fName := strings.Join(splitOnWordBoundaries(field.Name), "_")
		if structTag != "" {
			if tag, ok := field.Tag.Lookup(structTag); ok {
				tag = strings.Split(strings.TrimSpace(tag), ",")[0]
				fName = strings.Join(splitOnWordBoundaries(tag), "_")
			}
		}
		if val := val.Field(i); val.CanSet() {
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				if err := envSetFields(val, field.Type, envPrefix, structTag, separator); err != nil {
					return err
				}
			}
			if val.Kind() == reflect.Ptr {
				if val.IsNil() {
					val.Set(reflect.New(val.Type().Elem()))
				}
				val = val.Elem()
			}
			if val.Kind() == reflect.Struct {
				if err := envSetFields(val, val.Type(), envPrefix+fName+"_", structTag, separator); err != nil {
					return err
				}
			}
			if v := os.Getenv(strings.ToUpper(envPrefix + fName)); v != "" {
				if err := setField(val, v, separator); err != nil {
					return err
				}
			}
		}
	}
	return nil

}
