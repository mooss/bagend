// Package flag provides utilities to define CLI flags and parse arguments.
package flag

////////////////
// Interfaces //
////////////////

// sink is the interface used by the parser to process arguments.
// It is common to flags and positional arguments.
type sink interface {
	// consume takes a value and attempts to decode and store it.
	consume(string) error

	// arity returns the number of values the sink *must* consume.
	arity() int

	// kind returns a short explanation of the kind of sink this is (singleton, slice or positional
	// argument), with type information if pertinent.
	kind() string

	// names returns the names of the flag (the first is the canonical names, the rest are aliases).
	names() []string

	// prettyName returns the human-readable name of the sink.
	prettyName() string
}

// flag is the interface all registered flags must implement.
type flag interface {
	sink

	// docline returns the documentation line of the flag.
	docline() string

	// enforceDefault assigns the default value if the flag value has not been set.
	enforceDefault()
}

// BooleanFlag is the public interface to configure boolean flags.
type BooleanFlag[T any] interface {
	// Alias registers its arguments as aliases.
	Alias(...string) BooleanFlag[T]

	// Default sets the given value as the default.
	Default(T) BooleanFlag[T]
}

// SingletonFlag is the public interface to configure singleton flag.
type SingletonFlag[T any] interface {
	// Alias registers its arguments as aliases.
	Alias(...string) SingletonFlag[T]

	// Default sets the given value as the default.
	Default(T) SingletonFlag[T]

	// Allowed restricts the flag to the given values.
	Allowed(...T) SingletonFlag[T]

	// AllowedFun sets a function that determines whether a value is accepted.
	AllowedFun(func(T) bool) SingletonFlag[T]
}

// SliceFlag is the public interface to configure slice flags.
type SliceFlag[T any] interface {
	// Alias registers its arguments as aliases.
	Alias(...string) SliceFlag[T]

	// Default sets the given value as the default.
	Default([]T) SliceFlag[T]

	// Allowed restricts each slice element to the given values.
	Allowed(...T) SliceFlag[T]

	// AllowedFun sets a function that determines whether each element is accepted.
	AllowedFun(func(T) bool) SliceFlag[T]
}

//////////////////////////
// Positional arguments //
//////////////////////////

// PositionalArguments implements the subset of flags methods required to consume arguments.
type PositionalArguments []string

func (*PositionalArguments) arity() int { return -1 }

func (pa *PositionalArguments) consume(value string) error {
	*pa = append(*pa, value)
	return nil
}

func (*PositionalArguments) kind() string       { return "positional argument" }
func (*PositionalArguments) names() []string    { return []string{"POSARGS"} }
func (*PositionalArguments) prettyName() string { return "POSITIONAL" }
