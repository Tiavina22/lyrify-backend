package utils

import (
	"testing"
)

func TestValidateLRCFormat(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "Valid LRC with milliseconds",
			content: `[ar:Kalash]
[ti:Moment gâché]
[00:12.50]Premier vers
[00:18.30]Deuxième vers`,
			wantErr: false,
		},
		{
			name: "Valid LRC without milliseconds",
			content: `[00:12]Premier vers
[00:18]Deuxième vers`,
			wantErr: false,
		},
		{
			name: "Valid LRC with metadata only",
			content: `[ar:Artist]
[ti:Title]
[al:Album]
[00:00.00]Start`,
			wantErr: false,
		},
		{
			name:    "Empty content",
			content: "",
			wantErr: true,
		},
		{
			name:    "No timestamps",
			content: "[ar:Artist]\n[ti:Title]",
			wantErr: true,
		},
		{
			name:    "Invalid format",
			content: "00:12.50 Premier vers",
			wantErr: true,
		},
		{
			name:    "Invalid timestamp - minutes too large",
			content: "[100:12.50]Text",
			wantErr: true,
		},
		{
			name:    "Invalid timestamp - seconds too large",
			content: "[01:60.50]Text",
			wantErr: true,
		},
		{
			name: "Mixed valid and empty lines",
			content: `[00:12.50]Premier vers

[00:18.30]Deuxième vers`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLRCFormat(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLRCFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractLRCTimestamps(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedCount int
		wantErr       bool
	}{
		{
			name: "Simple LRC",
			content: `[00:12.50]First line
[00:18.30]Second line`,
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "LRC with metadata",
			content: `[ar:Artist]
[ti:Title]
[00:00.00]Start
[00:05.00]Next`,
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "LRC without milliseconds",
			content: `[00:12]First line
[00:18]Second line`,
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:          "Invalid LRC",
			content:       "Invalid content",
			expectedCount: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamps, err := ExtractLRCTimestamps(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractLRCTimestamps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(timestamps) != tt.expectedCount {
				t.Errorf("ExtractLRCTimestamps() got %d timestamps, want %d", len(timestamps), tt.expectedCount)
			}
		})
	}
}

func TestExtractLRCTimestampsContent(t *testing.T) {
	content := `[ar:Kalash]
[ti:Moment gâché]
[00:12.50]Premier vers de la chanson
[00:18.30]Deuxième vers`

	timestamps, err := ExtractLRCTimestamps(content)
	if err != nil {
		t.Fatalf("ExtractLRCTimestamps() error = %v", err)
	}

	if len(timestamps) != 2 {
		t.Fatalf("Expected 2 timestamps, got %d", len(timestamps))
	}

	// Check first timestamp
	expectedTime1 := int64(12*1000 + 500) // 12.50 seconds in milliseconds
	if timestamps[0].Time != expectedTime1 {
		t.Errorf("First timestamp time = %d, want %d", timestamps[0].Time, expectedTime1)
	}
	if timestamps[0].Text != "Premier vers de la chanson" {
		t.Errorf("First timestamp text = %q, want %q", timestamps[0].Text, "Premier vers de la chanson")
	}

	// Check second timestamp
	expectedTime2 := int64(18*1000 + 300) // 18.30 seconds in milliseconds
	if timestamps[1].Time != expectedTime2 {
		t.Errorf("Second timestamp time = %d, want %d", timestamps[1].Time, expectedTime2)
	}
	if timestamps[1].Text != "Deuxième vers" {
		t.Errorf("Second timestamp text = %q, want %q", timestamps[1].Text, "Deuxième vers")
	}
}

func TestValidateSongMetadata(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		artist   string
		duration int
		wantErr  bool
	}{
		{
			name:     "Valid metadata",
			title:    "Moment gâché",
			artist:   "Kalash",
			duration: 345,
			wantErr:  false,
		},
		{
			name:     "Empty title",
			title:    "",
			artist:   "Kalash",
			duration: 345,
			wantErr:  true,
		},
		{
			name:     "Empty artist",
			title:    "Moment gâché",
			artist:   "",
			duration: 345,
			wantErr:  true,
		},
		{
			name:     "Zero duration",
			title:    "Moment gâché",
			artist:   "Kalash",
			duration: 0,
			wantErr:  true,
		},
		{
			name:     "Negative duration",
			title:    "Moment gâché",
			artist:   "Kalash",
			duration: -10,
			wantErr:  true,
		},
		{
			name:     "Duration too long",
			title:    "Moment gâché",
			artist:   "Kalash",
			duration: 4000,
			wantErr:  true,
		},
		{
			name:     "Whitespace only title",
			title:    "   ",
			artist:   "Kalash",
			duration: 345,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSongMetadata(tt.title, tt.artist, tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSongMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFileHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "Valid SHA256 hash",
			hash:    "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
			wantErr: false,
		},
		{
			name:    "Empty hash (optional)",
			hash:    "",
			wantErr: false,
		},
		{
			name:    "Invalid length",
			hash:    "abc123",
			wantErr: true,
		},
		{
			name:    "Invalid characters",
			hash:    "g665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
			wantErr: true,
		},
		{
			name:    "Too long",
			hash:    "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3a",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileHash(tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimestampsToJSON(t *testing.T) {
	timestamps := []Timestamp{
		{Time: 12500, Text: "First line"},
		{Time: 18300, Text: "Second line"},
	}

	jsonStr, err := TimestampsToJSON(timestamps)
	if err != nil {
		t.Fatalf("TimestampsToJSON() error = %v", err)
	}

	// Convert back to verify
	result, err := JSONToTimestamps(jsonStr)
	if err != nil {
		t.Fatalf("JSONToTimestamps() error = %v", err)
	}

	if len(result) != len(timestamps) {
		t.Errorf("Got %d timestamps, want %d", len(result), len(timestamps))
	}

	for i := range result {
		if result[i].Time != timestamps[i].Time || result[i].Text != timestamps[i].Text {
			t.Errorf("Timestamp[%d] = {%d, %q}, want {%d, %q}",
				i, result[i].Time, result[i].Text, timestamps[i].Time, timestamps[i].Text)
		}
	}
}

func TestCleanLRCContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Content with HTML garbage from website",
			input: `BACK TO HOME

Search for keyword: damso

Débrouillard
4:03
Synced
Batterie faible - Damso

Preview
SYNCED LYRICS
PLAIN LYRICS

[00:16.46] Demandez pas c'que je fais dans la vie
[00:18.74] J'suis fonce-dé avec 2-3 youvois du quartier
[00:20.86] J'fume un pilon sur l'toit de la ville`,
			expected: `[00:16.46] Demandez pas c'que je fais dans la vie
[00:18.74] J'suis fonce-dé avec 2-3 youvois du quartier
[00:20.86] J'fume un pilon sur l'toit de la ville`,
		},
		{
			name: "Content with metadata tags",
			input: `Some random text
[ar:Damso]
[ti:Débrouillard]
[al:Batterie faible]
More random text
[00:16.46] Lyrics start here
Navigation menu`,
			expected: `[ar:Damso]
[ti:Débrouillard]
[al:Batterie faible]
[00:16.46] Lyrics start here`,
		},
		{
			name: "Already clean content",
			input: `[ar:Artist]
[ti:Title]
[00:12.50] Line one
[00:18.30] Line two`,
			expected: `[ar:Artist]
[ti:Title]
[00:12.50] Line one
[00:18.30] Line two`,
		},
		{
			name:     "Empty content",
			input:    "",
			expected: "",
		},
		{
			name:     "Only garbage, no LRC",
			input:    "Just some random text\nWith multiple lines\nBut no LRC content",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanLRCContent(tt.input)
			if result != tt.expected {
				t.Errorf("CleanLRCContent() =\n%q\n\nwant:\n%q", result, tt.expected)
			}
		})
	}
}
