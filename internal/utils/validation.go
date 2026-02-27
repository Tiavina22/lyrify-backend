package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Timestamp represents a single LRC timestamp with its corresponding text
type Timestamp struct {
	Time int64  `json:"time"` // Time in milliseconds
	Text string `json:"text"` // Lyric text for this timestamp
}

var (
	// LRC line format: [mm:ss.xx] text or [mm:ss] text
	lrcLineRegex = regexp.MustCompile(`^\[(\d{1,2}):(\d{2})(?:\.(\d{2,3}))?\]\s*(.*)$`)
	// LRC metadata format: [ar:Artist] or [ti:Title] etc.
	lrcMetaRegex = regexp.MustCompile(`^\[([a-z]{2}):(.*)\]$`)
)

// ValidateLRCFormat validates that the content is in proper LRC format
// LRC format: [mm:ss.xx] line of text
// Also supports metadata tags like [ar:Artist], [ti:Title], [al:Album]
func ValidateLRCFormat(content string) error {
	if strings.TrimSpace(content) == "" {
		return errors.New("LRC content cannot be empty")
	}

	lines := strings.Split(content, "\n")
	hasTimestamp := false

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check if it's a metadata tag
		if lrcMetaRegex.MatchString(line) {
			continue
		}

		// Check if it's a timestamp line
		if lrcLineRegex.MatchString(line) {
			hasTimestamp = true

			// Validate timestamp values
			matches := lrcLineRegex.FindStringSubmatch(line)
			if len(matches) >= 3 {
				minutes, _ := strconv.Atoi(matches[1])
				seconds, _ := strconv.Atoi(matches[2])

				if minutes < 0 || minutes > 99 {
					return fmt.Errorf("invalid minutes value at line %d: %d", i+1, minutes)
				}
				if seconds < 0 || seconds > 59 {
					return fmt.Errorf("invalid seconds value at line %d: %d", i+1, seconds)
				}
			}
			continue
		}

		// If line doesn't match any format, it's invalid
		return fmt.Errorf("invalid LRC format at line %d: %s", i+1, line)
	}

	if !hasTimestamp {
		return errors.New("LRC content must contain at least one timestamp")
	}

	return nil
}

// CleanLRCContent removes non-LRC lines from content (useful when copy-pasting from websites)
// Keeps only lines with timestamps [mm:ss.xx] and metadata tags [ar:Artist]
func CleanLRCContent(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	cleanedLines := []string{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Keep metadata tags [ar:Artist], [ti:Title], etc.
		if lrcMetaRegex.MatchString(line) {
			cleanedLines = append(cleanedLines, line)
			continue
		}

		// Keep timestamp lines [mm:ss.xx] text
		if lrcLineRegex.MatchString(line) {
			cleanedLines = append(cleanedLines, line)
			continue
		}

		// Skip all other lines (HTML, navigation, etc.)
	}

	return strings.Join(cleanedLines, "\n")
}

// ExtractLRCTimestamps parses LRC content and extracts all timestamps with their text
func ExtractLRCTimestamps(content string) ([]Timestamp, error) {
	if err := ValidateLRCFormat(content); err != nil {
		return nil, err
	}

	timestamps := []Timestamp{}
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and metadata
		if line == "" || lrcMetaRegex.MatchString(line) {
			continue
		}

		// Extract timestamp
		if matches := lrcLineRegex.FindStringSubmatch(line); len(matches) >= 4 {
			minutes, _ := strconv.Atoi(matches[1])
			seconds, _ := strconv.Atoi(matches[2])

			// Handle milliseconds (can be 2 or 3 digits)
			milliseconds := 0
			if matches[3] != "" {
				ms, _ := strconv.Atoi(matches[3])
				// Normalize to milliseconds
				if len(matches[3]) == 2 {
					milliseconds = ms * 10 // Convert centiseconds to milliseconds
				} else {
					milliseconds = ms
				}
			}

			text := matches[4]

			// Calculate total time in milliseconds
			totalMs := int64(minutes*60*1000 + seconds*1000 + milliseconds)

			timestamps = append(timestamps, Timestamp{
				Time: totalMs,
				Text: text,
			})
		}
	}

	return timestamps, nil
}

// ValidateSongMetadata validates song metadata fields
func ValidateSongMetadata(title, artist string, duration int) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("song title cannot be empty")
	}

	if strings.TrimSpace(artist) == "" {
		return errors.New("song artist cannot be empty")
	}

	if duration <= 0 {
		return errors.New("song duration must be greater than 0")
	}

	// Reasonable duration limits (0-60 minutes in seconds)
	if duration > 3600 {
		return errors.New("song duration exceeds maximum allowed (60 minutes)")
	}

	return nil
}

// TimestampsToJSON converts timestamps array to JSON string for storage
func TimestampsToJSON(timestamps []Timestamp) (string, error) {
	data, err := json.Marshal(timestamps)
	if err != nil {
		return "", fmt.Errorf("failed to marshal timestamps: %w", err)
	}
	return string(data), nil
}

// JSONToTimestamps converts JSON string back to timestamps array
func JSONToTimestamps(jsonStr string) ([]Timestamp, error) {
	var timestamps []Timestamp
	if err := json.Unmarshal([]byte(jsonStr), &timestamps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal timestamps: %w", err)
	}
	return timestamps, nil
}

// ValidateFileHash validates that a file hash is in proper format (SHA256)
func ValidateFileHash(hash string) error {
	if hash == "" {
		return nil // Hash is optional
	}

	// SHA256 produces a 64-character hexadecimal string
	if len(hash) != 64 {
		return errors.New("file hash must be 64 characters (SHA256)")
	}

	matched, _ := regexp.MatchString("^[a-fA-F0-9]{64}$", hash)
	if !matched {
		return errors.New("file hash must contain only hexadecimal characters")
	}

	return nil
}
