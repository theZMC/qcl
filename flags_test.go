package qcl

import (
	"flag"
	"os"
	"reflect"
	"testing"
	"time"
)

func Test_loadFromFlags(t *testing.T) {
	tests := map[string]struct {
		args []string
		want any
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
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			os.Args = append([]string{"test"}, test.args...)

			got := reflect.New(reflect.TypeOf(test.want).Elem()).Interface()
			if err := loadFromFlags(got); err != nil {
				t.Fatalf("LoadFromFlags() error = %v", err)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("LoadFromFlags() got = %v, want %v", got, test.want)
			}
		})
	}
}
