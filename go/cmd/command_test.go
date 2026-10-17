package cmd_test

import (
	"strings"
	"testing"

	"github.com/mooss/bagend/go/cmd"
)

func requireContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q", got, want)
	}
}

func TestCommandDefinition(t *testing.T) {
	var name string
	command := cmd.New("greet", "Print a greeting")
	command.Flags().String("name", &name, "Who to greet")

	if command.Name() != "greet" {
		t.Fatalf("Name() = %q, want %q", command.Name(), "greet")
	}
	if command.Docline() != "Print a greeting" {
		t.Fatalf("Docline() = %q, want %q", command.Docline(), "Print a greeting")
	}

	if err := command.Parse([]string{"--name", "Locke"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if name != "Locke" {
		t.Fatalf("parsed name = %q, want %q", name, "Locke")
	}
}

func TestSubcommandRegistrationOrderAndReplacement(t *testing.T) {
	root := cmd.New("root", "")
	first := cmd.New("first", "")
	second := cmd.New("second", "")
	replacement := cmd.New("first", "replacement")

	root.Add(first)
	root.Add(second)
	root.Add(replacement)

	got := root.Subcommands()
	if len(got) != 2 {
		t.Fatalf("Subcommands() returned %d commands, want 2", len(got))
	}
	if got[0] != replacement {
		t.Fatalf("first registered command = %p, want replacement %p", got[0], replacement)
	}
	if got[1] != second {
		t.Fatalf("second registered command = %p, want %p", got[1], second)
	}
}

func TestDispatchParsesSelectedCommand(t *testing.T) {
	var parentName, childName string
	root := cmd.New("root", "")
	root.Flags().String("name", &parentName, "Parent name")

	child := cmd.New("child", "")
	child.Flags().String("name", &childName, "Child name")
	root.Add(child)

	if err := root.Parse([]string{"child", "--name", "Locke"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	assertDispatchNames(t, parentName, childName, "", "Locke")

	if err := root.Parse([]string{"--name", "Sawyer"}); err != nil {
		t.Fatalf("Parse() returned an error for the root command: %v", err)
	}
	assertDispatchNames(t, parentName, childName, "Sawyer", "Locke")

	if err := root.Parse([]string{"--name", "Jack", "child", "--name", "Kate"}); err != nil {
		t.Fatalf("Parse() returned an error for both commands: %v", err)
	}
	assertDispatchNames(t, parentName, childName, "Jack", "Kate")
}

func assertDispatchNames(t *testing.T, parentName, childName, wantParent, wantChild string) {
	t.Helper()

	if parentName != wantParent {
		t.Errorf("parent name = %q, want %q", parentName, wantParent)
	}
	if childName != wantChild {
		t.Errorf("child name = %q, want %q", childName, wantChild)
	}
}

func TestSubcommandNameAsFlagValue(t *testing.T) {
	var rootName, childName string
	root := cmd.New("root", "")
	root.Flags().String("name", &rootName, "Root name")

	child := cmd.New("child", "")
	child.Flags().String("name", &childName, "Child name")
	root.Add(child)

	if err := root.Parse([]string{"--name", "child", "child", "--name", "Locke"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if rootName != "child" {
		t.Fatalf("root name = %q, want %q", rootName, "child")
	}
	if childName != "Locke" {
		t.Fatalf("child name = %q, want %q", childName, "Locke")
	}
}

func TestSubcommandNameAsSliceValue(t *testing.T) {
	var values []string
	root := cmd.New("root", "")
	root.Flags().StringSlice("values", &values, "Values")

	child := cmd.New("child", "")
	root.Add(child)

	if err := root.Parse([]string{"--values", "child"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if len(values) != 1 || values[0] != "child" {
		t.Fatalf("root values = %v, want [child]", values)
	}
}

func TestDispatchAfterSliceTerminator(t *testing.T) {
	var values []string
	root := cmd.New("root", "")
	root.Flags().StringSlice("values", &values, "Values")

	var childSelected bool
	child := cmd.New("child", "")
	child.Flags().Bool("selected", &childSelected, "Selected")
	root.Add(child)

	if err := root.Parse([]string{"--values", "item", "--", "child", "--selected"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if len(values) != 1 || values[0] != "item" {
		t.Fatalf("root values = %v, want [item]", values)
	}
	if !childSelected {
		t.Fatal("child was not selected")
	}
}

func TestUnmatchedPositionalBeforeKnownSubcommand(t *testing.T) {
	root := cmd.New("root", "")

	var childSelected bool
	child := cmd.New("child", "")
	child.Flags().Bool("selected", &childSelected, "Selected")
	root.Add(child)

	err := root.Parse([]string{"unknown", "child", "--selected"})
	if err == nil {
		t.Fatal("Parse() succeeded for an unmatched positional argument, want an error")
	}
	requireContains(t, err.Error(), `positional arguments are not allowed (got "unknown")`)
	if childSelected {
		t.Fatal("child was selected after an unmatched positional argument")
	}
}

func TestDispatchNestedCommand(t *testing.T) {
	var value string
	root := cmd.New("root", "")
	middle := cmd.New("middle", "")
	leaf := cmd.New("leaf", "")
	leaf.Flags().String("value", &value, "A value")
	middle.Add(leaf)
	root.Add(middle)

	if err := root.Parse([]string{"middle", "leaf", "--value", "nested"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if value != "nested" {
		t.Fatalf("parsed value = %q, want %q", value, "nested")
	}

	if err := root.Parse([]string{"middle", "--value", "nested"}); err == nil {
		t.Fatal("Parse() succeeded with --value passed to middle")
	}
	if err := root.Parse([]string{"--value", "nested"}); err == nil {
		t.Fatal("Parse() succeeded with --value passed to root")
	}
}

func TestFlagArgumentIsParsedByCurrentCommand(t *testing.T) {
	var mode string
	root := cmd.New("root", "")
	root.Flags().String("mode", &mode, "Execution mode")
	root.Add(cmd.New("run", ""))

	if err := root.Parse([]string{"--mode", "fast"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if mode != "fast" {
		t.Fatalf("parsed mode = %q, want %q", mode, "fast")
	}
}

func TestUnmatchedArgumentFallsBackToCurrentParser(t *testing.T) {
	root := cmd.New("root", "")
	root.Add(cmd.New("known", ""))

	err := root.Parse([]string{"unknown"})
	if err == nil {
		t.Fatal("Parse() succeeded for an unmatched positional argument, want an error")
	}
	requireContains(t, err.Error(), "positional arguments are not allowed")
}

func TestUnknownFlagInDispatchedCommandIncludesContext(t *testing.T) {
	root := cmd.New("root", "")
	root.Add(cmd.New("child", ""))

	err := root.Parse([]string{"child", "--unknown"})
	if err == nil {
		t.Fatal("Parse() succeeded for an unknown flag, want an error")
	}
	requireContains(t, err.Error(), "unknown flag: --unknown")
	requireContains(t, err.Error(), "child")
}

func TestSubcommandsConflictWithPositionalArguments(t *testing.T) {
	root := cmd.New("root", "")
	root.Flags().WithPositionalArguments()
	root.Add(cmd.New("child", ""))

	for _, arguments := range [][]string{nil, {"argument"}, {"child"}} {
		err := root.Parse(arguments)
		if err == nil {
			t.Fatalf("Parse(%v) succeeded, want an error", arguments)
		}
		requireContains(t, err.Error(), "cannot have both subcommands and positional arguments")
	}
}
