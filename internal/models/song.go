package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Song struct {
	ID         string `gorm:"type:uuid;primaryKey"`
	Title      string `gorm:"not null"`
	Artist     string `gorm:"not null"`
	Lyrics     string `gorm:"type:text"`
	Timestamps string `gorm:"type:text"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (s *Song) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New().String()
	return
}
