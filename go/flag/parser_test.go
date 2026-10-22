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
		{"double dash", []string{"--"}, false},
		{"double dash followed by positional", []string{"--", "value"}, false},
		{"flag without value", []string{"--flag"}, true},
		{"multiple flags", []string{"--flag1", "1", "--flag2", "2"}, false},
		{"mixed flags and positional", []string{"--flag", "1", "pos1", "pos2"}, false},
		{"flag after positional", []string{"pos1", "--flag", "1"}, false},
		{"multiple same flag", []string{"--flag", "1", "--flag", "2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i int
			var positional []string
			par := NewParser().Positional(&positional)

			// We don't care about where the flags go.
			par.Int("flag", &i, "test flag")
			par.Int("flag1", &i, "test flag1")
			par.Int("flag2", &i, "test flag2")

			err := par.Parse(tt.args)
			if tt.expectError {
				yesErr(t, err)
			} else {
				noErr(t, err)
			}
		})
	}
}

func TestParser_ParseOnlyOnce(t *testing.T) {
	t.Run("repeated after success", func(t *testing.T) {
		par := NewParser()
		noErr(t, par.Parse(nil))
		yesErr(t, par.Parse(nil))
	})

	t.Run("repeated after failure", func(t *testing.T) {
		par := NewParser()
		yesErr(t, par.Parse([]string{"--unknown"}))
		yesErr(t, par.Parse(nil))
	})
}

func TestParser_FlagRegistration(t *testing.T) {
	t.Run("duplicate flag", func(t *testing.T) {
		par := NewParser()
		var i int
		par.Int("flag", &i, "test flag")
		par.Int("flag", &i, "test flag")
		yesErr(t, par.Parse(nil))
	})

	t.Run("empty flag name", func(t *testing.T) {
		par := NewParser()
		var i int
		par.Int("", &i, "test flag")
		yesErr(t, par.Parse(nil))
	})

	t.Run("multiple aliases", func(t *testing.T) {
		par := NewParser()
		var i int
		par.Int("flag", &i, "test flag").Alias("f", "fl")
		noErr(t, par.Parse([]string{"-f", "1"}))
		eq(t, 1, i)
	})
}

func flagExpect[D Decoder[T], T comparable](t *testing.T, name string, args []string, expected T) {
	t.Helper()
	var dest T
	par := NewParser()

	Register[D](par, name, &dest, name)
	noErr(t, par.Parse(args))
	eq(t, expected, dest)
}

func flagExpectSlice[D Decoder[T], T comparable](
	t *testing.T, name string, args []string, expected []T,
) {
	var dest []T
	par := NewParser()

	RegisterSlice[D](par, name, &dest, name)
	noErr(t, par.Parse(args))
	eq(t, expected, dest)
}

func TestParser_FlagTypes(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected any
	}{
		{"int_flag", []string{"--int_flag", "123"}, 123},
		{"string_flag", []string{"--string_flag", "hello"}, "hello"},
		{"int_slice_flag", []string{"--int_slice_flag", "1", "--int_slice_flag", "2"}, []int{1, 2}},
		{"string_slice_flag",
			[]string{"--string_slice_flag", "hello", "--string_slice_flag", "world"},
			[]string{"hello", "world"}},
		{"zero_value", []string{}, 0},
		{"bool_present", []string{"--bool_present"}, true},
		{"bool_absent", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch concrete := tt.expected.(type) {
			case int:
				flagExpect[Int](t, tt.name, tt.args, concrete)
			case string:
				flagExpect[String](t, tt.name, tt.args, concrete)
			case bool:
				flagExpect[Bool](t, tt.name, tt.args, concrete)
			case []int:
				flagExpectSlice[Int](t, tt.name, tt.args, concrete)
			case []string:
				flagExpectSlice[String](t, tt.name, tt.args, concrete)
			default:
				t.Errorf("unsupported type %T", tt.expected)
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
		{"no flags", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"mixed flags and positional", []string{"--flag", "23", "a", "b"}, []string{"a", "b"}},
		{"positional after flags",
			[]string{"--flag", "23", "a", "--flag2", "42", "b"},
			[]string{"a", "b"}},
		{"only positional", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var positional []string
			par := NewParser().Positional(&positional)
			var i int
			par.Int("flag", &i, "test flag")
			par.Int("flag2", &i, "test flag2")
			noErr(t, par.Parse(tt.args))
			eq(t, tt.expected, positional)
		})
	}
}

func TestParser_SliceTerminator(t *testing.T) {
	var positional []string
	par := NewParser().Positional(&positional)
	var values []string
	par.StringSlice("values", &values, "test values")

	noErr(t, par.Parse([]string{"--values", "first", "--", "positional"}))
	eq(t, []string{"first"}, values)
	eq(t, []string{"positional"}, positional)
}

func TestParser_FlagTerminator(t *testing.T) {
	for _, tc := range []struct {
		name           string
		args           []string
		wantValue      int
		wantSliceValue []int
		wantPositional []string
	}{
		{
			name:           "flags after separator are positional",
			args:           []string{"--", "--flag", "23", "positional"},
			wantPositional: []string{"--flag", "23", "positional"},
		},
		{
			name:           "separator without following arguments",
			args:           []string{"--"},
			wantPositional: nil,
		},
		{
			name:           "flag-like arguments after separator are positional",
			args:           []string{"--", "--flag", "23", "--slice-flag", "1", "positional"},
			wantPositional: []string{"--flag", "23", "--slice-flag", "1", "positional"},
		},
		{
			name:           "flags before separator are parsed",
			args:           []string{"--flag", "7", "--", "--flag", "23", "positional"},
			wantValue:      7,
			wantPositional: []string{"--flag", "23", "positional"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var positional []string
			par := NewParser().Positional(&positional)
			var value int
			var sliceValue []int
			par.Int("flag", &value, "test flag")
			par.IntSlice("slice-flag", &sliceValue, "test slice flag")

			noErr(t, par.Parse(tc.args))
			eq(t, tc.wantValue, value)
			eq(t, tc.wantSliceValue, sliceValue)
			eq(t, tc.wantPositional, positional)
		})
	}
}

func TestParser_PositionalArgumentsPolicy(t *testing.T) {
	t.Run("refused by default", func(t *testing.T) {
		par := NewParser()
		yesErr(t, par.Parse([]string{"value"}))
	})

	t.Run("allowed after a flag", func(t *testing.T) {
		var positional []string
		par := NewParser().Positional(&positional)
		var value string
		par.String("flag", &value, "test flag")

		noErr(t, par.Parse([]string{"--flag", "flag value", "positional"}))
		eq(t, "flag value", value)
		eq(t, []string{"positional"}, []string(positional))
	})

	t.Run("refused after a flag", func(t *testing.T) {
		par := NewParser()
		var value string
		par.String("flag", &value, "test flag")

		yesErr(t, par.Parse([]string{"--flag", "flag value", "positional"}))
	})
}

func TestParser_DefaultValues(t *testing.T) {
	t.Run("int default", func(t *testing.T) {
		par := NewParser()
		i := 42
		par.Int("flag", &i, "test flag").Default(23)
		noErr(t, par.Parse([]string{}))
		eq(t, 23, i)
	})

	t.Run("string default", func(t *testing.T) {
		par := NewParser()
		s := "foo"
		par.String("flag", &s, "test flag").Default("bar")
		noErr(t, par.Parse([]string{}))
		eq(t, "bar", s)
	})

	t.Run("slice default", func(t *testing.T) {
		par := NewParser()
		var s []int
		par.IntSlice("flag", &s, "test flag").Default([]int{1, 2, 3})
		noErr(t, par.Parse([]string{}))
		eq(t, []int{1, 2, 3}, s)
	})

	t.Run("string slice default", func(t *testing.T) {
		par := NewParser()
		var s []string
		par.StringSlice("flag", &s, "test flag").Default([]string{"foo", "bar"})
		noErr(t, par.Parse([]string{}))
		eq(t, []string{"foo", "bar"}, s)
	})
}

func TestParser_SliceFlagDefaultAndAccumulation(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{"first invocation overwrites default", []string{"--values", "first"}, []string{"first"}},
		{"multiple invocations accumulate",
			[]string{"--values", "first", "--values", "second"},
			[]string{"first", "second"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			par := NewParser()
			var values []string
			par.StringSlice("values", &values, "test values").Default([]string{"default"})

			noErr(t, par.Parse(tt.args))
			eq(t, tt.expected, values)
		})
	}
}

func TestParser_BooleanFlags(t *testing.T) {
	t.Run("boolean flag default value", func(t *testing.T) {
		par := NewParser()
		b := true
		par.Bool("flag", &b, "test flag").Default(false)
		noErr(t, par.Parse([]string{}))
		eq(t, false, b)
	})
}

func TestParser_StringSlice(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{"single value", []string{"--flag", "hello"}, []string{"hello"}},
		{"multiple values",
			[]string{"--flag", "hello", "--flag", "world"},
			[]string{"hello", "world"}},
		{"mixed with other flags",
			[]string{"--flag", "hello", "--other", "42", "--flag", "world"},
			[]string{"hello", "world"}},
		{"no values", []string{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			par := NewParser()
			var s []string
			var i int
			par.StringSlice("flag", &s, "test flag")
			par.Int("other", &i, "other flag")
			noErr(t, par.Parse(tt.args))
			eq(t, tt.expected, s)
		})
	}
}

func TestParser_AllowedValues(t *testing.T) {
	t.Run("allowed values accept a matching value", func(t *testing.T) {
		par := NewParser()
		var value int
		par.Int("value", &value, "test value").Allowed(1, 2)

		noErr(t, par.Parse([]string{"--value", "2"}))
		eq(t, 2, value)
	})

	t.Run("allowed values reject a non-matching value", func(t *testing.T) {
		par := NewParser()
		value := 9
		par.Int("value", &value, "test value").Allowed(1, 2)

		yesErr(t, par.Parse([]string{"--value", "3"}))
		eq(t, 9, value)
	})

	t.Run("custom allowed function", func(t *testing.T) {
		par := NewParser()
		var value string
		par.String("value", &value, "test value").AllowedFun(func(value string) bool {
			return len(value) >= 3
		})

		noErr(t, par.Parse([]string{"--value", "valid"}))
		eq(t, "valid", value)

		yesErr(t, par.Parse([]string{"--value", "no"}))
	})

	t.Run("default values are allowed even when not listed", func(t *testing.T) {
		par := NewParser()
		var value int
		par.Int("value", &value, "test value").Default(3).Allowed(1, 2)

		noErr(t, par.Parse(nil))
		eq(t, 3, value)
	})
}

// TestParser_FluentMethodsForArgumentTypes tests that method chaining works properly.
func TestParser_FluentMethodsForArgumentTypes(t *testing.T) {
	par := NewParser()
	var intValue int
	var stringValue string
	var boolValue bool
	var intValues []int
	var stringValues []string

	par.Int("int", &intValue, "integer").
		Alias("i").
		Default(10).
		Allowed(10, 20).
		AllowedFun(func(value int) bool { return value >= 10 && value <= 20 })
	par.String("string", &stringValue, "string").
		Alias("s").
		Default("default").
		Allowed("hello", "world").
		AllowedFun(func(value string) bool { return len(value) >= 3 })
	par.Bool("bool", &boolValue, "boolean").Alias("b").Default(true)
	par.IntSlice("ints", &intValues, "integers").
		Alias("n").
		Default([]int{1}).
		Allowed(2, 3).
		AllowedFun(func(value int) bool { return value >= 2 })
	par.StringSlice("strings", &stringValues, "strings").
		Alias("w").
		Default([]string{"default"}).
		Allowed("world").
		AllowedFun(func(value string) bool { return len(value) >= 3 })

	noErr(t, par.Parse([]string{
		"-i", "20",
		"-s", "hello",
		"-b",
		"-n", "2",
		"-w", "world",
	}))
	eq(t, 20, intValue)
	eq(t, "hello", stringValue)
	eq(t, true, boolValue)
	eq(t, []int{2}, intValues)
	eq(t, []string{"world"}, stringValues)
}

func TestParser_SliceAllowedValues(t *testing.T) {
	t.Run("allowed values are checked individually", func(t *testing.T) {
		par := NewParser()
		var values []int
		par.IntSlice("values", &values, "test values").Allowed(1, 2)

		yesErr(t, par.Parse([]string{"--values", "1", "--values", "3"}))
		eq(t, []int{1}, values)
	})

	t.Run("custom allowed function checks each value", func(t *testing.T) {
		par := NewParser()
		var values []string
		par.StringSlice("values", &values, "test values").
			AllowedFun(func(value string) bool { return len(value) >= 3 })

		yesErr(t, par.Parse([]string{"--values", "valid", "--values", "no"}))
		eq(t, []string{"valid"}, values)
	})
}

func TestParser_InvalidIntArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"empty string", []string{"--int_flag", ""}},
		{"non-numeric string", []string{"--int_flag", "abc"}},
		{"float number", []string{"--int_flag", "1.23"}},
		{"number with spaces", []string{"--int_flag", " 123 "}},
		{"number with commas", []string{"--int_flag", "1,000"}},
		{"hex number", []string{"--int_flag", "0xFF"}},
		{"binary number", []string{"--int_flag", "0b1010"}},
		{"number with units", []string{"--int_flag", "123KB"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			par := NewParser()
			var i int
			par.Int("int_flag", &i, "test flag")
			yesErr(t, par.Parse(tt.args))
		})
	}
}
