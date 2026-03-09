package pm

type Caser interface {
	Convert(string) string
}

var (
	// CamelCaser converts a string to camelCase.
	CamelCaser Caser = new(camelCaser)

	// KebabCaser converts a string to kebab-case.
	KebabCaser Caser = new(kebabCaser)

	// PascalCaser converts a string to PascalCase.
	PascalCaser Caser = new(pascalCaser)

	// SnakeCaser converts a string to snake_case.
	SnakeCaser Caser = new(snakeCaser)
)

type camelCaser struct{}

func (*camelCaser) Convert(input string) string {
	if input == "" {
		return ""
	}

	runes := []rune(input)
	length := len(runes)
	result := make([]rune, 0, length)

	upperNext := false
	for i, r := range runes {
		if r == '_' || r == '-' || r == ' ' {
			upperNext = true
		} else {
			if upperNext {
				result = append(result, toAsciiUpperCase(r))
				upperNext = false
			} else {
				if i == 0 {
					result = append(result, toAsciiLowerCase(r))
				} else {
					result = append(result, r)
				}
			}
		}
	}

	return string(result)
}

type kebabCaser struct{}

func (*kebabCaser) Convert(input string) string {
	if input == "" {
		return ""
	}

	runes := []rune(input)
	length := len(runes)
	result := make([]rune, 0, length)

	for i, r := range runes {
		if isAsciiUpperCase(r) {
			if i > 0 {
				result = append(result, '-')
			}
			result = append(result, toAsciiLowerCase(r))
		} else if isAsciiLowerCase(r) || isAsciiDigit(r) {
			result = append(result, r)
		} else {
			if len(result) > 0 && result[len(result)-1] != '-' {
				result = append(result, '-')
			}
		}
	}

	// Remove trailing hyphen if exists
	if len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}

	return string(result)
}

type pascalCaser struct{}

func (*pascalCaser) Convert(input string) string {
	if input == "" {
		return ""
	}

	runes := []rune(input)
	length := len(runes)
	result := make([]rune, 0, length)

	upperNext := true
	for _, r := range runes {
		if r == '_' || r == '-' || r == ' ' {
			upperNext = true
		} else {
			if upperNext {
				result = append(result, toAsciiUpperCase(r))
				upperNext = false
			} else {
				result = append(result, r)
			}
		}
	}

	return string(result)
}

type snakeCaser struct{}

func (*snakeCaser) Convert(input string) string {
	if input == "" {
		return ""
	}

	runes := []rune(input)
	length := len(runes)
	result := make([]rune, 0, length)

	for i, r := range runes {
		if isAsciiUpperCase(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, toAsciiLowerCase(r))
		} else if isAsciiLowerCase(r) || isAsciiDigit(r) {
			result = append(result, r)
		} else {
			if len(result) > 0 && result[len(result)-1] != '_' {
				result = append(result, '_')
			}
		}
	}

	// Remove trailing underscore if exists
	if len(result) > 0 && result[len(result)-1] == '_' {
		result = result[:len(result)-1]
	}

	return string(result)
}

func isAsciiUpperCase(r rune) bool {
	return 0x41 <= r && r <= 0x5a
}

func isAsciiLowerCase(r rune) bool {
	return 0x61 <= r && r <= 0x7a
}

func isAsciiDigit(r rune) bool {
	return 0x30 <= r && r <= 0x39
}

func toAsciiUpperCase(r rune) rune {
	if isAsciiLowerCase(r) {
		return r - 0x20
	}
	return r
}

func toAsciiLowerCase(r rune) rune {
	if isAsciiUpperCase(r) {
		return r + 0x20
	}
	return r
}
