package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: tally FILE...")
		os.Exit(2)
	}
	for _, path := range os.Args[1:] {
		lines, words, bytes, err := count(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tally: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%8d %8d %8d %s\n", lines, words, bytes, path)
	}
}

func count(path string) (lines, words, bytes int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		lines++
		words += len(strings.Fields(line))
		bytes += len(line) + 1
	}
	return lines, words, bytes, sc.Err()
}
