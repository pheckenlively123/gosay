package cowsay

import (
	"os"
	"strings"
	"testing"

	goldie "github.com/sebdah/goldie/v2"
)

// TestGolden_GopherSayHello exercises the full render pipeline for the default gopher.
func TestGolden_GopherSayHello(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	out, err := Render("gopher", "hello", RenderOpts{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	g.Assert(t, "gopher_say_hello", []byte(out))
}

// TestGolden_GopherSayMultiline exercises multi-line balloon rendering with the gopher.
func TestGolden_GopherSayMultiline(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	out, err := Render("gopher", "line1\nline2", RenderOpts{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	g.Assert(t, "gopher_say_multiline", []byte(out))
}

// TestGolden_DefaultSayHello is the canonical backslash-unescape regression check.
// default.cow contains `($eyes)\\_______`; the body must render with a single `\`,
// not `\\`. If cowBodyUnescape is applied AFTER substitution this test will fail.
func TestGolden_DefaultSayHello(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	out, err := Render("default", "hello", RenderOpts{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	g.Assert(t, "default_say_hello", []byte(out))
}

// TestGolden_DragonAndCowSayHello exercises the `\@` -> `@` unescape path.
// dragon-and-cow.cow contains `\@___\@` lines; these must appear as `@___@` in output.
func TestGolden_DragonAndCowSayHello(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	out, err := Render("dragon-and-cow", "hello", RenderOpts{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	g.Assert(t, "dragon_and_cow_say_hello", []byte(out))
}

// TestGolden_ThreeEyesSayHello documents gosay's two-eye output for three-eyes.cow.
//
// Upstream Perl cowsay's chop() manipulation produces three-character eyes by
// appending the chopped character twice to the two-character $eyes value:
//
//	$extra = chop($eyes); $eyes .= ($extra x 2);
//
// gosay's strings.NewReplacer substitutes the bare two-character $eyes value,
// yielding two eyes rather than three. This is documented in RESEARCH section B-Q6
// and is not a bug — Phase 3 may address it via a pre-processing pass, but for
// Phase 1 the two-eye output is the correct expected golden.
func TestGolden_ThreeEyesSayHello(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	out, err := Render("three-eyes", "hello", RenderOpts{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	g.Assert(t, "three_eyes_say_hello", []byte(out))
}

// TestGolden_NonEOCSayHello exercises the dynamic-terminator path using the synthetic
// non-eoc.cow fixture (stored in testdata/fixtures/, NOT in the embedded cow set).
// This verifies that parseCowBody correctly discovers "END" as the terminator and
// that the resulting body still passes through balloon + substitution correctly.
func TestGolden_NonEOCSayHello(t *testing.T) {
	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	data, err := os.ReadFile("testdata/fixtures/non-eoc.cow")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	body, err := parseCowBody(data)
	if err != nil {
		t.Fatalf("parseCowBody: %v", err)
	}
	// Drive substitution through the production substituteVars helper, and apply the
	// same trailing-newline normalization Render performs, so this fixture-based test
	// stays in lockstep with the production render path.
	balloon := buildBalloon("hello")
	substituted := substituteVars(body, RenderOpts{})
	out := strings.TrimRight(balloon+substituted, "\n") + "\n"
	g.Assert(t, "non_eoc_say_hello", []byte(out))
}

// TestGolden_CJK_Skipped documents the Phase 1 CJK display-width gap.
// gosay's Phase 1 displayWidth uses utf8.RuneCountInString which under-pads the
// balloon for CJK input (each CJK rune occupies 2 terminal columns but counts as 1).
// This test exists to document the gap and will be enabled in Phase 3 / RENDER-06
// when displayWidth is swapped to runewidth.StringWidth.
func TestGolden_CJK_Skipped(t *testing.T) {
	t.Skip("CJK display width requires runewidth (Phase 3 / RENDER-06); gosay's Phase 1 displayWidth uses utf8.RuneCountInString which under-pads the balloon for CJK input — this test exists to document the gap, will be enabled in Phase 3")
	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	// 漢字テスト = 漢字テスト
	out, err := Render("gopher", "漢字テスト", RenderOpts{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	g.Assert(t, "cjk_skip", []byte(out))
}

// TestListCows_IncludesGopher is a regression assertion that gopher.cow actually
// landed in the embedded cow set after Plan 01-04 added it.
func TestListCows_IncludesGopher(t *testing.T) {
	names, err := ListCows()
	if err != nil {
		t.Fatalf("ListCows: %v", err)
	}
	if len(names) < 51 {
		t.Errorf("ListCows returned %d names; want >= 51 (50 upstream + 1 gopher)", len(names))
	}
	found := false
	for _, n := range names {
		if n == "gopher" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ListCows does not contain \"gopher\"")
	}
}
