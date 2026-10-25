package cmd_test

import (
	"testing"

	"github.com/mooss/bagend/go/cmd"
	"github.com/mooss/bagend/go/flag"
)

func TestParentFlagsBeforeSubcommandAreTakenIntoAccount(t *testing.T) {
	t.Run("string flag before subcommand", testParentStringFlagBeforeSubcommand)
	t.Run("bool flag before subcommand", testParentBoolFlagBeforeSubcommand)
	t.Run("multiple parent flags before subcommand", testParentMultipleFlagsBeforeSubcommand)
	t.Run("parent and child flags combined", testParentAndChildFlagsCombined)
	t.Run("string slice flag before subcommand", testParentSliceFlagBeforeSubcommand)
	t.Run("parent flags before nested subcommand", testParentFlagsBeforeNestedSubcommand)
	t.Run("parent flags via Execute dispatches child", testParentFlagsBeforeSubcommandViaExecute)
	t.Run("parent flags before alias", testParentFlagsBeforeAlias)
	t.Run(
		"parent flag after subcommand belongs to child",
		testParentFlagAfterSubcommandBelongsToChild,
	)
}

func testParentStringFlagBeforeSubcommand(t *testing.T) {
	var parentName string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().String("name", &parentName, "Parent name")
	root.RegisterSub(&cmd.Command{Name: "child"})

	if err := root.Parse([]string{"--name", "Locke", "child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if parentName != "Locke" {
		t.Fatalf("parent name = %q, want %q", parentName, "Locke")
	}
}

func testParentBoolFlagBeforeSubcommand(t *testing.T) {
	var verbose bool
	var childSelected bool
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().Bool("verbose", &verbose, "Verbose output")
	child := &cmd.Command{Name: "child", Flags: flag.NewParser()}
	child.FlagsParser().Bool("selected", &childSelected, "Selected")
	root.RegisterSub(child)

	if err := root.Parse([]string{"--verbose", "child", "--selected"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if !verbose {
		t.Fatal("parent verbose = false, want true")
	}
	if !childSelected {
		t.Fatal("child selected = false, want true")
	}
}

func testParentMultipleFlagsBeforeSubcommand(t *testing.T) {
	var name, mode string
	var verbose bool
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().String("name", &name, "Name")
	root.FlagsParser().String("mode", &mode, "Mode")
	root.FlagsParser().Bool("verbose", &verbose, "Verbose")
	root.RegisterSub(&cmd.Command{Name: "child"})

	args := []string{"--name", "Jack", "--mode", "fast", "--verbose", "child"}
	if err := root.Parse(args); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if name != "Jack" {
		t.Errorf("parent name = %q, want %q", name, "Jack")
	}
	if mode != "fast" {
		t.Errorf("parent mode = %q, want %q", mode, "fast")
	}
	if !verbose {
		t.Error("parent verbose = false, want true")
	}
}

func testParentAndChildFlagsCombined(t *testing.T) {
	var parentName, childName string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().String("name", &parentName, "Parent name")
	child := &cmd.Command{Name: "child", Flags: flag.NewParser()}
	child.FlagsParser().String("name", &childName, "Child name")
	root.RegisterSub(child)

	if err := root.Parse([]string{"--name", "Jack", "child", "--name", "Kate"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertDispatchNames(t, parentName, childName, "Jack", "Kate")
}

func testParentSliceFlagBeforeSubcommand(t *testing.T) {
	var values []string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().StringSlice("values", &values, "Values")
	child := &cmd.Command{Name: "child"}
	root.RegisterSub(child)

	args := []string{"--values", "one", "--values", "two", "--", "child"}
	if err := root.Parse(args); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if len(values) != 2 || values[0] != "one" || values[1] != "two" {
		t.Fatalf("parent values = %v, want [one two]", values)
	}
}

func testParentFlagsBeforeNestedSubcommand(t *testing.T) {
	var name string
	var leafValue string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().String("name", &name, "Name")
	middle := &cmd.Command{Name: "middle"}
	leaf := &cmd.Command{Name: "leaf", Flags: flag.NewParser()}
	leaf.FlagsParser().String("value", &leafValue, "A value")
	middle.RegisterSub(leaf)
	root.RegisterSub(middle)

	args := []string{"--name", "Locke", "middle", "leaf", "--value", "nested"}
	if err := root.Parse(args); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if name != "Locke" {
		t.Fatalf("parent name = %q, want %q", name, "Locke")
	}
	if leafValue != "nested" {
		t.Fatalf("leaf value = %q, want %q", leafValue, "nested")
	}
}

func testParentFlagsBeforeSubcommandViaExecute(t *testing.T) {
	var name string
	var childRan bool
	var childValue string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().String("name", &name, "Name")
	child := &cmd.Command{Name: "child", Flags: flag.NewParser()}
	child.FlagsParser().String("value", &childValue, "A value")
	child.Runner = func() error {
		childRan = true
		return nil
	}
	root.RegisterSub(child)

	if err := root.Execute([]string{"--name", "Sawyer", "child", "--value", "island"}); err != nil {
		t.Fatalf("Execute() returned an error: %v", err)
	}
	if name != "Sawyer" {
		t.Errorf("parent name = %q, want %q", name, "Sawyer")
	}
	if childValue != "island" {
		t.Errorf("child value = %q, want %q", childValue, "island")
	}
	if !childRan {
		t.Fatal("child runner was not called")
	}
}

func testParentFlagsBeforeAlias(t *testing.T) {
	var parentName, childValue string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().String("name", &parentName, "Parent name")
	child := &cmd.Command{Name: "child", Flags: flag.NewParser()}
	child.FlagsParser().String("value", &childValue, "A value")
	child.Aliases = append(child.Aliases, "c")
	root.RegisterSub(child)

	if err := root.Parse([]string{"--name", "Hurley", "c", "--value", "numbers"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if parentName != "Hurley" {
		t.Fatalf("parent name = %q, want %q", parentName, "Hurley")
	}
	if childValue != "numbers" {
		t.Fatalf("child value = %q, want %q", childValue, "numbers")
	}
}

func testParentFlagAfterSubcommandBelongsToChild(t *testing.T) {
	var parentName, childName string
	root := &cmd.Command{Name: "root", Flags: flag.NewParser()}
	root.FlagsParser().String("name", &parentName, "Parent name")
	child := &cmd.Command{Name: "child", Flags: flag.NewParser()}
	child.FlagsParser().String("name", &childName, "Child name")
	root.RegisterSub(child)

	if err := root.Parse([]string{"child", "--name", "Locke"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertDispatchNames(t, parentName, childName, "", "Locke")
}
