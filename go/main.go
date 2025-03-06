package main

import (
	"fmt"
	"strings"

	"github.com/mooss/bagend/flag"
)

var parser = flag.NewParser()

func flg[D flag.Decoder[T], T any](name string, dest *T, doc string) flag.FluentFlag[T] {
	return flag.Register[D](parser, name, dest, doc)
}

func sflg[D flag.Decoder[T], T any](name string, dest *[]T, doc string) flag.FluentFlag[[]T] {
	return flag.RegisterSlice[D](parser, name, dest, doc)
}

func main() {
	var (
		twentythree int
		eight       []int
		four        string
	)

	flg[flag.Int]("twentythree", &twentythree, "23").Default(23).Alias("23")
	sflg[flag.Int]("eight", &eight, "8").Default([]int{8}).Alias("8")
	flg[flag.String]("four", &four, "8").Default("4").Alias("4")

	if err := parser.Parse(strings.Split("4 -8 15 16 --23 42", " ")); err != nil {
		panic(err)
	}

	fmt.Println(":23", twentythree)
	fmt.Println(":8", eight)
	fmt.Println(":4", four)
}
