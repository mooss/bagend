package flag

import (
	"fmt"
	"strings"

	"github.com/mooss/bagend/go/fun/eager/lie"
)

// Help returns a formatted help string showing all registered flags and their documentation.
func (par *Parser) Help() string {
	align := 0
	mkdecl := func(flg flag) string {
		res := strings.Join(lie.Map(name2flag, flg.names()), ", ")
		align = max(align, len(res))
		return res
	}
	declarations := lie.Map(mkdecl, par.canonical)

	format := fmt.Sprintf("  %%-%ds  %%s\n", align)
	lines := make([]string, len(declarations))
	for i, decl := range declarations {
		lines[i] = fmt.Sprintf(format, decl, par.canonical[i].docline())
	}

	return strings.Join(lines, "")
}
