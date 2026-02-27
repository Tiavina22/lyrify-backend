package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RequestHistory tracks all song/lyrics search requests made by users
// This table is used for analytics and identifying popular songs that need lyrics
type RequestHistory struct {
	ID             string    `gorm:"type:uuid;primaryKey" json:"id"`
	DeviceID       string    `gorm:"type:varchar(255);index:idx_request_device" json:"device_id,omitempty"` // Optional device/user identifier
	UserID         string    `gorm:"type:varchar(255);index:idx_request_user" json:"user_id,omitempty"`     // Optional user identifier (for future auth)
	SearchTitle    string    `gorm:"type:varchar(500);not null" json:"search_title"`                        // Title searched by user
	SearchArtist   string    `gorm:"type:varchar(500);not null" json:"search_artist"`                       // Artist searched by user
	SearchDuration int       `gorm:"not null" json:"search_duration"`                                       // Duration searched (in seconds)
	StatusCode     int       `gorm:"not null;index:idx_request_status" json:"status_code"`                  // HTTP status code (200, 404, etc.)
	Found          bool      `gorm:"not null;default:false" json:"found"`                                   // Whether lyrics were found
	SongID         *string   `gorm:"type:uuid;index:idx_request_song" json:"song_id,omitempty"`             // Song ID if found (nullable)
	IPAddress      string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`                          // IPv4 or IPv6
	UserAgent      string    `gorm:"type:text" json:"user_agent,omitempty"`                                 // Browser/client user agent
	CreatedAt      time.Time `gorm:"autoCreateTime;index:idx_request_created" json:"created_at"`

	// Relationship (optional)
	Song *Song `gorm:"foreignKey:SongID" json:"song,omitempty"`
}

// BeforeCreate hook to generate UUID before creating
func (r *RequestHistory) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// TableName specifies the table name for RequestHistory model
func (RequestHistory) TableName() string {
	return "request_history"
}
