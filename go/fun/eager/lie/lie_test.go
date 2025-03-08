package lie

import (
	"testing"
)

func testMap[From any, To comparable](t *testing.T, name string, input []From, fn func(From) To, expected []To) {
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
		func(x int) string { return string(rune(x + '0')) },
		[]string{"1", "2", "3"},
	)
}
