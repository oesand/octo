package validate

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

// Regex returns a validator that checks the given string matches the provided
// regular expression.
//
// It panics if `regex` is nil.
func Regex(regex *regexp.Regexp) Validator[string] {
	if regex == nil {
		panic("octo: regex is nil")
	}
	return &stringRegexValidator{regex}
}

type stringRegexValidator struct {
	regex *regexp.Regexp
}

func (validator *stringRegexValidator) Validate(value string) Errors {
	if !validator.regex.MatchString(value) {
		return []string{"mismatch expected pattern"}
	}
	return nil
}

// RunesExactly returns a validator that ensures the string contains exactly
// `length` runes (Unicode code points).
func RunesExactly(length int) Validator[string] {
	return &stringLengthValidator{length}
}

type stringLengthValidator struct {
	length int
}

func (validator *stringLengthValidator) Validate(value string) Errors {
	if utf8.RuneCountInString(value) != validator.length {
		return []string{fmt.Sprintf("must have exactly %d characters", validator.length)}
	}
	return nil
}

// MinRunes returns a validator that ensures the string contains at least
// `min` runes (Unicode code points).
func MinRunes(min int) Validator[string] {
	return &stringMinValidator{min}
}

type stringMinValidator struct {
	min int
}

func (validator *stringMinValidator) Validate(value string) Errors {
	if utf8.RuneCountInString(value) < validator.min {
		return []string{fmt.Sprintf("must have at least %d characters", validator.min)}
	}
	return nil
}

// MaxRunes returns a validator that ensures the string contains at most
// `max` runes (Unicode code points).
func MaxRunes(max int) Validator[string] {
	return &stringMaxValidator{max}
}

type stringMaxValidator struct {
	max int
}

func (validator *stringMaxValidator) Validate(value string) Errors {
	if utf8.RuneCountInString(value) > validator.max {
		return []string{fmt.Sprintf("must have at most %d characters", validator.max)}
	}
	return nil
}
