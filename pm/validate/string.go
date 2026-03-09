package validate

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

// Regex creates a condition that validates a string matches the given regular expression.
func Regex(regex *regexp.Regexp) Validator[string] {
	if regex == nil {
		panic("octo: regex is nil")
	}
	return &stringRegexValidator{regex}
}

type stringRegexValidator struct {
	regex *regexp.Regexp
}

func (c *stringRegexValidator) Validate(value string) []string {
	if !c.regex.MatchString(value) {
		return []string{"mismatch expected pattern"}
	}
	return nil
}

// RunesExactly creates a condition that validates a string has exactly the specified length.
func RunesExactly(length int) Validator[string] {
	return &stringLengthValidator{length}
}

type stringLengthValidator struct {
	length int
}

func (c *stringLengthValidator) Validate(value string) []string {
	if utf8.RuneCountInString(value) != c.length {
		return []string{fmt.Sprintf("must have exactly %d characters", c.length)}
	}
	return nil
}

// MinRunes creates a condition that validates a string has at least the specified runes.
func MinRunes(min int) Validator[string] {
	return &stringMinValidator{min}
}

type stringMinValidator struct {
	min int
}

func (c *stringMinValidator) Validate(value string) []string {
	if utf8.RuneCountInString(value) < c.min {
		return []string{fmt.Sprintf("must have at least %d characters", c.min)}
	}
	return nil
}

// MaxRunes creates a condition that validates a string has at most the specified length.
func MaxRunes(max int) Validator[string] {
	return &stringMaxValidator{max}
}

type stringMaxValidator struct {
	max int
}

func (c *stringMaxValidator) Validate(value string) []string {
	if utf8.RuneCountInString(value) > c.max {
		return []string{fmt.Sprintf("must have at most %d characters", c.max)}
	}
	return nil
}
