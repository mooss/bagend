// This file implements the Parser, which is used to both register flags and parse the arguments.

package flag

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

////////////
// Parser //

type Parser struct {
	flags         flagset
	canonical     []flag
	flagDefErrors []error
	Positional    PositionalArguments

	printHelp bool
	usage     string
}

type ParserOpt func(*Parser)

// With help automatically registers a --help|-h flag and exit with an help page when the flag is
// enabled.
func WithHelp(arg0, usage string) func(*Parser) {
	return func(cfg *Parser) {
		cfg.usage = arg0 + " " + usage
		cfg.Bool("help", &cfg.printHelp, "Print this help page").Alias("h")
	}
}

func NewParser(opts ...ParserOpt) *Parser {
	res := Parser{flags: flagset{}}
	for _, opt := range opts {
		opt(&res)
	}

	return &res
}

// Parse parses the given arguments.
// It can be called multiple times.
func (par *Parser) Parse(arguments []string) error {
	expanded, err := par.validateAndExpand()
	if err != nil {
		return err
	}

	if err := par.processArguments(arguments, expanded); err != nil {
		return err
	}

	return par.finalizeParse()
}

// validateAndExpand checks definitions and expands the flags aliases.
func (par *Parser) validateAndExpand() (flagset, error) {
	if len(par.flagDefErrors) > 0 {
		msg := fmt.Errorf("%d flag definition errors, refusing to parse", len(par.flagDefErrors))
		return nil, errors.Join(append([]error{msg}, par.flagDefErrors...)...)
	}

	expanded, errs := par.flags.expand()
	if errs != nil {
		msg := fmt.Errorf("%d flag errors after aliases expansion, refusing to parse", len(errs))
		return nil, errors.Join(append([]error{msg}, errs...)...)
	}

	return expanded, nil
}

// processArguments loops over all the arguments and fills the given flagset.
//
//nolint:revive // Can't easily lower cognitive complexity.
func (par *Parser) processArguments(arguments []string, flags flagset) error {
	var dest sink = &par.Positional

	for i, arg := range arguments {
		if !strings.HasPrefix(arg, "-") { // Value.
			if dest.full() {
				dest = &par.Positional
			}

			if err := dest.consume(arg); err != nil {
				return fmt.Errorf("when consuming %s (%s): %w", dest.names()[0], dest.kind(), err)
			}

			continue
		}

		// Flag.
		dest = flags[arg]
		switch {
		case dest == nil:
			return fmt.Errorf("unknown flag: %s", arg)
		case is[*singletonflag[bool, Bool]](dest):
			dest.consume("true") //nolint:errcheck // Cannot fail.
		case i == len(arguments)-1: // No more arguments to consume.
			return fmt.Errorf("flag %s requires a value but none was provided", arg)
		}
	}

	return nil
}

// finalizeParse handles the help page and enforce default values.
func (par *Parser) finalizeParse() error {
	if par.printHelp {
		fmt.Print(par.Help())
		os.Exit(0)
	}

	for _, flg := range par.canonical {
		flg.enforceDefault()
	}

	return nil
}

/////////////
// flagset //

type flagset map[string]flag

// add adds a new key-value association to the flagset, returning an error if fullname is invalid
// or if it already exists.
// flagname is the full flag name, that is to say it is prefixed by `-` or `--`.
func (fs flagset) add(flagname string, value flag) error {
	if len(flagname) == 0 {
		return fmt.Errorf("%s (names: %v) has empty names", value.kind(), value.names())
	}

	if _, exists := fs[flagname]; exists {
		return fmt.Errorf("flag %s already exists", flagname)
	}

	fs[flagname] = value
	return nil
}

// expand returns a new flagset containing all the original flags and the aliases, as well as all
// errors that occurred during the process.
func (fs flagset) expand() (flagset, []error) {
	var errs []error
	res := flagset{}

	for canonical, flg := range fs {
		res[canonical] = flg

		for _, alias := range flg.names()[1:] {
			if err := res.add(name2flag(alias), flg); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return res, errs
}

///////////////
// Utilities //

func (par *Parser) errdef(err error) {
	par.flagDefErrors = append(par.flagDefErrors, err)
}

func (par *Parser) registerflag(flg flag) {
	if err := par.flags.add(name2flag(flg.names()[0]), flg); err != nil {
		par.errdef(err)
	}

	par.canonical = append(par.canonical, flg)
}

func name2flag(name string) string {
	switch len(name) {
	case 0:
		return ""
	case 1:
		return "-" + name
	default:
		return "--" + name
	}
}

func is[T any](value any) bool {
	_, res := value.(T)
	return res
}
