package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pheckenlively/gosay/internal/cowsay"
)

// run assembles the message from args, renders it, and writes the result.
// It returns the process exit code so the logic is testable without os.Exit.
// args is os.Args[1:] (the program arguments, excluding the binary name).
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: gosay <message>")
		return 1
	}
	message := strings.Join(args, " ")
	out, err := cowsay.Render("gopher", message, cowsay.RenderOpts{})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprint(stdout, out)
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
