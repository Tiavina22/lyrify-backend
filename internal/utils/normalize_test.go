package utils

import (
	"testing"
)

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple lowercase",
			input:    "Kalash",
			expected: "kalash",
		},
		{
			name:     "Remove accents",
			input:    "Beyoncé",
			expected: "beyonce",
		},
		{
			name:     "Remove feat variations",
			input:    "Kalash Ft. Satori",
			expected: "kalash",
		},
		{
			name:     "Remove featuring",
			input:    "Damso (feat. Siboy)",
			expected: "damso",
		},
		{
			name:     "Remove special characters",
			input:    "Artist-Name (2023)",
			expected: "artist name 2023",
		},
		{
			name:     "Complex case",
			input:    "KALASH ft. Satori & Damso",
			expected: "kalash",
		},
		{
			name:     "Multiple spaces",
			input:    "Artist    Name",
			expected: "artist name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeString(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompareDurations(t *testing.T) {
	tests := []struct {
		name      string
		d1        int
		d2        int
		tolerance int
		expected  bool
	}{
		{
			name:      "Exact match",
			d1:        300,
			d2:        300,
			tolerance: 2,
			expected:  true,
		},
		{
			name:      "Within tolerance (positive)",
			d1:        300,
			d2:        301,
			tolerance: 2,
			expected:  true,
		},
		{
			name:      "Within tolerance (negative)",
			d1:        301,
			d2:        300,
			tolerance: 2,
			expected:  true,
		},
		{
			name:      "At tolerance boundary",
			d1:        300,
			d2:        302,
			tolerance: 2,
			expected:  true,
		},
		{
			name:      "Outside tolerance",
			d1:        300,
			d2:        303,
			tolerance: 2,
			expected:  false,
		},
		{
			name:      "Large difference",
			d1:        345, // 5m45
			d2:        358, // 5m58
			tolerance: 2,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareDurations(tt.d1, tt.d2, tt.tolerance)
			if result != tt.expected {
				t.Errorf("CompareDurations(%d, %d, %d) = %v, want %v",
					tt.d1, tt.d2, tt.tolerance, result, tt.expected)
			}
		})
	}
}

func TestNormalizeDuration(t *testing.T) {
	tests := []struct {
		name        string
		duration    int
		tolerance   int
		expectedMin int
		expectedMax int
	}{
		{
			name:        "Normal case",
			duration:    300,
			tolerance:   2,
			expectedMin: 298,
			expectedMax: 302,
		},
		{
			name:        "Zero duration",
			duration:    0,
			tolerance:   2,
			expectedMin: 0,
			expectedMax: 2,
		},
		{
			name:        "Small duration",
			duration:    5,
			tolerance:   2,
			expectedMin: 3,
			expectedMax: 7,
		},
		{
			name:        "Large duration",
			duration:    3600,
			tolerance:   5,
			expectedMin: 3595,
			expectedMax: 3605,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := NormalizeDuration(tt.duration, tt.tolerance)
			if min != tt.expectedMin || max != tt.expectedMax {
				t.Errorf("NormalizeDuration(%d, %d) = (%d, %d), want (%d, %d)",
					tt.duration, tt.tolerance, min, max, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestExtractArtistName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single artist",
			input:    "Kalash",
			expected: "kalash",
		},
		{
			name:     "With feat",
			input:    "Kalash feat Satori",
			expected: "kalash",
		},
		{
			name:     "With ft.",
			input:    "Kalash ft. Satori",
			expected: "kalash",
		},
		{
			name:     "With featuring",
			input:    "Kalash featuring Satori",
			expected: "kalash",
		},
		{
			name:     "With ampersand",
			input:    "Kalash & Satori",
			expected: "kalash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractArtistName(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractArtistName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractFeaturedArtists(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "No featured artists",
			input:    "Kalash",
			expected: []string{},
		},
		{
			name:     "One featured artist",
			input:    "Kalash feat Satori",
			expected: []string{"satori"},
		},
		{
			name:     "Multiple featured artists with comma",
			input:    "Kalash ft. Satori, Damso",
			expected: []string{"satori", "damso"},
		},
		{
			name:     "Multiple featured artists with ampersand",
			input:    "Kalash featuring Satori & Damso",
			expected: []string{"satori", "damso"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFeaturedArtists(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ExtractFeaturedArtists(%q) returned %d artists, want %d",
					tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("ExtractFeaturedArtists(%q)[%d] = %q, want %q",
						tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}
