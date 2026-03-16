package pm

import "testing"

func TestCamelCaser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single word",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "snake_case",
			input:    "hello_world_test",
			expected: "helloWorldTest",
		},
		{
			name:     "kebab-case",
			input:    "hello-world-test",
			expected: "helloWorldTest",
		},
		{
			name:     "PascalCase",
			input:    "HelloWorldTest",
			expected: "helloWorldTest",
		},
		{
			name:     "with numbers",
			input:    "hello_world_123",
			expected: "helloWorld123",
		},
		{
			name:     "multiple separators",
			input:    "hello___world__test",
			expected: "helloWorldTest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CamelCaser.Convert(tt.input)
			if result != tt.expected {
				t.Errorf("CamelCaser.Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestKebabCaser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single word", "hello", "hello"},
		{"camelCase", "helloWorldTest", "hello-world-test"},
		{"snake_case", "hello_world_test", "hello-world-test"},
		{"PascalCase", "HelloWorldTest", "hello-world-test"},
		{"numbers", "helloWorld123", "hello-world123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := KebabCaser.Convert(tt.input)
			if result != tt.expected {
				t.Errorf("KebabCaser.Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPascalCaser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single word", "hello", "Hello"},
		{"camelCase", "helloWorldTest", "HelloWorldTest"},
		{"snake_case", "hello_world_test", "HelloWorldTest"},
		{"kebab-case", "hello-world-test", "HelloWorldTest"},
		{"mixed", "hello_world-test", "HelloWorldTest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PascalCaser.Convert(tt.input)
			if result != tt.expected {
				t.Errorf("PascalCaser.Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSnakeCaser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single word", "hello", "hello"},
		{"camelCase", "helloWorldTest", "hello_world_test"},
		{"kebab-case", "hello-world-test", "hello_world_test"},
		{"PascalCase", "HelloWorldTest", "hello_world_test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SnakeCaser.Convert(tt.input)
			if result != tt.expected {
				t.Errorf("SnakeCaser.Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkCamelCaser(b *testing.B) {
	input := "hello_world_test_case_with_long_string"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CamelCaser.Convert(input)
	}
}

func BenchmarkKebabCaser(b *testing.B) {
	input := "helloWorldTestCase"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		KebabCaser.Convert(input)
	}
}
