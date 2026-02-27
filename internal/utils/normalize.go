package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// NormalizeString normalizes a string for matching purposes
// - Converts to lowercase
// - Removes accents/diacritics
// - Removes special characters
// - Removes "feat.", "ft.", "featuring" variations
// - Trims spaces
func NormalizeString(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove accents/diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)

	// Remove feat variations
	featPatterns := []string{
		`\s*\(?feat\.?\s+[^)]*\)?`,
		`\s*\(?ft\.?\s+[^)]*\)?`,
		`\s*\(?featuring\s+[^)]*\)?`,
		`\s*\(?with\s+[^)]*\)?`,
		`\s*&\s*`,
	}
	for _, pattern := range featPatterns {
		re := regexp.MustCompile(pattern)
		s = re.ReplaceAllString(s, " ")
	}

	// Remove all special characters except spaces
	re := regexp.MustCompile(`[^a-z0-9\s]`)
	s = re.ReplaceAllString(s, " ")

	// Replace multiple spaces with single space
	re = regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")

	// Trim spaces
	s = strings.TrimSpace(s)

	return s
}

// NormalizeDuration returns the min and max duration range with tolerance
// tolerance is in seconds (e.g., 2 means ±2 seconds)
func NormalizeDuration(duration int, tolerance int) (min, max int) {
	min = duration - tolerance
	max = duration + tolerance

	// Ensure min is not negative
	if min < 0 {
		min = 0
	}

	return min, max
}

// CompareDurations checks if two durations match within the given tolerance
// d1 and d2 are durations in seconds
// tolerance is in seconds (e.g., 2 means ±2 seconds)
func CompareDurations(d1, d2, tolerance int) bool {
	diff := d1 - d2
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

// ExtractArtistName extracts the primary artist name from a string
// For example: "Kalash ft Satori" -> "Kalash"
func ExtractArtistName(s string) string {
	s = strings.ToLower(s)

	// Split by common separators
	separators := []string{" feat ", " feat. ", " ft ", " ft. ", " featuring ", " with ", " & ", " x "}

	for _, sep := range separators {
		if idx := strings.Index(s, sep); idx != -1 {
			return strings.TrimSpace(s[:idx])
		}
	}

	return strings.TrimSpace(s)
}

// ExtractFeaturedArtists extracts all featured artist names from a string
// For example: "Kalash ft Satori & Damso" -> ["Satori", "Damso"]
func ExtractFeaturedArtists(s string) []string {
	s = strings.ToLower(s)
	featured := []string{}

	// Patterns to identify featured artists
	patterns := []string{
		` feat\.?\s+`,
		` ft\.?\s+`,
		` featuring\s+`,
		` with\s+`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if loc := re.FindStringIndex(s); loc != nil {
			// Extract everything after the pattern
			remainder := s[loc[1]:]

			// Remove parentheses
			remainder = strings.Trim(remainder, "()")

			// Split by & or , or "and"
			artistsStr := regexp.MustCompile(`\s*[&,]\s*|\s+and\s+`).Split(remainder, -1)

			for _, artist := range artistsStr {
				artist = strings.TrimSpace(artist)
				if artist != "" {
					featured = append(featured, artist)
				}
			}
			break
		}
	}

	return featured
}
