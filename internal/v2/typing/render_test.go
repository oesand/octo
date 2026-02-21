package typing_test

import (
	"testing"

	"github.com/oesand/octo/internal/v2/typing"
)

type testCtx struct {
	pkgs map[string]string
}

func (t *testCtx) ImportAlias(pkg string) string {
	return t.pkgs[pkg]
}

func TestRenderAll(t *testing.T) {
	tests := []struct {
		name     string
		renderer typing.Renderer
		op       typing.Operation
		pkgs     map[string]string
		expected string
	}{
		{
			name:     "base type",
			renderer: typing.NewNamed("", "string", nil),
			expected: "string",
		},
		{
			name:     "slice base type",
			renderer: typing.NewSlice(0, typing.NewNamed("", "string", nil)),
			expected: "[]string",
		},
		{
			name:     "array base type",
			renderer: typing.NewSlice(20, typing.NewNamed("", "string", nil)),
			expected: "[20]string",
		},
		{
			name:     "pointer[Decl] base type",
			op:       typing.DeclOp,
			renderer: typing.NewPointer(1, typing.NewNamed("", "string", nil)),
			expected: "*string",
		},
		{
			name:     "pointer[Call] base type",
			op:       typing.CallOp,
			renderer: typing.NewPointer(1, typing.NewNamed("", "string", nil)),
			expected: "&string",
		},
		{
			name:     "slice of pointers base type",
			op:       typing.DeclOp,
			renderer: typing.NewSlice(0, typing.NewPointer(1, typing.NewNamed("", "string", nil))),
			expected: "[]*string",
		},

		// Package Type
		{
			name:     "package type",
			renderer: typing.NewNamed("com/my/package", "MyStruct", nil),
			pkgs:     map[string]string{"com/my/package": "my_als"},
			expected: "my_als.MyStruct",
		},
		{
			name:     "slice package type",
			renderer: typing.NewSlice(0, typing.NewNamed("com/my/package", "MyStruct", nil)),
			pkgs:     map[string]string{"com/my/package": "my_als"},
			expected: "[]my_als.MyStruct",
		},
		{
			name:     "pointer[Call] package type",
			op:       typing.CallOp,
			renderer: typing.NewPointer(1, typing.NewNamed("com/my/package", "MyStruct", nil)),
			pkgs:     map[string]string{"com/my/package": "my_als"},
			expected: "&my_als.MyStruct",
		},
		{
			name:     "pointer[Decl] package type",
			op:       typing.DeclOp,
			renderer: typing.NewPointer(1, typing.NewNamed("com/my/package", "MyStruct", nil)),
			pkgs:     map[string]string{"com/my/package": "my_als"},
			expected: "*my_als.MyStruct",
		},
		{
			name: "package type generics",
			op:   typing.CallOp,
			renderer: typing.NewPointer(1, typing.NewNamed("com/my/package", "MyStruct", []typing.Renderer{
				typing.NewNamed("", "string", nil),
				typing.NewPointer(1, typing.NewNamed("com/other/package", "Struct", nil)),
			})),
			pkgs: map[string]string{
				"com/my/package":    "my_als",
				"com/other/package": "other_als",
			},
			expected: "&my_als.MyStruct[string, *other_als.Struct]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := &testCtx{pkgs: test.pkgs}

			actual := test.renderer.Render(ctx, test.op)
			if actual != test.expected {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}
