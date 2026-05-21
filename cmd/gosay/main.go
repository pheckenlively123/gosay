package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gosay <message>")
		os.Exit(1)
	}
	message := strings.Join(os.Args[1:], " ")
	fmt.Printf("gosay (walking skeleton): would render gopher saying: %s\n", message)
}
