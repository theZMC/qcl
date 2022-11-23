package qcl

import (
	"flag"
	"os"
	"reflect"
	"testing"
	"time"
)

type TestConfigWithFlagTag struct {
	HTTPHost string `flag:"host"`
	HTTPPort int    `flag:"port"`
}

func Test_UseFlags(t *testing.T) {
	lc := LoadConfig{
		Loaders: make(map[Source]Loader),
	}
	UseFlags()(&lc)
	if len(lc.Sources) != 1 {
		t.Errorf("UseFlags() should add one source")
	}
	if lc.Sources[0] != Flag {
		t.Errorf("UseFlags() should add Flag source")
	}
	if lc.Loaders[Flag] == nil {
		t.Errorf("UseFlags() should add Flag loader")
	}
}

func Test_loadFromFlags(t *testing.T) {
	tests := map[string]struct {
		args    []string
		want    any
		wantErr bool
	}{
		"no flags": {
			args: []string{},
			want: &TestConfig{
				Host: "",
				Port: 0,
			},
		},
		"host": {
			args: []string{"-host", "localhost"},
			want: &TestConfig{
				Host: "localhost",
				Port: 0,
			},
		},
		"host and port": {
			args: []string{"-host", "localhost", "-port", "8080"},
			want: &TestConfig{
				Host: "localhost",
				Port: 8080,
			},
		},
		"nested config": {
			want: &TestNestedConfig{
				Host: "localhost",
				Port: 8080,
				SSL:  true,
				DB: TestDBConfig{
					Host: "localhost",
					Port: 5432,
					SSL:  true,
				},
			},
			args: []string{
				"-host", "localhost",
				"-port", "8080",
				"-ssl", "true",
				"-db.host", "localhost",
				"-db.port", "5432",
				"-db.ssl", "true",
			},
		},
		"all supported types": {
			want: &AllSupportedTypes{
				Bool:     true,
				Int:      1,
				Int8:     2,
				Int16:    3,
				Int32:    4,
				Int64:    5,
				Uint:     6,
				Uint8:    7,
				Uint16:   8,
				Uint32:   9,
				Uint64:   10,
				Float:    11.1,
				Float8:   12.2,
				Duration: 13 * time.Second,
			},
			args: []string{
				"-bool", "true",
				"-int", "1",
				"-int8", "2",
				"-int16", "3",
				"-int32", "4",
				"-int64", "5",
				"-uint", "6",
				"-uint8", "7",
				"-uint16", "8",
				"-uint32", "9",
				"-uint64", "10",
				"-float", "11.1",
				"-float8", "12.2",
				"-duration", "13s",
			},
		},
		"slice": {
			want: &TestSliceConfig{
				Hosts: []string{"localhost", "somehost"},
				Ports: []int{8080, 8081},
			},
			args: []string{
				"-hosts", "localhost,somehost",
				"-ports", "8080,8081",
			},
		},
		"map": {
			want: &TestMapConfig{
				Hosts: map[string]string{
					"localhost": "127.0.0.1",
					"somehost":  "10.0.0.1",
				},
				Ports: map[string]int{
					"localhost": 8080,
					"somehost":  8081,
				},
			},
			args: []string{
				"-hosts", "localhost=127.0.0.1,somehost=10.0.0.1",
				"-ports", "localhost=8080,somehost=8081",
			},
		},
		"pointer": {
			want: &TestPointerConfig{
				Host: ptr("localhost"),
				Port: ptr(8080),
			},
			args: []string{
				"-host", "localhost",
				"-port", "8080",
			},
		},
		"nested pointer": {
			want: &TestNestedPointerConfig{
				Host: ptr("localhost"),
				Port: ptr(8080),
				SSL:  ptr(true),
				DB: &TestDBConfig{
					Host: "localhost",
					Port: 5432,
					SSL:  true,
				},
			},
			args: []string{
				"-host", "localhost",
				"-port", "8080",
				"-ssl", "true",
				"-db.host", "localhost",
				"-db.port", "5432",
				"-db.ssl", "true",
			},
		},
		"embedded struct": {
			want: &TestEmbeddedConfig{
				TestConfig: TestConfig{
					Host: "localhost",
					Port: 8080,
				},
			},
			args: []string{
				"-host", "localhost",
				"-port", "8080",
			},
		},
		"unsupported type": {
			want: &UnsupportedStruct{
				Unsupported: make(chan int),
			},
			args: []string{
				"-unsupported", "unsupported",
			},
			wantErr: true,
		},
		"nested unsupported type": {
			want: &struct {
				Unsupported UnsupportedStruct
			}{},
			args: []string{
				"-unsupported-unsupported", "unsupported",
			},
			wantErr: true,
		},
		"nested anonymous unsupported type": {
			want: &struct {
				UnsupportedStruct
			}{},
			args: []string{
				"-unsupported", "unsupported",
			},
			wantErr: true,
		},
		"flag tag override": {
			want: &TestConfigWithFlagTag{
				HTTPHost: "localhost",
				HTTPPort: 8080,
			},
			args: []string{
				"-host", "localhost",
				"-port", "8080",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			os.Args = append([]string{"test"}, test.args...)

			got := reflect.New(reflect.TypeOf(test.want).Elem()).Interface()
			if err := loadFromFlags(got); err != nil && !test.wantErr {
				t.Errorf("loadFromFlags() error = %v, wantErr %v", err, test.wantErr)
			}

			if !test.wantErr && !reflect.DeepEqual(got, test.want) {
				t.Errorf("LoadFromFlags() got = %v, want %v", got, test.want)
			}
		})
	}
	t.Run("non-pointer config", func(t *testing.T) {
		if err := loadFromFlags(TestConfig{}); err == nil {
			t.Error("LoadFromFlags() expected error, got nil")
		}
	})
}

func Test_bindFlag(t *testing.T) {
	t.Run("unsettable type", func(t *testing.T) {
		if err := bindFlag(reflect.ValueOf(make(chan bool)), "test"); err == nil {
			t.Error("bindFlag() expected error, got nil")
		}
	})
}

func Test_boolValue(t *testing.T) {
	tests := map[string]struct {
		value   string
		want    bool
		wantErr bool
	}{
		"true": {
			value: "true",
			want:  true,
		},
		"false": {
			value: "false",
			want:  false,
		},
		"invalid": {
			value:   "invalid",
			wantErr: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got bool
			bv := boolValue{reflect.ValueOf(&got).Elem()}
			if err := bv.Set(test.value); err != nil && !test.wantErr {
				t.Errorf("boolValue.Set() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && got != test.want {
				t.Errorf("boolValue.Set() got = %v, want %v", got, test.want)
			}
		})
	}
}

func Test_intValue(t *testing.T) {
	tests := map[string]struct {
		value   string
		want    int
		wantErr bool
	}{
		"valid": {
			value: "123",
			want:  123,
		},
		"invalid": {
			value:   "invalid",
			wantErr: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got int
			iv := intValue{reflect.ValueOf(&got).Elem()}
			if err := iv.Set(test.value); err != nil && !test.wantErr {
				t.Errorf("intValue.Set() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && got != test.want {
				t.Errorf("intValue.Set() got = %v, want %v", got, test.want)
			}
		})
	}
	t.Run("unsupported type", func(t *testing.T) {
		iv := intValue{reflect.ValueOf(make(chan int))}
		if err := iv.Set("123"); err == nil {
			t.Error("intValue.Set() expected error, got nil")
		}
	})
}

func Test_uintValue(t *testing.T) {
	tests := map[string]struct {
		value   string
		want    uint
		wantErr bool
	}{
		"valid": {
			value: "123",
			want:  123,
		},
		"invalid": {
			value:   "invalid",
			wantErr: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got uint
			uv := uintValue{reflect.ValueOf(&got).Elem()}
			if err := uv.Set(test.value); err != nil && !test.wantErr {
				t.Errorf("uintValue.Set() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && got != test.want {
				t.Errorf("uintValue.Set() got = %v, want %v", got, test.want)
			}
		})
	}
	t.Run("unsupported type", func(t *testing.T) {
		uv := uintValue{reflect.ValueOf(make(chan uint))}
		if err := uv.Set("123"); err == nil {
			t.Error("uintValue.Set() expected error, got nil")
		}
	})
}

func Test_floatValue(t *testing.T) {
	tests := map[string]struct {
		value   string
		want    float64
		wantErr bool
	}{
		"valid": {
			value: "123.45",
			want:  123.45,
		},
		"invalid": {
			value:   "invalid",
			wantErr: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got float64
			fv := floatValue{reflect.ValueOf(&got).Elem()}
			if err := fv.Set(test.value); err != nil && !test.wantErr {
				t.Errorf("floatValue.Set() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && got != test.want {
				t.Errorf("floatValue.Set() got = %v, want %v", got, test.want)
			}
		})
	}
	t.Run("float32 invalid", func(t *testing.T) {
		var got float32
		fv := floatValue{reflect.ValueOf(&got).Elem()}
		if err := fv.Set("invalid"); err == nil {
			t.Error("floatValue.Set() expected error, got nil")
		}
	})
	t.Run("unsupported type", func(t *testing.T) {
		fv := floatValue{reflect.ValueOf(make(chan float64))}
		if err := fv.Set("123"); err == nil {
			t.Error("floatValue.Set() expected error, got nil")
		}
	})
}

func Test_mapValue(t *testing.T) {
	tests := map[string]struct {
		value   string
		want    map[string]string
		wantErr bool
	}{
		"valid": {
			value: "key1=value1,key2=value2",
			want:  map[string]string{"key1": "value1", "key2": "value2"},
		},
		"invalid": {
			value:   "key1=value1,key2=value2,key3",
			wantErr: true,
		},
		"empty value": {
			value:   "key1",
			wantErr: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got map[string]string
			mv := mapValue{reflect.ValueOf(&got).Elem()}
			if err := mv.Set(test.value); err != nil && !test.wantErr {
				t.Errorf("mapValue.Set() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && !reflect.DeepEqual(got, test.want) {
				t.Errorf("mapValue.Set() got = %v, want %v", got, test.want)
			}
		})
	}
	t.Run("unsupported type", func(t *testing.T) {
		mv := mapValue{reflect.ValueOf(make(chan map[string]string))}
		if err := mv.Set("key1=value1,key2=value2"); err == nil {
			t.Error("mapValue.Set() expected error, got nil")
		}
	})
}
