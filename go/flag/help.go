package flag

import (
	"fmt"
	"strings"

	"github.com/mooss/bagend/go/fun/eager/lie"
)

// Help returns a formatted help string showing all registered flags and their documentation.
func (par *Parser) Help() string {
	var builder strings.Builder

	///////////
	// Usage //

	builder.WriteString("Usage: ")
	builder.WriteString(par.usage)

	///////////
	// Flags //

	builder.WriteString("\n\nFlags:\n")

	align := 0
	mkdecl := func(flg flag) string {
		res := strings.Join(lie.Map(name2flag, flg.names()), ", ")
		align = max(align, len(res))
		return res
	}
	declarations := lie.Map(mkdecl, par.canonical)

	// Format with proper alignment.
	format := fmt.Sprintf("  %%-%ds  %%s\n", align)
	for i, decl := range declarations {
		builder.WriteString(fmt.Sprintf(format, decl, par.canonical[i].docline()))
	}

	return builder.String()
}
