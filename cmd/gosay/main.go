package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pheckenlively/gosay/internal/cowsay"
)

// isTTY reports whether f is connected to a character device (interactive terminal).
func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// run assembles the message from args or stdin, renders it, and writes the result.
// It returns the process exit code so the logic is testable without os.Exit.
// args is os.Args[1:] (the program arguments, excluding the binary name).
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gosay", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var cowName string
	fs.StringVar(&cowName, "f", "gopher", "cow `name` to use")

	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: gosay [-f name] [message...]")
	}

	if err := fs.Parse(args); err != nil {
		return 1
	}

	// Resolve the message: positional args win over stdin.
	var message string
	if fs.NArg() > 0 {
		// Positional args present — join them and do NOT read stdin.
		message = strings.Join(fs.Args(), " ")
	} else {
		// No positional args — check whether stdin is piped or a TTY.
		if isTTY(os.Stdin) {
			// Interactive terminal with no args: print usage and exit.
			fmt.Fprintln(stderr, "usage: gosay [-f name] [message...]")
			return 1
		}
		// Stdin is piped: read it.
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		// Trim exactly one trailing newline to match upstream echo behavior.
		message = strings.TrimSuffix(string(data), "\n")
	}

	out, err := cowsay.Render(cowName, message, cowsay.RenderOpts{})
	if err != nil {
		if errors.Is(err, cowsay.ErrUnknownCow) {
			fmt.Fprintf(stderr, "gosay: unknown cowfile %q\n", cowName)
			return 1
		}
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprint(stdout, out)
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
