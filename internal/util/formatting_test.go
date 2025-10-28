package util

import "testing"

func TestFormatCompletion(t *testing.T) {
	res := FormatCompletion([]string{"header1\thead2", "comp1\tc1", "c2\tc22"})
	expected := `header1 head2
comp1   c1
c2      c22
`
	if res != expected {
		t.Errorf("FormatCompletion() = %q, want %q", res, expected)
	}
}
