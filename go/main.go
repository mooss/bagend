package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mooss/bagend/go/cmd"
	"github.com/mooss/bagend/go/flag"
)

func noerr(err error) {
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
}

func newGreetCommand(dest *string) *cmd.Command {
	greet := &cmd.Command{
		Name:    "greet",
		Docline: "Print a greeting",
		Flags:   flag.NewParser(),
	}
	greet.Aliases = append(greet.Aliases, "gr")
	greet.FlagsParser().String("name", dest, "Who to greet").Default("world")
	greet.Runner = func() error {
		fmt.Println("greet:", *dest)
		return nil
	}
	return greet
}

func newSumCommand(dest *[]int) *cmd.Command {
	sum := &cmd.Command{
		Name:    "sum",
		Docline: "Sum numbers",
		Flags:   flag.NewParser(),
	}
	sum.FlagsParser().IntSlice("numbers", dest, "Numbers to sum")
	sum.Runner = func() error {
		fmt.Println("sum:", *dest)
		return nil
	}
	return sum
}

func main() {
	var (
		twentythree int
		eight       []int
		four        string
		hatch       bool
		name        string
		numbers     []int
		positional  []string
	)

	root := &cmd.Command{
		Name:  os.Args[0],
		Flags: flag.NewParser(),
	}
	root.Usage = "POSITIONAL [FLAGS]"
	rootf := root.FlagsParser().Positional(&positional)
	rootf.Int("twentythree", &twentythree, "Shephard").Default(23).Alias("23")
	rootf.IntSlice("eight", &eight, "Reyes").Default([]int{8}).Alias("8")
	rootf.String("four", &four, "Locke").Default("4").Alias("4")
	rootf.Bool("hatch", &hatch, "The hatch")

	root.Runner = func() error {
		fmt.Println(":23", twentythree)
		fmt.Println(":8", eight)
		fmt.Println(":4", four)
		fmt.Println(":hatch", hatch)
		fmt.Println(":positional", positional)
		return nil
	}

	noerr(root.Execute(strings.Split("4 -8 15 16 --23 42 --hatch 3", " ")))

	second := &cmd.Command{
		Name:    "second",
		Docline: "Subcommand example",
		Flags:   flag.NewParser(),
	}
	second.Description = `A command to showcase subcommands.

some

empty

...

space.`
	second.RegisterSub(newGreetCommand(&name))
	second.RegisterSub(newSumCommand(&numbers))

	noerr(second.Execute([]string{"greet", "--name", "Locke"}))
	noerr(second.Execute([]string{"sum", "--numbers", "4", "8"}))

	second.Usage = "not Penny's boat"
	second.FlagsParser().EnableHelp()
	noerr(second.Execute([]string{"--help", "sum", "--numbers", "4", "8"}))
}
