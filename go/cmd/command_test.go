package cmd_test

import (
	"errors"
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
	command.Flags.String("name", &name, "Who to greet")

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

func TestCommandUsage(t *testing.T) {
	t.Run("sets usage and enables help", testCommandUsageSetsUsageAndEnablesHelp)
	t.Run("empty usage", testCommandUsageEmpty)
	t.Run("repeated usage() replace without duplicate help flags", testCommandUsageRepeatedCalls)
}

func testCommandUsageSetsUsageAndEnablesHelp(t *testing.T) {
	command := cmd.New("tool", "")
	if got := command.Usage("[args]"); got != command {
		t.Fatal("Usage() did not return the command for fluent configuration")
	}

	help := command.Help()
	requireContains(t, help, "Usage: tool [args]")
	requireContains(t, help, "--help")

	if err := command.Parse(nil); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
}

func testCommandUsageEmpty(t *testing.T) {
	command := cmd.New("tool", "").Usage("")

	help := command.Help()
	requireContains(t, help, "Usage: tool\n\nFlags:")
	requireContains(t, help, "--help")
	if strings.Contains(help, "Usage: tool \n") {
		t.Fatalf("Help() included a trailing usage space: %q", help)
	}

	if err := command.Parse(nil); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
}

func testCommandUsageRepeatedCalls(t *testing.T) {
	command := cmd.New("tool", "").Usage("[first]").Usage("[second]")

	help := command.Help()
	requireContains(t, help, "Usage: tool [second]")
	if strings.Contains(help, "[first]") {
		t.Fatalf("Help() retained the previous usage string: %q", help)
	}

	if err := command.Parse(nil); err != nil {
		t.Fatalf("Parse() returned an error after repeated Usage() calls: %v", err)
	}
}

func TestCommandDescription(t *testing.T) {
	command := cmd.New("tool", "Run a tool").
		Usage("[args]").
		Description("A longer description.")

	if got := command.Description("A longer description."); got != command {
		t.Fatal("Description() did not return the command for fluent configuration")
	}

	help := command.Help()
	usageIndex := strings.Index(help, "Usage: tool [args]")
	descriptionIndex := strings.Index(help, "A longer description.")
	flagsIndex := strings.Index(help, "Flags:")
	if usageIndex == -1 || descriptionIndex == -1 || flagsIndex == -1 {
		t.Fatalf("Help() is missing usage, description, or flags: %q", help)
	}
	if usageIndex >= descriptionIndex || descriptionIndex >= flagsIndex {
		t.Fatalf("Help() should place the description between usage and flags: %q", help)
	}

	got := cmd.New("plain", "").Help()
	if strings.Contains(got, "Flags:") && strings.Contains(got, "\\n\\n\\nFlags:") {
		t.Fatalf("Help() included an extra blank line without a description: %q", got)
	}
}

func TestCommandHelp(t *testing.T) {
	var name string
	root := cmd.New("tool", "Run a tool").Usage("[args]")
	root.Flags.String("name", &name, "Who to greet")

	child := cmd.New("child", "Run a child command")
	child.Subcommand(cmd.New("nested", "A nested command"))
	root.Subcommand(child)

	help := root.Help()
	for _, want := range []string{
		"Run a tool",
		"Usage: tool [args]",
		"Flags:",
		"--name",
		"Who to greet",
		"Commands:",
		"child  Run a child command",
	} {
		requireContains(t, help, want)
	}
	if strings.Contains(help, "A nested command") {
		t.Fatal("Help() included a nested command")
	}
}

func TestSubcommandRegistrationOrderAndReplacement(t *testing.T) {
	root := cmd.New("root", "")
	first := cmd.New("first", "")
	second := cmd.New("second", "")
	replacement := cmd.New("first", "replacement")

	first.Alias("old-first")
	replacement.Alias("new-first")
	root.Subcommand(first)
	root.Subcommand(second)
	root.Subcommand(replacement)

	if err := root.Parse([]string{"old-first"}); err == nil {
		t.Fatal("Parse() succeeded with a stale alias after command replacement")
	}
	if err := root.Parse([]string{"new-first"}); err != nil {
		t.Fatalf("Parse() returned an error for the replacement alias: %v", err)
	}

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

func TestRegister(t *testing.T) {
	root := cmd.New("root", "")
	first := root.Register("first", "First command")
	second := root.Register("second", "Second command")
	replacement := root.Register("first", "Replacement command")

	assertDifferentCommands(t, first, replacement)
	assertRegisteredCommands(t, root, replacement, second)
	assertReplacementDetails(t, replacement)
	assertCommandsParse(t, root, "first", "second")
}

func assertDifferentCommands(t *testing.T, first, replacement *cmd.Command) {
	t.Helper()

	if first == replacement {
		t.Fatal("Register() returned the same command for a repeated name")
	}
}

func assertRegisteredCommands(t *testing.T, root *cmd.Command, replacement, second *cmd.Command) {
	t.Helper()

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

func assertReplacementDetails(t *testing.T, replacement *cmd.Command) {
	t.Helper()

	if replacement.Name() != "first" {
		t.Fatalf("replacement name = %q, want %q", replacement.Name(), "first")
	}
	if replacement.Docline() != "Replacement command" {
		t.Fatalf("replacement docline = %q, want %q", replacement.Docline(), "Replacement command")
	}
}

func assertCommandsParse(t *testing.T, root *cmd.Command, names ...string) {
	t.Helper()

	for _, name := range names {
		if err := root.Parse([]string{name}); err != nil {
			t.Fatalf("Parse(%q) returned an error: %v", name, err)
		}
	}
}

func TestSubcommandAliases(t *testing.T) {
	var value string
	var childRan bool

	root := cmd.New("root", "")
	child := cmd.New("child", "")
	child.Flags.String("value", &value, "A value")
	child.Runner(func() error {
		childRan = true
		return nil
	})
	child.Alias("c")
	root.Subcommand(child)

	if got := root.Subcommands(); len(got) != 1 || got[0] != child {
		t.Fatalf("Subcommands() = %v, want only the canonical child", got)
	}
	if err := root.Parse([]string{"c", "--value", "alias"}); err != nil {
		t.Fatalf("Parse() with alias returned an error: %v", err)
	}
	if value != "alias" {
		t.Fatalf("parsed value = %q, want %q", value, "alias")
	}
	if err := root.Execute([]string{"c"}); err != nil {
		t.Fatalf("Execute() with alias returned an error: %v", err)
	}
	if !childRan {
		t.Fatal("alias did not select the child command runner")
	}
}

func TestDispatchParsesSelectedCommand(t *testing.T) {
	var parentName, childName string
	root := cmd.New("root", "")
	root.Flags.String("name", &parentName, "Parent name")

	child := cmd.New("child", "")
	child.Flags.String("name", &childName, "Child name")
	root.Subcommand(child)

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
	root.Flags.String("name", &rootName, "Root name")

	child := cmd.New("child", "")
	child.Flags.String("name", &childName, "Child name")
	root.Subcommand(child)

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
	root.Flags.StringSlice("values", &values, "Values")

	child := cmd.New("child", "")
	root.Subcommand(child)

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
	root.Flags.StringSlice("values", &values, "Values")

	var childSelected bool
	child := cmd.New("child", "")
	child.Flags.Bool("selected", &childSelected, "Selected")
	root.Subcommand(child)

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
	child.Flags.Bool("selected", &childSelected, "Selected")
	root.Subcommand(child)

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
	leaf.Flags.String("value", &value, "A value")
	middle.Subcommand(leaf)
	root.Subcommand(middle)

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
	root.Flags.String("mode", &mode, "Execution mode")
	root.Subcommand(cmd.New("run", ""))

	if err := root.Parse([]string{"--mode", "fast"}); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if mode != "fast" {
		t.Fatalf("parsed mode = %q, want %q", mode, "fast")
	}
}

func TestUnmatchedArgumentFallsBackToCurrentParser(t *testing.T) {
	root := cmd.New("root", "")
	root.Subcommand(cmd.New("known", ""))

	err := root.Parse([]string{"unknown"})
	if err == nil {
		t.Fatal("Parse() succeeded for an unmatched positional argument, want an error")
	}
	requireContains(t, err.Error(), "positional arguments are not allowed")
}

func TestUnknownFlagInDispatchedCommandIncludesContext(t *testing.T) {
	root := cmd.New("root", "")
	root.Subcommand(cmd.New("child", ""))

	err := root.Parse([]string{"child", "--unknown"})
	if err == nil {
		t.Fatal("Parse() succeeded for an unknown flag, want an error")
	}
	requireContains(t, err.Error(), "unknown flag: --unknown")
	requireContains(t, err.Error(), "child")
}

func TestSubcommandsConflictWithPositionalArguments(t *testing.T) {
	root := cmd.New("root", "")
	root.Flags.WithPositionalArguments()
	root.Subcommand(cmd.New("child", ""))

	for _, arguments := range [][]string{nil, {"argument"}, {"child"}} {
		err := root.Parse(arguments)
		if err == nil {
			t.Fatalf("Parse(%v) succeeded, want an error", arguments)
		}
		requireContains(t, err.Error(), "cannot have both subcommands and positional arguments")
	}
}

func TestExecuteDispatchesNestedCommandAndParsesParentFlags(t *testing.T) {
	var name string
	var leafRan bool

	root := cmd.New("root", "")
	root.Flags.String("name", &name, "Name")

	middle := cmd.New("middle", "")
	leaf := cmd.New("leaf", "")
	leaf.Runner(func() error {
		leafRan = true
		return nil
	})
	middle.Subcommand(leaf)
	root.Subcommand(middle)

	if err := root.Execute([]string{"--name", "Locke", "middle", "leaf"}); err != nil {
		t.Fatalf("Execute() returned an error: %v", err)
	}
	if name != "Locke" {
		t.Errorf("parsed name = %q, want %q", name, "Locke")
	}
	if !leafRan {
		t.Fatal("nested command runner was not called")
	}
}

func TestExecuteParseErrorDoesNotRunCommand(t *testing.T) {
	var childRan bool
	root := cmd.New("root", "")
	child := cmd.New("child", "")
	child.Runner(func() error {
		childRan = true
		return nil
	})
	root.Subcommand(child)

	if err := root.Execute([]string{"unknown", "child"}); err == nil {
		t.Fatal("Execute() succeeded for invalid arguments, want an error")
	}
	if childRan {
		t.Fatal("command runner was called after a parse error")
	}
}

func TestExecuteReturnsRunnerError(t *testing.T) {
	want := errors.New("runner failed")
	root := cmd.New("root", "")
	root.Runner(func() error {
		return want
	})

	if err := root.Execute(nil); !errors.Is(err, want) {
		t.Fatalf("Execute() error = %v, want %v", err, want)
	}
}

func TestExecuteRequiresRunner(t *testing.T) {
	root := cmd.New("root", "")

	err := root.Execute(nil)
	if err == nil {
		t.Fatal("Execute() succeeded without a runner, want an error")
	}
	requireContains(t, err.Error(), `command "root" has no runner`)
}

func TestParseDoesNotRunCommand(t *testing.T) {
	var runnerCalled bool
	root := cmd.New("root", "")
	root.Runner(func() error {
		runnerCalled = true
		return nil
	})

	if err := root.Parse(nil); err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if runnerCalled {
		t.Fatal("Parse() called the command runner")
	}
}
