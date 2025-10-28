package util

import "testing"

func TestDumpLineSanitizesControlCharacters(t *testing.T) {
	input := []string{"alpha", "beta\x1b[31m", ""}
	got := DumpLine(input)
	expected := "alpha\tbeta\tNone"
	if got != expected {
		t.Fatalf("DumpLine() = %q, want %q", got, expected)
	}
}

func TestSanitizeControlCharacters(t *testing.T) {
	cases := map[string]string{
		"plain":             "plain",
		"contains\x00null":  "containsnull",
		"escape\x1b[1mcode": "escapecode",
		"":                  "",
	}

	for input, want := range cases {
		if got := sanitizeControlCharacters(input); got != want {
			t.Fatalf("sanitizeControlCharacters(%q) = %q, want %q", input, got, want)
		}
	}
}
