package models

import (
	"time"

	"github.com/Tiavina22/lyrify-backend/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Artist represents a music artist
type Artist struct {
	ID             string    `gorm:"type:uuid;primaryKey" json:"id"`
	Name           string    `gorm:"type:varchar(255);not null" json:"name"`
	NormalizedName string    `gorm:"type:varchar(255);not null;index:idx_artist_normalized_name" json:"normalized_name"`
	Country        string    `gorm:"type:varchar(100)" json:"country,omitempty"`
	Biography      string    `gorm:"type:text" json:"biography,omitempty"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Songs []Song `gorm:"many2many:song_artists;" json:"songs,omitempty"`
}

// BeforeCreate hook to generate UUID and normalize name before creating
func (a *Artist) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	// Normalize the name for searching
	a.NormalizedName = utils.NormalizeString(a.Name)

	return nil
}

// BeforeSave hook to normalize name before saving
func (a *Artist) BeforeSave(tx *gorm.DB) error {
	// Always update normalized name when name changes
	a.NormalizedName = utils.NormalizeString(a.Name)
	return nil
}

// TableName specifies the table name for Artist model
func (Artist) TableName() string {
	return "artists"
}
