package cowsay

import (
	"sort"
	"testing"
)

// TestListCows_CountAndSort verifies that ListCows returns at least 50 names,
// that the returned slice is sorted alphabetically, and that names do not carry
// the .cow suffix.  The floor is >= 50 because Plan 01-04 will add gopher.cow,
// bringing the total to 51.
func TestListCows_CountAndSort(t *testing.T) {
	names, err := ListCows()
	if err != nil {
		t.Fatalf("ListCows() error: %v", err)
	}

	if len(names) < 50 {
		t.Errorf("expected >= 50 cow names, got %d", len(names))
	}

	// Verify alphabetical sort.
	sorted := make([]string, len(names))
	copy(sorted, names)
	sort.Strings(sorted)
	for i := range names {
		if names[i] != sorted[i] {
			t.Errorf("ListCows() is not sorted: names[%d]=%q, want %q", i, names[i], sorted[i])
		}
	}

	// Verify no .cow suffix.
	for _, n := range names {
		if len(n) > 4 && n[len(n)-4:] == ".cow" {
			t.Errorf("cow name %q still has .cow suffix", n)
		}
	}
}

// TestListCows_KeyFilesPresent verifies that specific cow files expected in the
// embedded set are actually present.  daemon is included to verify that the
// politically-charged file is not accidentally omitted (see NOTICE for provenance).
func TestListCows_KeyFilesPresent(t *testing.T) {
	names, err := ListCows()
	if err != nil {
		t.Fatalf("ListCows() error: %v", err)
	}

	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}

	required := []string{"daemon", "default", "tux", "three-eyes", "dragon-and-cow"}
	for _, want := range required {
		if !nameSet[want] {
			t.Errorf("expected cow %q to be in embedded set, but it was not found", want)
		}
	}
}
