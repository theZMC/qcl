package qcl

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

func Test_Load(t *testing.T) {
	tests := map[string]struct {
		want    *TestConfig
		args    []string
		wantErr bool
	}{
		"no prefix": {
			want: &TestConfig{
				Host: "localhost",
				Port: 8080,
			},
			args: []string{"-host", "localhost", "-port", "8080"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			os.Args = append([]string{"test"}, test.args...)
			got, err := Load[TestConfig](nil)
			if test.wantErr && err == nil {
				t.Errorf("Load[TestConfig](nil) error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && !reflect.DeepEqual(got, test.want) {
				t.Errorf("Load() got = %v, want %v", got, test.want)
			}
		})
	}
	t.Run("error", func(t *testing.T) {
		_, err := Load[UnsupportedStruct](nil)
		if err == nil {
			t.Errorf("Load[UnsupportedStruct](nil) error = %v, wantErr %v", err, true)
		}
	})
}
