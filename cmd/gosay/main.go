package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
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

// formatCowList formats a sorted slice of cow names into the canonical -l output:
// a "Cow files:" header followed by names joined with a single space, wrapped at
// approximately 76 characters per line (matching the upstream cowsay column shape).
func formatCowList(names []string) string {
	const wrapWidth = 76
	var sb strings.Builder
	sb.WriteString("Cow files:\n")
	line := ""
	for _, name := range names {
		if line == "" {
			line = name
		} else if len(line)+1+len(name) > wrapWidth {
			sb.WriteString(line)
			sb.WriteByte('\n')
			line = name
		} else {
			line += " " + name
		}
	}
	if line != "" {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	return sb.String()
}

const helpText = `gosay — make a gopher say something

Usage: gosay [flags] [message...]

Flags:
  -e <eyes>     Set eye characters (default "oo")
  -f <name>     Select animal from embedded set (default "gopher")
  -l            List available animals
  -n            Disable word wrapping (preserve all input whitespace; overrides -W)
  -T <tongue>   Set tongue characters (default "  ")
  -W <cols>     Wrap message at this many display columns (default 40)
  --random      Pick a random animal
  --think       Use thought bubble ( ) instead of speech bubble < >

Examples:
  gosay hello
  echo hi | gosay -f tux
  gosay --think -e "^^" "thinking..."`

// run assembles the message from args or stdin, renders it, and writes the result.
// It returns the process exit code so the logic is testable without os.Exit.
// args is os.Args[1:] (the program arguments, excluding the binary name).
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gosay", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var cowName string
	var listFlag bool
	var randomFlag bool
	var wrapWidth int
	var noWrap bool
	var think bool
	var eyes string
	var tongue string
	// NOTE: -e/-T/-W register empty-string/zero defaults as "not set" sentinels;
	// the real user-facing defaults ("oo", "  ", 40 cols) are resolved downstream
	// in substituteVars/Render. The help-text strings above describe those effective
	// defaults, which is why they intentionally differ from the values registered here.
	fs.StringVar(&cowName, "f", "gopher", "cow `name` to use")
	fs.BoolVar(&listFlag, "l", false, "list available cows")
	fs.BoolVar(&randomFlag, "random", false, "pick a random cow")
	fs.IntVar(&wrapWidth, "W", 0, "wrap message at this many display `cols` (default 40)")
	fs.BoolVar(&noWrap, "n", false, "disable word wrapping")
	fs.BoolVar(&think, "think", false, "use thought bubble instead of speech bubble")
	fs.StringVar(&eyes, "e", "", "set eye characters (default \"oo\")")
	fs.StringVar(&tongue, "T", "", "set tongue characters (default \"  \")")

	// CRITICAL: no-op before Parse to suppress automatic stderr print on -h/--help.
	// Without this, flag.FlagSet prints usage to stderr before returning flag.ErrHelp,
	// causing help to bleed onto stderr (RESEARCH Pattern 4, Pitfall 2).
	fs.Usage = func() {}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(stdout, helpText)
			return 0
		}
		fmt.Fprintln(stderr, "usage: gosay [flags] [message...]")
		return 1
	}

	// Detect whether -f / -W were explicitly set (vs. their zero/default values).
	fExplicit := false
	wExplicit := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "f":
			fExplicit = true
		case "W":
			wExplicit = true
		}
	})

	// Reject an explicitly non-positive -W. An unset -W (0) is the "use default 40"
	// sentinel resolved in Render; but a user who explicitly asks for -W 0 or a
	// negative width gets a usage error rather than a silent fallback to 40.
	if wExplicit && wrapWidth < 1 {
		fmt.Fprintln(stderr, "gosay: -W must be a positive number of columns")
		return 1
	}

	// D-06: -f + --random together is a usage error.
	if randomFlag && fExplicit {
		fmt.Fprintln(stderr, "gosay: cannot combine -f and --random")
		return 1
	}

	// D-09: -l cannot be combined with a message or animal selection.
	if listFlag && (len(fs.Args()) > 0 || randomFlag || fExplicit) {
		fmt.Fprintln(stderr, "gosay: -l cannot be combined with a message or animal selection")
		return 1
	}

	// -l branch: print the columnar cow listing and exit.
	if listFlag {
		names, err := cowsay.ListCows()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprint(stdout, formatCowList(names))
		return 0
	}

	// Determine the animal: --random picks from the full ListCows() pool.
	animal := cowName
	if randomFlag {
		names, err := cowsay.ListCows()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		animal = names[rand.Intn(len(names))]
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
			fmt.Fprintln(stderr, "usage: gosay [flags] [message...]")
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

	opts := cowsay.RenderOpts{
		Eyes:   eyes,
		Tongue: tongue,
		Width:  wrapWidth,
		NoWrap: noWrap,
		Think:  think,
	}
	out, err := cowsay.Render(animal, message, opts)
	if err != nil {
		if errors.Is(err, cowsay.ErrUnknownCow) {
			fmt.Fprintf(stderr, "gosay: unknown cowfile %q\n", animal)
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
