// This file implements the Parser, which is used to both register flags and parse the arguments.

package flag

import (
	"errors"
	"fmt"
	"strings"
)

////////////
// Parser //

type Parser struct {
	flags         flagset
	canonical     []flag
	flagDefErrors []error
	Positional    PositionalArguments
}

func NewParser() *Parser {
	return &Parser{flags: flagset{}}
}

// Parse parses the given arguments.
// It must be called only once.
func (par *Parser) Parse(arguments []string) error {
	if len(par.flagDefErrors) > 0 {
		msg := fmt.Errorf("%d flag definition errors, refusing to parse", len(par.flagDefErrors))
		return errors.Join(append([]error{msg}, par.flagDefErrors...)...)
	}

	allFlags, errs := par.flags.expand()
	if errs != nil {
		msg := fmt.Errorf("%d flag errors after aliases expansion, refusing to parse", len(errs))
		return errors.Join(append([]error{msg}, errs...)...)
	}

	var sink interface {
		consume(string) error
		full() bool
		what() string
	}
	sink = &par.Positional
	lastFlag := ""

	for i, arg := range arguments {
		if strings.HasPrefix(arg, "-") {
			var known bool
			sink, known = allFlags[arg]
			if !known {
				return fmt.Errorf("unknown flag: %s", arg)
			}

			// Handle boolean flags.
			if is[*singletonflag[bool, Bool]](sink) {
				if err := sink.consume("true"); err != nil {
					return fmt.Errorf("when consuming %s (%s): %w", arg, sink.what(), err)
				}

				sink = &par.Positional
				continue
			}

			// Non-boolean flags need an associated value.
			if i == len(arguments)-1 {
				return fmt.Errorf("flag %s requires a value but none was provided", arg)
			}

			lastFlag = arg
			continue
		}

		if sink.full() {
			sink = &par.Positional
		}

		if err := sink.consume(arg); err != nil {
			// consume cannot fail on positional arguments, this error can only be triggered by
			// flags with decoders who can return an error.
			return fmt.Errorf("when consuming %s (%s): %w", lastFlag, sink.what(), err)
		}
	}

	for _, flg := range par.canonical {
		flg.enforceDefault()
	}

	return nil
}

//////////////////
// Registration //

// Register registers a singleton flag to a parser.
// Registering different flags to the same destination is undefined behavior.
func Register[D Decoder[T], T any](par *Parser, name string, dest *T, docline string) FluentFlag[T] {
	flg := singletonflag[T, D]{
		flagBase[T]{
			dest:       dest,
			docLine:    docline,
			namesStore: []string{name},
		},
	}

	par.registerflag(&flg)

	return &flg
}

// RegisterSlice registers a slice flag to a parser.
// Registering different flags to the same destination is undefined behavior.
func RegisterSlice[Dec Decoder[T], T any](par *Parser, name string, dest *[]T, docline string) FluentFlag[[]T] {
	flg := sliceFlag[T, Dec]{
		flagBase[[]T]{
			dest:       dest,
			docLine:    docline,
			namesStore: []string{name},
		},
	}

	par.registerflag(&flg)

	return &flg
}

/////////////
// flagset //

type flagset map[string]flag

// add adds a new key-value association to the flagset, returning an error if fullname is invalid
// or if it already exists.
// flagname is the full flag name, that is to say it is prefixed by `-` or `--`.
func (fs flagset) add(flagname string, value flag) error {
	if len(flagname) == 0 {
		return fmt.Errorf("%s (names: %v) has empty names", value.what(), value.names())
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
