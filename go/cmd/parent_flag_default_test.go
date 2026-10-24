package cmd_test

import (
	"testing"

	"github.com/mooss/bagend/go/cmd"
	"github.com/mooss/bagend/go/flag"
)

func TestParentFlagDefaultsBeforeSubcommandAreValid(t *testing.T) {
	t.Run("string default retained", testParentStringDefaultRetained)
	t.Run("bool default retained", testParentBoolDefaultRetained)
	t.Run("slice default retained", testParentSliceDefaultRetained)
	t.Run("override default before subcommand", testParentDefaultOverrideBeforeSubcommand)
	t.Run("defaults retained for nested subcommand", testParentDefaultsRetainedForNested)
	t.Run("defaults visible to runner via Execute", testParentDefaultsVisibleToRunner)
	t.Run("child defaults unaffected by parent", testChildDefaultsUnaffectedByParent)
}

func testParentStringDefaultRetained(t *testing.T) {
	var parentName string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().String("name", &parentName, "Parent name").Default("default-parent")
	root.RegisterSub(&cmd.Command{Name: "child"})

	if err := root.Parse([]string{"child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if parentName != "default-parent" {
		t.Fatalf("parent name = %q, want default %q", parentName, "default-parent")
	}
}

func testParentBoolDefaultRetained(t *testing.T) {
	var verbose bool
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().Bool("verbose", &verbose, "Verbose output").Default(true)
	root.RegisterSub(&cmd.Command{Name: "child"})

	if err := root.Parse([]string{"child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if !verbose {
		t.Fatal("parent verbose = false, want default true")
	}
}

func testParentSliceDefaultRetained(t *testing.T) {
	var values []string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().StringSlice("values", &values, "Values").Default([]string{"default-value"})
	root.RegisterSub(&cmd.Command{Name: "child"})

	if err := root.Parse([]string{"child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if len(values) != 1 || values[0] != "default-value" {
		t.Fatalf("parent values = %v, want [default-value]", values)
	}
}

func testParentDefaultOverrideBeforeSubcommand(t *testing.T) {
	var parentName string
	var childName string
	newRoot := func() *cmd.Command {
		parentName = ""
		childName = ""
		root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
		root.FlagsParser().String("name", &parentName, "Parent name").Default("default-parent")
		child := &cmd.Command{Name: "child", Flags: flag.NewParser()}
		child.FlagsParser().String("name", &childName, "Child name").Default("default-child")
		root.RegisterSub(child)
		return root
	}

	root := newRoot()
	if err := root.Parse([]string{"--name", "overridden", "child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if parentName != "overridden" {
		t.Fatalf("parent name = %q, want %q", parentName, "overridden")
	}
	if childName != "default-child" {
		t.Fatalf("child name = %q, want default %q", childName, "default-child")
	}

	root = newRoot()
	if err := root.Parse([]string{"--name", "overridden", "child", "--name", "kid"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertDispatchNames(t, parentName, childName, "overridden", "kid")
}

func testParentDefaultsRetainedForNested(t *testing.T) {
	var parentName string
	newRoot := func() *cmd.Command {
		parentName = ""
		root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
		root.FlagsParser().String("name", &parentName, "Parent name").Default("default-parent")
		middle := &cmd.Command{Name: "middle"}
		leaf := &cmd.Command{Name: "leaf"}
		middle.RegisterSub(leaf)
		root.RegisterSub(middle)
		return root
	}

	root := newRoot()
	if err := root.Parse([]string{"middle", "leaf"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if parentName != "default-parent" {
		t.Fatalf("parent name = %q, want default %q", parentName, "default-parent")
	}

	root = newRoot()
	if err := root.Parse([]string{"--name", "custom", "middle", "leaf"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if parentName != "custom" {
		t.Fatalf("parent name = %q, want %q", parentName, "custom")
	}
}

func testParentDefaultsVisibleToRunner(t *testing.T) {
	var parentName string
	var observed string
	var leafRan bool
	newRoot := func() *cmd.Command {
		parentName = ""
		root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
		root.FlagsParser().String("name", &parentName, "Parent name").Default("default-parent")
		middle := &cmd.Command{Name: "middle"}
		leaf := &cmd.Command{Name: "leaf"}
		leaf.Runner = func() error {
			observed = parentName
			leafRan = true
			return nil
		}
		middle.RegisterSub(leaf)
		root.RegisterSub(middle)
		return root
	}

	root := newRoot()
	if err := root.Execute([]string{"middle", "leaf"}); err != nil {
		t.Fatalf("Execute() returned an error: %v", err)
	}
	if !leafRan {
		t.Fatal("leaf runner was not called")
	}
	if observed != "default-parent" {
		t.Fatalf("runner observed parent name = %q, want default %q", observed, "default-parent")
	}

	leafRan = false
	root = newRoot()
	if err := root.Execute([]string{"--name", "Jack", "middle", "leaf"}); err != nil {
		t.Fatalf("Execute() returned an error: %v", err)
	}
	if observed != "Jack" {
		t.Fatalf("runner observed parent name = %q, want %q", observed, "Jack")
	}
}

func testChildDefaultsUnaffectedByParent(t *testing.T) {
	var parentName string
	var childName string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().String("name", &parentName, "Parent name")
	child := &cmd.Command{Name: "child", Flags: flag.NewParser()}
	child.FlagsParser().String("name", &childName, "Child name").Default("default-child")
	root.RegisterSub(child)

	if err := root.Parse([]string{"--name", "parent-value", "child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if parentName != "parent-value" {
		t.Fatalf("parent name = %q, want %q", parentName, "parent-value")
	}
	if childName != "default-child" {
		t.Fatalf("child name = %q, want default %q", childName, "default-child")
	}
}
