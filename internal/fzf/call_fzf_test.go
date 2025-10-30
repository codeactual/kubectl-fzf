package fzf

import "testing"

func TestNameColumnIndex(t *testing.T) {
	tests := map[string]struct {
		header    string
		expectIdx int
		found     bool
	}{
		"standard": {
			header:    "NAMESPACE NAME READY",
			expectIdx: 2,
			found:     true,
		},
		"lowercase": {
			header:    "namespace name age",
			expectIdx: 2,
			found:     true,
		},
		"missing": {
			header:    "namespace label",
			expectIdx: 0,
			found:     false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			idx, ok := nameColumnIndex(tc.header)
			if ok != tc.found {
				t.Fatalf("expected found=%t, got %t", tc.found, ok)
			}
			if idx != tc.expectIdx {
				t.Fatalf("expected index %d, got %d", tc.expectIdx, idx)
			}
		})
	}
}
