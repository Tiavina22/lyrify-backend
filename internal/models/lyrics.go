package models

import (
	"time"

	"github.com/Tiavina22/lyrify-backend/internal/errors"
	"github.com/Tiavina22/lyrify-backend/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Lyrics represents synchronized lyrics for a song
type Lyrics struct {
	ID            string    `gorm:"type:uuid;primaryKey" json:"id"`
	SongID        string    `gorm:"type:uuid;not null;index:idx_lyrics_song_version,priority:1" json:"song_id"`
	LRCContent    string    `gorm:"type:text;not null" json:"lrc_content"`
	LRCTimestamps string    `gorm:"type:jsonb" json:"lrc_timestamps"` // JSON array of timestamps
	Duration      int       `gorm:"not null" json:"duration"`         // Duration in seconds, must match song duration
	Version       int       `gorm:"not null;default:1;index:idx_lyrics_song_version,priority:2" json:"version"`
	Upvotes       int       `gorm:"default:0" json:"upvotes"`
	Language      string    `gorm:"type:varchar(10)" json:"language,omitempty"` // e.g., "en", "fr", "es"
	Offset        int       `gorm:"default:0" json:"offset"`                    // Timing offset in milliseconds
	Source        string    `gorm:"type:varchar(100)" json:"source,omitempty"`  // e.g., "community", "official", "auto"
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationship
	Song Song `gorm:"foreignKey:SongID;constraint:OnDelete:CASCADE" json:"song,omitempty"`
}

// BeforeCreate hook to generate UUID and validate before creating
func (l *Lyrics) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}

	// Validate LRC format
	if err := utils.ValidateLRCFormat(l.LRCContent); err != nil {
		return errors.ErrInvalidLRCFormat
	}

	// Extract and store timestamps as JSON
	timestamps, err := utils.ExtractLRCTimestamps(l.LRCContent)
	if err != nil {
		return err
	}

	timestampsJSON, err := utils.TimestampsToJSON(timestamps)
	if err != nil {
		return err
	}
	l.LRCTimestamps = timestampsJSON

	// Validate duration matches song (will be validated in service layer)
	if l.Duration <= 0 {
		return errors.ErrInvalidDuration
	}

	return nil
}

// BeforeSave hook to validate before saving
func (l *Lyrics) BeforeSave(tx *gorm.DB) error {
	// Validate LRC format
	if err := utils.ValidateLRCFormat(l.LRCContent); err != nil {
		return errors.ErrInvalidLRCFormat
	}

	// Update timestamps if content changed
	timestamps, err := utils.ExtractLRCTimestamps(l.LRCContent)
	if err != nil {
		return err
	}

	timestampsJSON, err := utils.TimestampsToJSON(timestamps)
	if err != nil {
		return err
	}
	l.LRCTimestamps = timestampsJSON

	return nil
}

// ValidateDurationMatch checks if lyrics duration matches song duration within tolerance
func (l *Lyrics) ValidateDurationMatch(songDuration int, tolerance int) error {
	if !utils.CompareDurations(l.Duration, songDuration, tolerance) {
		return errors.ErrDurationMismatch
	}
	return nil
}

// TableName specifies the table name for Lyrics model
func (Lyrics) TableName() string {
	return "lyrics"
}
