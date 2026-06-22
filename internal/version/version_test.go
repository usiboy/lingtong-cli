// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package version

import "testing"

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		// Equal, with/without prefix and whitespace.
		{"1.0.0", "1.0.0", 0},
		{"v1.0.0", "1.0.0", 0},
		{"V1.0.0", "v1.0.0", 0},
		{" 1.0.0 ", "1.0.0", 0},

		// a < b
		{"1.0.0", "1.0.1", -1},
		{"1.0.0", "1.1.0", -1},
		{"1.0.0", "2.0.0", -1},
		{"1.2.3", "1.2.4", -1},
		{"1.9.0", "1.10.0", -1}, // numeric, not lexical

		// a > b
		{"1.0.1", "1.0.0", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.10.0", "1.9.0", 1},

		// Differing lengths: missing segments are zero.
		{"1.0", "1.0.0", 0},
		{"1.0", "1.0.1", -1},
		{"1.0.1", "1.0", 1},

		// Pre-release ordering: release outranks pre-release.
		{"1.2.0-beta", "1.2.0", -1},
		{"1.2.0", "1.2.0-beta", 1},
		{"1.2.0-alpha", "1.2.0-beta", -1},
		{"1.2.0-rc.1", "1.2.0-rc.2", -1},
		{"1.2.0-rc.2", "1.2.0-rc.10", -1}, // numeric pre-release segment
		{"1.2.0-alpha", "1.2.0-alpha.1", -1},
		{"1.2.0-beta", "1.2.0-beta", 0},

		// Build metadata is ignored.
		{"1.2.0+build1", "1.2.0+build2", 0},
		{"1.2.0+build", "1.2.0", 0},

		// Non-numeric / garbage segments degrade to zero.
		{"1.x.0", "1.0.0", 0},
		{"", "", 0},
		{"", "0.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			if got := Compare(tt.a, tt.b); got != tt.want {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCompareIsAntisymmetric(t *testing.T) {
	pairs := [][2]string{
		{"1.0.0", "1.0.1"},
		{"2.0.0", "1.9.9"},
		{"1.2.0-rc", "1.2.0"},
	}
	for _, p := range pairs {
		ab := Compare(p[0], p[1])
		ba := Compare(p[1], p[0])
		if ab != -ba {
			t.Errorf("Compare(%q,%q)=%d but Compare(%q,%q)=%d (not antisymmetric)",
				p[0], p[1], ab, p[1], p[0], ba)
		}
	}
}
