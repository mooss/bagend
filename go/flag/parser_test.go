package flag

import (
	"reflect"
	"testing"
)

// noErr asserts that the given error is nil.
func noErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// yesErr asserts that the given error is not nil.
func yesErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

// eq asserts that expected equal to got.
func eq(t *testing.T, expected any, got any) {
	t.Helper()
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestName2Flag(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"", ""},
		{"a", "-a"},
		{"abc", "--abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := name2flag(tt.name); got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestParser_BasicParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"no flags", []string{}, false},
		{"unknown flag", []string{"-a"}, true},
		{"empty flag name", []string{"-"}, true},
		{"double dash", []string{"--"}, true},
		{"flag without value", []string{"--flag"}, true}, // TODO: fix parser to fail here.
		{"multiple flags", []string{"--flag1", "1", "--flag2", "2"}, false},
		{"mixed flags and positional", []string{"--flag", "1", "pos1", "pos2"}, false},
		{"flag after positional", []string{"pos1", "--flag", "1"}, false},
		{"multiple same flag", []string{"--flag", "1", "--flag", "2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i int
			par := NewParser()

			// We don't care about the flags.
			Register[Int](par, "flag", &i, "test flag")
			Register[Int](par, "flag1", &i, "test flag1")
			Register[Int](par, "flag2", &i, "test flag2")

			err := par.Parse(tt.args)
			if tt.expectError {
				yesErr(t, err)
			} else {
				noErr(t, err)
			}
		})
	}
}

func TestParser_FlagRegistration(t *testing.T) {
	t.Run("duplicate flag", func(t *testing.T) {
		par := NewParser()
		var i int
		Register[Int](par, "flag", &i, "test flag")
		Register[Int](par, "flag", &i, "test flag")
		yesErr(t, par.Parse(nil))
	})

	t.Run("empty flag name", func(t *testing.T) {
		par := NewParser()
		var i int
		Register[Int](par, "", &i, "test flag")
		yesErr(t, par.Parse(nil))
	})

	t.Run("multiple aliases", func(t *testing.T) {
		par := NewParser()
		var i int
		Register[Int](par, "flag", &i, "test flag").Alias("f", "fl")
		noErr(t, par.Parse([]string{"-f", "1"}))
		eq(t, 1, i)
	})
}

func flagExpect[D Decoder[T], T any](t *testing.T, name string, args []string, expected T) {
	t.Helper()
	var dest T
	par := NewParser()

	Register[D](par, name, &dest, name)
	noErr(t, par.Parse(args))
	eq(t, expected, dest)
}

func flagExpectS[D Decoder[T], T any](t *testing.T, name string, args []string, expected []T) {
	var dest []T
	par := NewParser()

	RegisterSlice[D](par, name, &dest, name)
	noErr(t, par.Parse(args))
	eq(t, expected, dest)
}

func TestParser_FlagTypes(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectValue any
	}{
		{
			name:        "int_flag",
			args:        []string{"--int_flag", "123"},
			expectValue: 123,
		},
		{
			name:        "string_flag",
			args:        []string{"--string_flag", "hello"},
			expectValue: "hello",
		},
		{
			name:        "int_slice_flag",
			args:        []string{"--int_slice_flag", "1", "--int_slice_flag", "2"},
			expectValue: []int{1, 2},
		},
		{
			name:        "string_slice_flag",
			args:        []string{"--string_slice_flag", "hello", "--string_slice_flag", "world"},
			expectValue: []string{"hello", "world"},
		},
		{
			name:        "zero_value",
			args:        []string{},
			expectValue: 0,
		},
		{
			name:        "invalid_int",
			args:        []string{"--int_flag", "abc"},
			expectValue: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch concrete := tt.expectValue.(type) {
			case int:
				if tt.name == "invalid_int" {
					par := NewParser()
					var i int
					Register[Int](par, "int_flag", &i, "test flag")
					yesErr(t, par.Parse(tt.args))
				} else {
					flagExpect[Int](t, tt.name, tt.args, concrete)
				}
			case string:
				flagExpect[String](t, tt.name, tt.args, concrete)
			case []int:
				flagExpectS[Int](t, tt.name, tt.args, concrete)
			case []string:
				flagExpectS[String](t, tt.name, tt.args, concrete)
			}
		})
	}
}

func TestParser_PositionalArguments(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "no flags",
			args:     []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "mixed flags and positional",
			args:     []string{"--flag", "23", "a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "positional after flags",
			args:     []string{"--flag", "23", "a", "--flag2", "42", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "only positional",
			args:     []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			par := NewParser()
			var i int
			Register[Int](par, "flag", &i, "test flag")
			Register[Int](par, "flag2", &i, "test flag2")
			noErr(t, par.Parse(tt.args))
			eq(t, tt.expected, []string(par.Positional))
		})
	}
}

func TestParser_DefaultValues(t *testing.T) {
	t.Run("int default", func(t *testing.T) {
		par := NewParser()
		var i int = 42
		Register[Int](par, "flag", &i, "test flag").Default(23)
		noErr(t, par.Parse([]string{}))
		eq(t, 23, i)
	})

	t.Run("string default", func(t *testing.T) {
		par := NewParser()
		var s string = "foo"
		Register[String](par, "flag", &s, "test flag").Default("bar")
		noErr(t, par.Parse([]string{}))
		eq(t, "bar", s)
	})

	t.Run("slice default", func(t *testing.T) {
		par := NewParser()
		var s []int
		RegisterSlice[Int](par, "flag", &s, "test flag").Default([]int{1, 2, 3})
		noErr(t, par.Parse([]string{}))
		eq(t, []int{1, 2, 3}, s)
	})
}
