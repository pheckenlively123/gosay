package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pheckenlively/gosay/internal/cowsay"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gosay <message>")
		os.Exit(1)
	}
	message := strings.Join(os.Args[1:], " ")
	out, err := cowsay.Render("gopher", message, cowsay.RenderOpts{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(out)
}
