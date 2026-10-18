package flag

import (
	"fmt"
	"strings"

	"github.com/mooss/bagend/go/fun/eager/lie"
)

// Help returns a formatted help string showing all registered flags and their documentation.
func (par *Parser) Help() string {
	align := 0
	type helpline struct {
		decl string
		doc  string
	}
	declarations := lie.Map(func(flg flag) helpline {
		res := helpline{
			decl: strings.Join(lie.Map(name2flag, flg.names()), ", "),
			doc:  flg.docline(),
		}

		align = max(align, len(res.decl))
		return res
	}, par.canonical)

	format := fmt.Sprintf("  %%-%ds  %%s", align)
	return strings.Join(lie.Map(func(h helpline) string {
		return fmt.Sprintf(format, h.decl, h.doc)
	}, declarations), "\n")
}
