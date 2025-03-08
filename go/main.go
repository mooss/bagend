package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mooss/bagend/go/flag"
)

func noerr(err error) {
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
}

func main() {
	var (
		twentythree int
		eight       []int
		four        string
		hatch       bool
	)

	parser := flag.NewParser(flag.WithHelp(os.Args[0], "POSITIONAL [FLAGS]"))
	parser.Int("twentythree", &twentythree, "Shephard").Default(23).Alias("23")
	parser.IntSlice("eight", &eight, "Reyes").Default([]int{8}).Alias("8")
	parser.String("four", &four, "Locke").Default("4").Alias("4")
	parser.Bool("hatch", &hatch, "The hatch")

	noerr(parser.Parse(strings.Split("4 -8 15 16 --23 42 --hatch 3", " ")))

	fmt.Println(":23", twentythree)
	fmt.Println(":8", eight)
	fmt.Println(":4", four)
	fmt.Println(":hatch", hatch)
	fmt.Println(":positional", parser.Positional)

	noerr(parser.Parse([]string{"-h"}))
}
