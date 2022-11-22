package qcl

import (
	"reflect"
	"testing"
	"time"
)

func Test_loadFromEnv(t *testing.T) {
	tests := map[string]struct {
		prefix    string
		structTag string
		want      any
		envs      map[string]string
		wantErr   bool
	}{
		"no prefix": {
			want: &TestConfig{
				Host: "localhost",
				Port: 8080,
			},
			envs: map[string]string{
				"HOST": "localhost",
				"PORT": "8080",
			},
		},
		"prefix": {
			prefix: "TEST",
			want: &TestConfig{
				Host: "localhost",
				Port: 8080,
			},
			envs: map[string]string{
				"TEST_HOST": "localhost",
				"TEST_PORT": "8080",
			},
		},
		"prefix with underscore": {
			prefix: "TEST_",
			want: &TestConfig{
				Host: "localhost",
				Port: 8080,
			},
			envs: map[string]string{
				"TEST_HOST": "localhost",
				"TEST_PORT": "8080",
			},
		},
		"prefix with multiple underscores": {
			prefix: "TEST__",
			want: &TestConfig{
				Host: "localhost",
				Port: 8080,
			},
			envs: map[string]string{
				"TEST__HOST": "localhost",
				"TEST__PORT": "8080",
			},
		},
		"nested config": {
			prefix: "TEST",
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
			envs: map[string]string{
				"TEST_HOST":    "localhost",
				"TEST_PORT":    "8080",
				"TEST_SSL":     "true",
				"TEST_DB_HOST": "localhost",
				"TEST_DB_PORT": "5432",
				"TEST_DB_SSL":  "true",
			},
		},
		"all supported types": {
			prefix: "TEST",
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
			envs: map[string]string{
				"TEST_BOOL":     "true",
				"TEST_INT":      "1",
				"TEST_INT8":     "2",
				"TEST_INT16":    "3",
				"TEST_INT32":    "4",
				"TEST_INT64":    "5",
				"TEST_UINT":     "6",
				"TEST_UINT8":    "7",
				"TEST_UINT16":   "8",
				"TEST_UINT32":   "9",
				"TEST_UINT64":   "10",
				"TEST_FLOAT":    "11.1",
				"TEST_FLOAT8":   "12.2",
				"TEST_DURATION": "13s",
			},
		},
		"slice": {
			prefix: "TEST",
			want: &TestSliceConfig{
				Hosts: []string{"localhost", "somehost"},
				Ports: []int{8080, 8081},
			},
			envs: map[string]string{
				"TEST_HOSTS": "localhost,somehost",
				"TEST_PORTS": "8080,8081",
			},
		},
		"map": {
			prefix: "TEST",
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
			envs: map[string]string{
				"TEST_HOSTS": "localhost=127.0.0.1,somehost=10.0.0.1",
				"TEST_PORTS": "localhost=8080,somehost=8081",
			},
		},
		"pointer": {
			prefix: "TEST",
			want: &TestPointerConfig{
				Host: ptr("localhost"),
				Port: ptr(8080),
			},
			envs: map[string]string{
				"TEST_HOST": "localhost",
				"TEST_PORT": "8080",
			},
		},
		"nested pointer": {
			prefix: "TEST",
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
			envs: map[string]string{
				"TEST_HOST":    "localhost",
				"TEST_PORT":    "8080",
				"TEST_SSL":     "true",
				"TEST_DB_HOST": "localhost",
				"TEST_DB_PORT": "5432",
				"TEST_DB_SSL":  "true",
			},
		},
		"config with custom tag": {
			prefix:    "TEST",
			structTag: "mytag",
			want: &TestConfigWithStructTag{
				NotHost: "localhost",
				NotPort: 8080,
			},
			envs: map[string]string{
				"TEST_HOST": "localhost",
				"TEST_PORT": "8080",
			},
		},
		"embedded config": {
			prefix: "TEST",
			want: &TestEmbeddedConfig{
				TestConfig: TestConfig{
					Host: "localhost",
					Port: 8080,
				},
			},
			envs: map[string]string{
				"TEST_HOST": "localhost",
				"TEST_PORT": "8080",
			},
		},
		"unparseable bool": {
			prefix: "TEST",
			want:   &AllSupportedTypes{},
			envs: map[string]string{
				"TEST_BOOL": "not a bool",
			},
			wantErr: true,
		},
		"unparseable int": {
			prefix: "TEST",
			want:   &AllSupportedTypes{},
			envs: map[string]string{
				"TEST_INT": "not an int",
			},
			wantErr: true,
		},
		"unparseable uint": {
			prefix: "TEST",
			want:   &AllSupportedTypes{},
			envs: map[string]string{
				"TEST_UINT": "not a uint",
			},
			wantErr: true,
		},
		"unparseable float": {
			prefix: "TEST",
			want:   &AllSupportedTypes{},
			envs: map[string]string{
				"TEST_FLOAT": "not a float",
			},
			wantErr: true,
		},
		"unparseable duration": {
			prefix: "TEST",
			want:   &AllSupportedTypes{},
			envs: map[string]string{
				"TEST_DURATION": "not a duration",
			},
			wantErr: true,
		},
		"unparseable slice": {
			prefix: "TEST",
			want:   &TestSliceConfig{},
			envs: map[string]string{
				"TEST_HOSTS": "localhost,somehost",
				"TEST_PORTS": "8080,8081,not an int",
			},
			wantErr: true,
		},
		"unparseable map": {
			prefix: "TEST",
			want:   &TestMapConfig{},
			envs: map[string]string{
				"TEST_HOSTS": "localhost=127.0.0.1,somehost=10.0.0.1",
				"TEST_PORTS": "localhost=8080,somehost=not an int",
			},
			wantErr: true,
		},
		"key with no value": {
			prefix: "TEST",
			want:   &TestMapConfig{},
			envs: map[string]string{
				"TEST_HOSTS": "localhost",
			},
			wantErr: true,
		},
		"unsupported type": {
			prefix: "TEST",
			want: &struct {
				UnsupportedType chan int
			}{},
			envs: map[string]string{
				"TEST_UNSUPPORTED_TYPE": "unsupported",
			},
			wantErr: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for k, v := range test.envs {
				t.Setenv(k, v)
			}

			got := reflect.New(reflect.TypeOf(test.want).Elem()).Interface()
			envConf := new(envConfig)
			envConf.prefix = test.prefix
			envConf.separator = ","
			envConf.structTag = test.structTag

			err := loadFromEnv(envConf)(got)
			if (err != nil) != test.wantErr {
				t.Errorf("loadFromEnv() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr && !reflect.DeepEqual(got, test.want) {
				t.Errorf("loadFromEnv() got = %v, want %v", got, test.want)
			}
		})
	}
	t.Run("non-pointer config", func(t *testing.T) {
		err := loadFromEnv(nil)(TestConfig{})
		if err == nil {
			t.Error("loadFromEnv()() should return an error for non-pointer config")
		}
	})
}

func ptr[T any](v T) *T { return &v }
