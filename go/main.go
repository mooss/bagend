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
	greet := cmd.New("greet", "Print a greeting")
	greet.Parser().String("name", dest, "Who to greet").Default("world")
	return greet
}

func newSumCommand(dest *[]int) *cmd.Command {
	sum := cmd.New("sum", "Sum numbers")
	sum.Parser().IntSlice("numbers", dest, "Numbers to sum")
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
	)

	root := cmd.New(os.Args[0], "", flag.WithHelp(os.Args[0], "POSITIONAL [FLAGS]"))
	parser := root.Parser()
	parser.Int("twentythree", &twentythree, "Shephard").Default(23).Alias("23")
	parser.IntSlice("eight", &eight, "Reyes").Default([]int{8}).Alias("8")
	parser.String("four", &four, "Locke").Default("4").Alias("4")
	parser.Bool("hatch", &hatch, "The hatch")

	root.Add(newGreetCommand(&name))
	root.Add(newSumCommand(&numbers))

	noerr(root.Parse(strings.Split("4 -8 15 16 --23 42 --hatch 3", " ")))

	fmt.Println(":23", twentythree)
	fmt.Println(":8", eight)
	fmt.Println(":4", four)
	fmt.Println(":hatch", hatch)
	fmt.Println(":positional", parser.Positional)

	noerr(root.Parse([]string{"greet", "--name", "Locke"}))
	fmt.Println("greet:", name)

	noerr(root.Parse([]string{"sum", "--numbers", "4", "8"}))
	fmt.Println("sum:", numbers)

	noerr(root.Parse([]string{"-h"}))
}
