package utils

import (
	"reflect"
	"testing"
)

func TestLast(t *testing.T) {
	type args[E any] struct {
		s []E
	}
	type testCase[E any] struct {
		name string
		args args[E]
		want E
		ok   bool
	}

	tests := []testCase[string]{
		{
			name: "returns last string in slice",
			args: args[string]{s: []string{"foo", "bar"}},
			want: "bar",
			ok:   true,
		},
		{
			name: "zero length slice",
			args: args[string]{s: []string{}},
			want: "",
			ok:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := Last(tt.args.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Last() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.ok {
				t.Errorf("Last() got1 = %v, want %v", got1, tt.ok)
			}
		})
	}
}
