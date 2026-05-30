package cowsay

import (
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed cows/*.cow
var cowFS embed.FS

// ListCows returns every embedded cow name (without the .cow extension),
// sorted alphabetically. Used by Phase 2 to implement the -l flag.
func ListCows() ([]string, error) {
	entries, err := cowFS.ReadDir("cows")
	if err != nil {
		return nil, fmt.Errorf("listing embedded cows: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		n := e.Name()
		if !strings.HasSuffix(n, ".cow") {
			continue
		}
		names = append(names, strings.TrimSuffix(n, ".cow"))
	}
	sort.Strings(names)
	return names, nil
}

// readCowFile returns the raw bytes of a cow file from the embedded FS.
// name is the cow name WITHOUT the .cow suffix.
// Plan 01-03 cowfile.go calls this function.
func readCowFile(name string) ([]byte, error) {
	return cowFS.ReadFile("cows/" + name + ".cow")
}
