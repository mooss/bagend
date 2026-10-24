package cmd_test

import (
	"testing"

	"github.com/mooss/bagend/go/cmd"
	"github.com/mooss/bagend/go/flag"
)

func TestParentSliceDefaultCompletelyReplacedNotExtended(t *testing.T) {
	t.Run("single value replaces single default", testParentSliceSingleValueReplacesDefault)
	t.Run("repeated flag replaces single default", testParentSliceRepeatedFlagReplacesDefault)
	t.Run("single value replaces multiple defaults", testParentSliceSingleValueReplacesMultipleDefaults)
	t.Run("multiple values replace multiple defaults", testParentSliceMultipleValuesReplaceMultipleDefaults)
	t.Run("replacement before nested subcommand", testParentSliceReplacementBeforeNestedSubcommand)
	t.Run("replacement visible to runner via Execute", testParentSliceReplacementVisibleToRunner)
	t.Run("child slice unaffected by parent replacement", testParentSliceChildUnaffectedByParentReplacement)
	t.Run("parent default retained when only child supplies slice", testParentSliceRetainedWhenChildSuppliesSlice)
	t.Run("default slice not mutated across parses", testParentSliceDefaultSliceNotMutated)
}

func assertSliceExactly(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("slice = %v (len %d), want %v (len %d): default was extended instead of replaced", got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("slice = %v, want %v: mismatch at index %d", got, want, i)
		}
	}
}

func testParentSliceSingleValueReplacesDefault(t *testing.T) {
	var values []string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().StringSlice("values", &values, "Values").Default([]string{"default-value"})
	root.RegisterSub(&cmd.Command{Name: "child"})

	if err := root.Parse([]string{"--values", "actual", "--", "child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, values, []string{"actual"})
}

func testParentSliceRepeatedFlagReplacesDefault(t *testing.T) {
	var values []string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().StringSlice("values", &values, "Values").Default([]string{"default-value"})
	root.RegisterSub(&cmd.Command{Name: "child"})

	if err := root.Parse([]string{"--values", "a", "--values", "b", "--", "child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, values, []string{"a", "b"})
}

func testParentSliceSingleValueReplacesMultipleDefaults(t *testing.T) {
	var values []string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().StringSlice("values", &values, "Values").Default([]string{"d1", "d2", "d3"})
	root.RegisterSub(&cmd.Command{Name: "child"})

	if err := root.Parse([]string{"--values", "actual", "--", "child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, values, []string{"actual"})
}

func testParentSliceMultipleValuesReplaceMultipleDefaults(t *testing.T) {
	var values []string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().StringSlice("values", &values, "Values").Default([]string{"d1", "d2"})
	root.RegisterSub(&cmd.Command{Name: "child"})

	if err := root.Parse([]string{"--values", "a", "--values", "b", "--values", "c", "--",  "child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, values, []string{"a", "b", "c"})
}

func testParentSliceReplacementBeforeNestedSubcommand(t *testing.T) {
	var values []string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().StringSlice("values", &values, "Values").Default([]string{"default-value"})
	middle := &cmd.Command{Name: "middle"}
	leaf := &cmd.Command{Name: "leaf"}
	middle.RegisterSub(leaf)
	root.RegisterSub(middle)

	if err := root.Parse([]string{"--values", "actual", "--", "middle", "leaf"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, values, []string{"actual"})

	var values2 []string
	root2 := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root2.FlagsParser().StringSlice("values", &values2, "Values").Default([]string{"d1", "d2"})
	middle2 := &cmd.Command{Name: "middle"}
	leaf2 := &cmd.Command{Name: "leaf"}
	middle2.RegisterSub(leaf2)
	root2.RegisterSub(middle2)

	if err := root2.Parse([]string{"--values", "x", "--values", "y", "--", "middle", "leaf"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, values2, []string{"x", "y"})
}

func testParentSliceReplacementVisibleToRunner(t *testing.T) {
	var values []string
	var observed []string
	var leafRan bool
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().StringSlice("values", &values, "Values").Default([]string{"default-value"})
	middle := &cmd.Command{Name: "middle"}
	leaf := &cmd.Command{Name: "leaf"}
	leaf.Runner = func() error {
		observed = append([]string(nil), values...)
		leafRan = true
		return nil
	}
	middle.RegisterSub(leaf)
	root.RegisterSub(middle)

	if err := root.Execute([]string{"--values", "actual", "--", "middle", "leaf"}); err != nil {
		t.Fatalf("Execute() returned an error: %v", err)
	}
	if !leafRan {
		t.Fatal("leaf runner was not called")
	}
	assertSliceExactly(t, observed, []string{"actual"})
	assertSliceExactly(t, values, []string{"actual"})

	var values2 []string
	var observed2 []string
	var leafRan2 bool
	root2 := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root2.FlagsParser().StringSlice("values", &values2, "Values").Default([]string{"default-value"})
	leaf2 := &cmd.Command{Name: "leaf"}
	leaf2.Runner = func() error {
		observed2 = append([]string(nil), values2...)
		leafRan2 = true
		return nil
	}
	root2.RegisterSub(leaf2)

	if err := root2.Execute([]string{"--values", "a", "--values", "b", "--", "leaf"}); err != nil {
		t.Fatalf("Execute() returned an error: %v", err)
	}
	if !leafRan2 {
		t.Fatal("leaf runner was not called")
	}
	assertSliceExactly(t, observed2, []string{"a", "b"})
}

func testParentSliceChildUnaffectedByParentReplacement(t *testing.T) {
	setup := func() (*cmd.Command, *[]string, *[]string) {
		var parentValues []string
		var childValues []string
		root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
		root.FlagsParser().StringSlice("values", &parentValues, "Parent values").
			Default([]string{"parent-default"})
		child := &cmd.Command{Name: "child", Flags: flag.NewParser()}
		child.FlagsParser().StringSlice("values", &childValues, "Child values").
			Default([]string{"child-default"})
		root.RegisterSub(child)
		return root, &parentValues, &childValues
	}

	root, parentValues, childValues := setup()
	if err := root.Parse([]string{"--values", "parent-actual", "--", "child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, *parentValues, []string{"parent-actual"})
	assertSliceExactly(t, *childValues, []string{"child-default"})

	root2, parentValues2, childValues2 := setup()
	if err := root2.Parse([]string{"--values", "parent-actual", "--", "child", "--values", "child-actual"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, *parentValues2, []string{"parent-actual"})
	assertSliceExactly(t, *childValues2, []string{"child-actual"})
}

func testParentSliceRetainedWhenChildSuppliesSlice(t *testing.T) {
	var parentValues []string
	var childValues []string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().StringSlice("values", &parentValues, "Parent values").Default([]string{"parent-default"})
	child := &cmd.Command{Name: "child", Flags: flag.NewParser()}
	child.FlagsParser().StringSlice("values", &childValues, "Child values").Default([]string{"child-default"})
	root.RegisterSub(child)

	if err := root.Parse([]string{"child", "--values", "child-actual"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, parentValues, []string{"parent-default"})
	assertSliceExactly(t, childValues, []string{"child-actual"})
}

func testParentSliceDefaultSliceNotMutated(t *testing.T) {
	def := []string{"default-value"}
	defCopy := append([]string(nil), def...)

	var values []string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().StringSlice("values", &values, "Values").Default(def)
	root.RegisterSub(&cmd.Command{Name: "child"})

	if err := root.Parse([]string{"--values", "actual", "--",  "child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, values, []string{"actual"})
	assertSliceExactly(t, def, defCopy)

	var values2 []string
	root2 := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root2.FlagsParser().StringSlice("values", &values2, "Values").Default([]string{"default-value"})
	root2.RegisterSub(&cmd.Command{Name: "child"})

	if err := root2.Parse([]string{"child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertSliceExactly(t, values2, []string{"default-value"})
}
