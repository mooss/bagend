package fun

import "testing"

func TestIdentity(t *testing.T) {
	value := 42
	if got := Identity(value); got != value {
		t.Fatalf("Identity(%d) = %d, want %d", value, got, value)
	}
}

func TestCoalesce(t *testing.T) {
	tests := []struct {
		name  string
		first int
		rest  []int
		want  int
	}{
		{name: "first is non-zero", first: 1, rest: []int{2, 3}, want: 1},
		{name: "uses first non-zero fallback", first: 0, rest: []int{0, 3, 4}, want: 3},
		{name: "all values are zero", first: 0, rest: []int{0, 0}, want: 0},
		{name: "no fallbacks", first: 0, want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Coalesce(test.first, test.rest...); got != test.want {
				t.Fatalf("Coalesce(%d, %v) = %d, want %d", test.first, test.rest, got, test.want)
			}
		})
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "zero integer", value: 0, want: true},
		{name: "non-zero integer", value: 1, want: false},
		{name: "empty string", value: "", want: true},
		{name: "non-empty string", value: "value", want: false},
		{name: "nil pointer", value: (*int)(nil), want: true},
		{name: "non-nil pointer", value: new(int), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsZero(test.value); got != test.want {
				t.Fatalf("IsZero(%#v) = %t, want %t", test.value, got, test.want)
			}
		})
	}
}

func TestZero(t *testing.T) {
	//nolint:unused
	type sample struct {
		count int
		name  string
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{name: "integer", got: Zero[int](), want: 0},
		{name: "string", got: Zero[string](), want: ""},
		{name: "struct", got: Zero[sample](), want: sample{}},
		{name: "pointer", got: Zero[*int](), want: (*int)(nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("Zero() = %#v, want %#v", test.got, test.want)
			}
		})
	}
}

func TestZeroDeref(t *testing.T) {
	t.Run("nil pointer returns zero value", func(t *testing.T) {
		if got := ZeroDeref[int](nil); got != 0 {
			t.Fatalf("ZeroDeref(nil) = %d, want 0", got)
		}
	})

	t.Run("non-nil pointer returns pointed value", func(t *testing.T) {
		value := 42
		if got := ZeroDeref(&value); got != value {
			t.Fatalf("ZeroDeref(&value) = %d, want %d", got, value)
		}
	})

	t.Run("non-nil pointer to zero returns zero value", func(t *testing.T) {
		value := 0
		if got := ZeroDeref(&value); got != 0 {
			t.Fatalf("ZeroDeref(&value) = %d, want 0", got)
		}
	})
}
