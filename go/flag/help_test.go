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
			`  --intflag  integer flag
  --strflag  string flag`,
		},
		{
			"with aliases",
			func(par *Parser) {
				par.Bool("boolflag", new(bool), "boolean flag").Alias("b", "bool")
			},
			`  --boolflag, -b, --bool  boolean flag`,
		},
		{
			"with usage",
			func(par *Parser) {
				EnableHelp()(par)
				par.Int("intflag", new(int), "integer flag")
			},
			`  --help, -h  Print this help page
  --intflag   integer flag`,
		},
		{
			"empty parser",
			func(*Parser) {},
			``,
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
