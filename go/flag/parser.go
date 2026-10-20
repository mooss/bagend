// This file implements the Parser, which is used to both register flags and parse the arguments.

package flag

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mooss/bagend/go/errs"
)

////////////
// Parser //

type Parser struct {
	flags         flagset
	canonical     []flag
	flagDefErrors []error
	parseCalled   bool
	Positional    PositionalArguments

	printHelp       bool
	customHelpPage  func() string
	allowPositional bool
}

type ParserOpt func(*Parser)

func NewParser(opts ...ParserOpt) *Parser {
	res := Parser{flags: flagset{}}
	for _, opt := range opts {
		opt(&res)
	}

	return &res
}

func (par *Parser) AllowsPositionalArguments() bool {
	return par.allowPositional
}

// SetHelpPage overrides the help page printed when the help flag is enabled.
func (par *Parser) SetHelpPage(helpPage func() string) {
	par.customHelpPage = helpPage
}

// Parse parses the given arguments.
// It may only be called once.
func (par *Parser) Parse(arguments []string) error {
	if par.parseCalled {
		return errors.New("parser may only be parsed once")
	}
	par.parseCalled = true

	_, _, err := par.ParseUntil(arguments, nil)
	return err
}

// ParseUntil parses arguments until stop matches a positional argument, if any.
// It returns the unconsumed arguments and whether parsing stopped early.
func (par *Parser) ParseUntil(arguments []string, stop func(string) bool) ([]string, bool, error) {
	expanded, err := par.validateAndExpand()
	if err != nil {
		return nil, false, err
	}

	remaining, stopped, err := par.processArguments(arguments, expanded, stop)
	if err != nil {
		return nil, false, err
	}

	if err := par.finalizeParse(); err != nil {
		return nil, false, err
	}

	return remaining, stopped, nil
}

// validateAndExpand checks definitions and expands the flags aliases.
func (par *Parser) validateAndExpand() (flagset, error) {
	if len(par.flagDefErrors) > 0 {
		msg := fmt.Errorf("%d flag definition errors, refusing to parse", len(par.flagDefErrors))
		return nil, errors.Join(append([]error{msg}, par.flagDefErrors...)...)
	}

	expanded, problems := par.flags.expand()
	if problems != nil {
		msg := "%d flag errors after aliases expansion, refusing to parse"
		err := fmt.Errorf(msg, len(problems))
		return nil, errors.Join(append([]error{err}, problems...)...)
	}

	return expanded, nil
}

// isSliceTerminator reports whether arg ends the current slice consumption.
func isSliceTerminator(arg string, dest sink, remaining int) bool {
	return arg == "--" && remaining != 0 && dest.arity() == -1 && !is[*PositionalArguments](dest)
}

// processArguments processes arguments using the provided flag set.
//
// arguments contains the command-line arguments to process.
// flags maps flag names to their destinations.
// stop (can be nil) determines whether a positional argument should stop parsing.
//
// Returns the unprocessed arguments and true when stop requests an end to parsing.
// Otherwise, it returns nil and false.
// If an argument cannot be processed, it returns an error.
//
//nolint:revive // Can't easily lower cognitive complexity.
func (par *Parser) processArguments(
	arguments []string, flags flagset, stop func(string) bool,
) ([]string, bool, error) {
	var dest sink = &par.Positional
	remaining := -1

	for i, arg := range arguments {
		if isSliceTerminator(arg, dest, remaining) {
			// End of slice, resuming positional/subcommand.
			dest = &par.Positional
			remaining = 0
			continue
		}

		if !strings.HasPrefix(arg, "-") { // Value.
			if remaining == 0 {
				dest = &par.Positional
			}
			if is[*PositionalArguments](dest) && stop != nil && stop(arg) {
				return arguments[i:], true, nil
			}
			if is[*PositionalArguments](dest) && !par.allowPositional {
				return nil, false, fmt.Errorf("positional arguments are not allowed (got %q)", arg)
			}

			if err := dest.consume(arg); err != nil {
				err = errs.Wrapf(err, "when consuming %s", dest.prettyName())
				return nil, false, err
			}

			remaining--
			continue
		}

		// Flag.
		dest = flags[arg]
		switch {
		case dest == nil:
			return nil, false, fmt.Errorf("unknown flag: %s", arg)
		case is[*singletonFlag[bool, Bool]](dest):
			dest.consume("true") //nolint:errcheck // Cannot fail.
			remaining = 0
		default:
			remaining = dest.arity()
		}
	}

	if remaining > 0 {
		return nil, false, fmt.Errorf("flag `%s` requires a value", name2flag(dest.names()[0]))
	}

	return nil, false, nil
}

// finalizeParse handles the help page and enforce default values.
func (par *Parser) finalizeParse() error {
	if par.printHelp {
		var help string
		if par.customHelpPage == nil {
			help = par.Help()
		} else {
			help = par.customHelpPage()
		}

		fmt.Print(help)
		os.Exit(0)
	}

	for _, flg := range par.canonical {
		flg.enforceDefault()
	}

	return nil
}

////////////////////
// Parser options //

// WithPositionalArguments enables positional arguments for the parser.
func WithPositionalArguments() ParserOpt {
	return func(cfg *Parser) {
		cfg.allowPositional = true
	}
}

// WithPositionalArguments enables positional arguments for the parser.
func (par *Parser) WithPositionalArguments() *Parser {
	WithPositionalArguments()(par)
	return par
}

// EnableHelp registers a --help|-h flag that prints the help page when enabled.
// Does nothing when a help flag has already been defined.
func EnableHelp() ParserOpt {
	return func(cfg *Parser) {
		if _, exists := cfg.flags[name2flag("help")]; exists {
			return
		}
		cfg.Bool("help", &cfg.printHelp, "Print this help page").Alias("h")
	}
}

// EnableHelp registers a --help|-h flag that prints the help page when enabled.
// Does nothing when a help flag has already been defined.
func (par *Parser) EnableHelp() *Parser {
	EnableHelp()(par)
	return par
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
	var problems []error
	res := flagset{}

	for canonical, flg := range fs {
		res[canonical] = flg

		for _, alias := range flg.names()[1:] {
			if err := res.add(name2flag(alias), flg); err != nil {
				problems = append(problems, err)
			}
		}
	}

	return res, problems
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
