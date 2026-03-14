package cli

import "testing"

func TestDefaultLanguageIndex(t *testing.T) {
	if got := defaultLanguageIndex("zh"); got != 0 {
		t.Fatalf("expected zh index 0, got %d", got)
	}
	if got := defaultLanguageIndex("en"); got != 1 {
		t.Fatalf("expected en index 1, got %d", got)
	}
	if got := defaultLanguageIndex(""); got != 1 {
		t.Fatalf("expected default index 1, got %d", got)
	}
}

func TestWrapIndex(t *testing.T) {
	if got := wrapIndex(-1, 2); got != 1 {
		t.Fatalf("expected wrapped index 1, got %d", got)
	}
	if got := wrapIndex(2, 2); got != 0 {
		t.Fatalf("expected wrapped index 0, got %d", got)
	}
}

func TestRenderSelectorLine(t *testing.T) {
	if got := renderSelectorLine("中文", true); got != "\033[7m中文\033[0m" {
		t.Fatalf("unexpected selected line %q", got)
	}
	if got := renderSelectorLine("English", false); got != "English" {
		t.Fatalf("unexpected unselected line %q", got)
	}
}
