package qcl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	InvalidMapValueError = func(values ...string) error { return fmt.Errorf("invalid map value: %s", values) }
	UnsupportedTypeError = func(kind reflect.Kind) error { return fmt.Errorf("unsupported type: %s", kind) }
)

// splitOnWordBoundaries splits a string on word boundaries. Word boundaries are capitalized letters followed immediately
// by a lowercase letter. For example, "FooBar" is split into "Foo" and "Bar". The first letter is always capitalized.
// This is useful for converting a camelCase or PascalCase string into a slice of words. It also handles acronyms,
// such as "HTTPServer" being split into "HTTP" and "Server".
func splitOnWordBoundaries(s string) []string {
	if len(s) == 0 {
		return []string{}
	}

	for i, c := range s {
		if i == 0 {
			continue
		}
		if i == len(s)-1 {
			return []string{s}
		}
		if unicode.IsUpper(c) { // if we hit an uppercase letter...
			if unicode.IsLower(rune(s[i-1])) { // ...and the previous character was lowercase
				return append([]string{s[:i]}, splitOnWordBoundaries(s[i:])...) // we found a word boundary. Split and recurse.
			}
			if unicode.IsUpper(rune(s[i-1])) { // ...and the previous character was uppercase...
				if unicode.IsLower(rune(s[i+1])) { // ...and the next character is lowercase
					return append([]string{s[:i]}, splitOnWordBoundaries(s[i:])...) // we found an initialism. Split and recurse.
				}
			}
		}
	}
	return []string{s}
}

func setMapKeysAndValues(v reflect.Value, keys, values []string, separator string) error {
	if v.Kind() != reflect.Map {
		return NotAMapError
	}

	if len(keys) != len(values) {
		return InvalidMapValueError(append(keys, values...)...)
	}
	// create a new map with the correct type and set it on the value if the map is nil
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}
	for i, key := range keys {
		newVal := reflect.New(v.Type().Elem())
		if err := setField(newVal.Elem(), values[i], separator); err != nil {
			return err
		}
		v.SetMapIndex(reflect.ValueOf(key), newVal.Elem())
	}
	return nil
}

func setSliceValues(v reflect.Value, values []string, separator string) error {
	if v.Kind() != reflect.Slice {
		return NotASliceError
	}
	if v.IsNil() {
		v.Set(reflect.MakeSlice(v.Type(), 0, len(values)))
	}
	for _, value := range values {
		newVal := reflect.New(v.Type().Elem())
		if err := setField(newVal.Elem(), value, separator); err != nil {
			return err
		}
		v.Set(reflect.Append(v, newVal.Elem()))
	}
	return nil
}

func setField(v reflect.Value, value string, separator string) error {
	if !v.CanSet() {
		return nil
	}
	// need to handle time.Duration before the switch..case since it qualifies as an int
	if v.Type().String() == "time.Duration" {
		d, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(d))
		return nil
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString(value)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	case reflect.Slice:
		return setSliceValues(v, strings.Split(value, separator), separator)
	case reflect.Map:
		kv := strings.Split(value, separator)
		keys := make([]string, len(kv))
		values := make([]string, len(kv))
		for i, kv := range kv {
			kv := strings.SplitN(kv, "=", 2)
			if len(kv) != 2 {
				return InvalidMapValueError(kv...)
			}
			keys[i] = kv[0]
			values[i] = kv[1]
		}
		return setMapKeysAndValues(v, keys, values, separator)
	default:
		return UnsupportedTypeError(v.Kind())
	}
	return nil
}
