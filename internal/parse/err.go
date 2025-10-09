package parse

import (
	"errors"
	"fmt"
	"go/token"
)

func locatedMsg(fileSet *token.FileSet, pos token.Pos, format string, a ...any) string {
	ps := fileSet.Position(pos)

	formatted := fmt.Sprintf(format, a...)
	return fmt.Sprintf("%s:%d: %s", ps.Filename, ps.Line, formatted)
}

func locatedErr(fileSet *token.FileSet, pos token.Pos, format string, a ...any) error {
	return errors.New(locatedMsg(fileSet, pos, format, a...))
}
