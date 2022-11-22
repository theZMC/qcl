package qcl

import (
	"reflect"
	"testing"
	"time"
)

type (
	TestConfig struct {
		Host string
		Port int
	}

	TestNestedConfig struct {
		Host string
		Port int
		SSL  bool
		DB   TestDBConfig
	}

	TestDBConfig struct {
		Host string
		Port int
		SSL  bool
	}

	AllSupportedTypes struct {
		Bool     bool
		Int      int
		Int8     int8
		Int16    int16
		Int32    int32
		Int64    int64
		Uint     uint
		Uint8    uint8
		Uint16   uint16
		Uint32   uint32
		Uint64   uint64
		Float    float32
		Float8   float64
		Duration time.Duration
	}

	TestSliceConfig struct {
		Hosts []string
		Ports []int
	}

	TestMapConfig struct {
		Hosts map[string]string
		Ports map[string]int
	}

	TestPointerConfig struct {
		Host *string
		Port *int
	}

	TestNestedPointerConfig struct {
		Host *string
		Port *int
		SSL  *bool
		DB   *TestDBConfig
	}

	TestConfigWithStructTag struct {
		NotHost string `mytag:"HOST"`
		NotPort int    `mytag:"PORT"`
	}

	TestEmbeddedConfig struct {
		TestConfig
	}
)

func Test_splitOnWordBoundaries(t *testing.T) {
	tests := map[string]struct {
		want []string
	}{
		"": {
			want: []string{},
		},
		"h": {
			want: []string{"h"},
		},
		"Host": {
			want: []string{"Host"},
		},
		"HostPort": {
			want: []string{"Host", "Port"},
		},
		"HostPortSSL": {
			want: []string{"Host", "Port", "SSL"},
		},
		"HostPortSSLTimeout": {
			want: []string{"Host", "Port", "SSL", "Timeout"},
		},
		"HostPortSSLTimeoutMaxIdleConns": {
			want: []string{"Host", "Port", "SSL", "Timeout", "Max", "Idle", "Conns"},
		},
		"HostPortSSLTimeoutMaxIdleConnsMaxIdleConnsPerHost": {
			want: []string{"Host", "Port", "SSL", "Timeout", "Max", "Idle", "Conns", "Max", "Idle", "Conns", "Per", "Host"},
		},
		"HostPortSSLTimeoutMaxIdleConnsMaxIdleConnsPerHostDB": {
			want: []string{"Host", "Port", "SSL", "Timeout", "Max", "Idle", "Conns", "Max", "Idle", "Conns", "Per", "Host", "DB"},
		},
	}
	for input, test := range tests {
		t.Run(input, func(t *testing.T) {
			if got := splitOnWordBoundaries(input); !reflect.DeepEqual(got, test.want) {
				t.Errorf("splitOnWordBoundaries(%v) = %v, want %v", input, got, test.want)
			}
		})
	}
}

func Test_setMapKeysAndValues(t *testing.T) {
	tests := map[string]struct {
		inputKeys []string
		inputVals []string
		want      any
		wantErr   bool
	}{
		"empty": {
			inputKeys: []string{},
			inputVals: []string{},
			want:      map[string]string{},
		},
		"one": {
			inputKeys: []string{"key"},
			inputVals: []string{"val"},
			want:      map[string]string{"key": "val"},
		},
		"two": {
			inputKeys: []string{"key1", "key2"},
			inputVals: []string{"val1", "val2"},
			want:      map[string]string{"key1": "val1", "key2": "val2"},
		},
		"key value mismatch": {
			inputKeys: []string{"key1", "key2"},
			inputVals: []string{"val1"},
			wantErr:   true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := make(map[string]string)
			err := setMapKeysAndValues(reflect.ValueOf(got), test.inputKeys, test.inputVals, "")
			if (err != nil) != test.wantErr {
				t.Errorf("setMapKeysAndValues() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr && !reflect.DeepEqual(got, test.want) {
				t.Errorf("setMapKeysAndValues() = %v, want %v", got, test.want)
			}
		})
	}
	t.Run("not a map", func(t *testing.T) {
		got := ""
		err := setMapKeysAndValues(reflect.ValueOf(got), []string{}, []string{}, "")
		if err == nil {
			t.Errorf("setMapKeysAndValues() error = %v, wantErr %v", err, true)
		}
	})
	t.Run("unsettable type", func(t *testing.T) {
		got := map[string]int{}
		err := setMapKeysAndValues(reflect.ValueOf(got), []string{"something"}, []string{"this isn't an int"}, "")
		if err == nil {
			t.Errorf("setMapKeysAndValues() error = %v, wantErr %v", err, true)
		}
	})
}

func Test_setSliceValues(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  any
	}{
		"empty": {
			input: []string{},
			want:  []string{},
		},
		"one": {
			input: []string{"one"},
			want:  []string{"one"},
		},
		"two": {
			input: []string{"one", "two"},
			want:  []string{"one", "two"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := make([]string, 0)
			err := setSliceValues(reflect.ValueOf(&got).Elem(), test.input, "")
			if err != nil {
				t.Errorf("setSliceValues() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("setSliceValues() = %v, want %v", got, test.want)
			}
		})
	}
	t.Run("not a slice", func(t *testing.T) {
		got := ""
		err := setSliceValues(reflect.ValueOf(got), []string{}, "")
		if err == nil {
			t.Errorf("setSliceValues() error = %v, wantErr %v", err, true)
		}
	})
	t.Run("unsettable type", func(t *testing.T) {
		got := []int{}
		err := setSliceValues(reflect.ValueOf(got), []string{"this isn't an int"}, "")
		if err == nil {
			t.Errorf("setSliceValues() error = %v, wantErr %v", err, true)
		}
	})
}
