package parse

import (
	"fmt"
	"go/token"
)

func locatedErr(fileSet *token.FileSet, pos token.Pos, text string, a ...any) error {
	ps := fileSet.Position(pos)

	formatted := fmt.Sprintf(text, a...)
	return fmt.Errorf("%s:%d: %s", ps.Filename, ps.Line, formatted)
}
