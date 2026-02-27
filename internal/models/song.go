package models

import (
	"time"

	"github.com/Tiavina22/lyrify-backend/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Song represents a music song/track
type Song struct {
	ID               string    `gorm:"type:uuid;primaryKey" json:"id"`
	Title            string    `gorm:"type:varchar(500);not null" json:"title"`
	NormalizedTitle  string    `gorm:"type:varchar(500);not null;index:idx_song_search,priority:1" json:"normalized_title"`
	NormalizedArtist string    `gorm:"type:varchar(500);not null;index:idx_song_search,priority:2" json:"normalized_artist"`
	Duration         int       `gorm:"not null;index:idx_song_search,priority:3" json:"duration"` // Duration in seconds
	Album            string    `gorm:"type:varchar(500)" json:"album,omitempty"`
	Year             int       `gorm:"type:int" json:"year,omitempty"`
	FileHash         string    `gorm:"type:varchar(64);index:idx_song_hash" json:"file_hash,omitempty"` // SHA256 hash
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Artists []Artist `gorm:"many2many:song_artists;" json:"artists,omitempty"`
	Lyrics  []Lyrics `gorm:"foreignKey:SongID;constraint:OnDelete:CASCADE" json:"lyrics,omitempty"`
}

// BeforeCreate hook to generate UUID and normalize fields before creating
func (s *Song) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	// Normalize title for searching
	s.NormalizedTitle = utils.NormalizeString(s.Title)

	// NormalizedArtist will be set by the repository after associating artists
	// Don't set it here to avoid normalizing itself

	return nil
}

// BeforeSave hook to normalize fields before saving
func (s *Song) BeforeSave(tx *gorm.DB) error {
	// Always update normalized title when title changes
	s.NormalizedTitle = utils.NormalizeString(s.Title)

	// NormalizedArtist should be updated by the repository when artists change
	// Don't normalize it here to avoid recursive normalization

	return nil
}

// TableName specifies the table name for Song model
func (Song) TableName() string {
	return "songs"
}
