package flag

import (
	"testing"
)

func TestParser_Help(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Parser)
		expected string
	}{
		{
			"basic flags",
			func(par *Parser) {
				par.Int("intflag", new(int), "integer flag")
				par.String("strflag", new(string), "string flag")
			},
			`Usage: 

Flags:
  --intflag  integer flag
  --strflag  string flag
`,
		},
		{
			"with aliases",
			func(par *Parser) {
				par.Bool("boolflag", new(bool), "boolean flag").Alias("b", "bool")
			},
			`Usage: 

Flags:
  --boolflag, -b, --bool  boolean flag
`,
		},
		{
			"with usage",
			func(par *Parser) {
				WithHelp("testprog", "[options]")(par)
				par.Int("intflag", new(int), "integer flag")
			},
			`Usage: testprog [options]

Flags:
  --help, -h  Print this help page
  --intflag   integer flag
`,
		},
		{
			"empty parser",
			func(*Parser) {},
			`Usage: 

Flags:
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			par := NewParser()
			tt.setup(par)
			help := par.Help()
			if help != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, help)
			}
		})
	}
}
