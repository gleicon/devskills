package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: greet <name>")
		os.Exit(1)
	}
	fmt.Println(greeting(os.Args[1]))
}
