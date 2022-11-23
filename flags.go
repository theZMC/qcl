package qcl

import (
	"errors"
	"flag"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const Flag Source = "flag"

// UseFlags enables configuration from command line flags. Currently, the flag loader is not configurable. It will
// use the struct field names as the flag names, but lowercased and spit on word boundaries with a dash. For example,
// the field name "FooBar" will be converted to "foo-bar". You can override the flag name by using the "flag" struct
// tag. Examples:
//
//	type Config struct {
//	    FooBar string // will look for -foo-bar flag
//	}
//	type Config struct {
//	    FooBar string `flag:"foo"` // will look for -foo flag
//	}
//	type Config struct {
//	    FooBar string `flag:"foo.bar"` // will look for -foo.bar flag
//	}
//
// By default, calling Load() without any LoadOptions will use the flag loader as well as the environment loader, with
// the flag loader taking precedence. If you want to use only the flag loader, you can call Load with just the UseFlags
// option:
//
//	Load(&config, UseFlags()) // will only use flags
func UseFlags() LoadOption {
	return func(o *LoadConfig) {
		o.Sources = append(o.Sources, Flag)
		o.Loaders[Flag] = loadFromFlags
	}
}

func loadFromFlags(config any) error {
	if len(os.Args) < 2 {
		return nil
	}

	if reflect.TypeOf(config).Kind() != reflect.Ptr {
		return ConfigTypeError
	}
	val := reflect.ValueOf(config).Elem()
	typ := val.Type()

	if err := bindFlags(val, typ, ""); err != nil {
		return err
	}

	flag.Parse()
	return nil
}

func bindFlags(val reflect.Value, typ reflect.Type, name string) error {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous {
			if err := bindFlags(val.Field(i), field.Type, ""); err != nil {
				return err
			}
			continue
		}
		flagName := strings.ToLower(field.Name)
		if tag := field.Tag.Get("flag"); tag != "" {
			flagName = tag
		}
		if name != "" && !strings.HasSuffix(name, ".") {
			name += "."
		}
		flagName = name + strings.Join(splitOnWordBoundaries(flagName), ".")
		if val := val.Field(i); val.CanSet() {
			if val.Kind() == reflect.Ptr {
				if val.IsNil() {
					val.Set(reflect.New(val.Type().Elem()))
				}
				val = val.Elem()
			}
			if val.Kind() == reflect.Struct {
				if err := bindFlags(val, val.Type(), flagName); err != nil {
					return err
				}
				continue
			}
			if err := bindFlag(val, flagName); err != nil {
				return err
			}
		}
	}
	return nil
}

func bindFlag(v reflect.Value, flagName string) error {
	if !v.CanSet() {
		return UnsupportedTypeError{v.Kind()}
	}
	if v.Type().String() == "time.Duration" {
		flag.DurationVar(v.Addr().Interface().(*time.Duration), flagName, time.Duration(0), "")
		return nil
	}
	switch v.Kind() {
	case reflect.String:
		flag.Var(&stringValue{v}, flagName, "")
	case reflect.Bool:
		flag.Var(&boolValue{v}, flagName, "")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flag.Var(&intValue{v}, flagName, "")
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		flag.Var(&uintValue{v}, flagName, "")
	case reflect.Float32, reflect.Float64:
		flag.Var(&floatValue{v}, flagName, "")
	case reflect.Slice:
		if v.IsNil() {
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}
		flag.Var(&sliceValue{v}, flagName, "")
	case reflect.Map:
		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
		flag.Var(&mapValue{v}, flagName, "")
	default:
		return UnsupportedTypeError{v.Kind()}
	}
	return nil
}

type (
	stringValue struct{ reflect.Value }
	boolValue   struct{ reflect.Value }
	sliceValue  struct{ reflect.Value }
	mapValue    struct{ reflect.Value }
	intValue    struct{ reflect.Value }
	uintValue   struct{ reflect.Value }
	floatValue  struct{ reflect.Value }
)

func (s *stringValue) Set(value string) error {
	s.SetString(value)
	return nil
}
func (b *boolValue) Set(value string) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	b.SetBool(v)
	return nil
}
func (s *sliceValue) Set(value string) error {
	vals := strings.Split(value, ",")
	return setSliceValues(s.Value, vals, "")
}
func (m *mapValue) Set(value string) error {
	parts := strings.SplitN(value, ",", 2)
	keys := make([]string, 0)
	values := make([]string, 0)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return errors.New("invalid map value")
		}
		keys = append(keys, kv[0])
		values = append(values, kv[1])
	}
	return setMapKeysAndValues(m.Value, keys, values, "")
}
func (i *intValue) Set(value string) error {
	kind := i.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, i.Type().Bits())
		if err != nil {
			return err
		}
		i.SetInt(v)
	default:
		return UnsupportedTypeError{kind}
	}
	return nil
}
func (u *uintValue) Set(value string) error {
	kind := u.Kind()
	switch kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, u.Type().Bits())
		if err != nil {
			return err
		}
		u.SetUint(v)
	default:
		return UnsupportedTypeError{kind}
	}
	return nil
}
func (f *floatValue) Set(value string) error {
	kind := f.Kind()
	switch kind {
	case reflect.Float32:
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return err
		}
		f.SetFloat(v)
	case reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.SetFloat(v)
	default:
		return UnsupportedTypeError{kind}
	}
	return nil
}
