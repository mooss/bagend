// Package flag provides utilities to define CLI flags and parse arguments.
package flag

// This file defines the following interfaces:
// - FluentFlags, used to provide additional options to flags after declaration.
// - flag, used by the parser to register and manipulate declared flags.
//
// With those 2 interfaces, adding support to parse any new type is simply a matter of implementing
// a Decoder.

////////////////
// Interfaces //
////////////////

// sink is the interface used by the parser to process arguments.
// It is common to flags and positional arguments.
type sink interface {
	// consume takes a value and attempts to decode and store it.
	consume(string) error
	// full returns true when the sink has consumed all the values it can.
	full() bool
	// kind returns a short explanation of kind kind of sink this is (singleton, slice or positional
	// argument).
	kind() string
}

// flag is the interface all registered flags must implement.
type flag interface {
	sink

	// Since sink is common to flags and positional arguments, it follows that the following methods
	// are the essence of a flag.

	// names returns the names of the flag (the first is the canonical names, the rest are aliases).
	names() []string
	// docline returns the documentation line of the flag.
	docline() string
	// enforceDefault assigns the default value if the flag value has not been set.
	enforceDefault()
}

// FluentFlag is the interface that is used for additional configuration of registered flags.
type FluentFlag[T any] interface {
	// Alias register its arguments as aliases.
	Alias(...string) FluentFlag[T]
	// Default sets the given value as the default.
	Default(T) FluentFlag[T]
}

//////////////
// flagBase //
//////////////

type flagBase[T any] struct {
	def        T
	dest       *T
	docLine    string
	namesStore []string
	alreadySet bool
}

/////////////////////////////////////////
// FluentFlag interface implementation //

func (fai *flagBase[T]) Alias(aliases ...string) FluentFlag[T] {
	fai.namesStore = append(fai.namesStore, aliases...)
	return fai
}

func (fai *flagBase[T]) Default(value T) FluentFlag[T] {
	fai.def = value
	return fai
}

///////////////////////////////////////////
// Part of flag interface implementation //

func (fai flagBase[T]) names() []string {
	return fai.namesStore
}
func (fai flagBase[T]) docline() string {
	return fai.docLine
}

func (ffs *flagBase[T]) enforceDefault() {
	if !ffs.alreadySet {
		*ffs.dest = ffs.def
	}
}

//////////////////////////
// Positional arguments //
//////////////////////////

// PositionalArguments implements the subset of flags methods required to consume arguments.
type PositionalArguments []string

func (pa *PositionalArguments) consume(value string) error {
	*pa = append(*pa, value)
	return nil
}

func (pa *PositionalArguments) full() bool   { return false }
func (pa *PositionalArguments) kind() string { return "positional argument" }
