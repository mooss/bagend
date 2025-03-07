package flag

import (
	"reflect"
	"testing"
)

// noErr asserts that the given error is nil.
func noErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// yesErr asserts that the given error is not nil.
func yesErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			par := NewParser()
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
}

func flagExpect[D Decoder[T], T any](t *testing.T, name string, args []string, expected T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch concrete := tt.expectValue.(type) {
			case int:
				flagExpect[Int](t, tt.name, tt.args, concrete)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			par := NewParser()
			var i int
			Register[Int](par, "flag", &i, "test flag")
			noErr(t, par.Parse(tt.args))
			eq(t, tt.expected, []string(par.Positional))
		})
	}
}
