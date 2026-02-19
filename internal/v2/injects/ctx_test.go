package injects

import (
	"maps"
	"reflect"
	"testing"
)

func TestImports(t *testing.T) {
	tests := []struct {
		name    string
		imports []string
		want    map[string]string
	}{
		{
			name: "duplicated imports",
			imports: []string{
				"example.com/test",
				"example.com/test",
			},
			want: map[string]string{
				"test": "example.com/test",
			},
		},
		{
			name: "same name",
			imports: []string{
				"example.com/test",
				"lol.com/test",
			},
			want: map[string]string{
				"test":  "example.com/test",
				"test1": "lol.com/test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewCtx()
			for _, pkg := range tt.imports {
				ctx.Import(pkg)
			}

			got := maps.Collect(ctx.Imports())
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
