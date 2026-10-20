package lie

import (
	"testing"
)

func testMap[From any, To comparable](
	t *testing.T, name string, input []From, fn func(From) To, expected []To,
) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		result := Map(fn, input)
		if len(result) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(result))
		}
		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("at index %d: expected %v, got %v", i, expected[i], result[i])
			}
		}
	})
}

//revive:disable:cognitive-complexity
func TestMapFilter(t *testing.T) {
	t.Run("filters nil results and preserves order", func(t *testing.T) {
		result := MapFilter(func(x int) *int {
			if x%2 != 0 {
				return nil
			}
			return &x
		}, []int{1, 2, 3, 4})

		expected := []int{2, 4}
		if len(result) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(result))
		}
		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], result[i])
			}
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := MapFilter(func(x int) *int { return &x }, []int{})
		if len(result) != 0 {
			t.Fatalf("expected empty result, got %v", result)
		}
	})

	t.Run("all results nil", func(t *testing.T) {
		result := MapFilter(func(int) *int { return nil }, []int{1, 2})
		if len(result) != 0 {
			t.Fatalf("expected empty result, got %v", result)
		}
	})
}

func TestMap(t *testing.T) {
	testMap(t, "int to square",
		[]int{1, 2, 3, 4},
		func(x int) int { return x * x },
		[]int{1, 4, 9, 16},
	)

	testMap(t, "string to length",
		[]string{"a", "bb", "ccc"},
		func(s string) int { return len(s) },
		[]int{1, 2, 3},
	)

	testMap(t, "empty slice",
		[]int{},
		func(x int) int { return x * 2 },
		[]int{},
	)

	testMap(t, "digit to string",
		[]int{1, 2, 3},
		func(x int) string { return string(rune(x + '0')) }, //nolint:gosec
		[]string{"1", "2", "3"},
	)
}
