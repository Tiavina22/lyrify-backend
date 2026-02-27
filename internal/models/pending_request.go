package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RequestStatus represents the status of a pending lyrics request
type RequestStatus string

const (
	StatusPending    RequestStatus = "pending"
	StatusInProgress RequestStatus = "in_progress"
	StatusCompleted  RequestStatus = "completed"
	StatusRejected   RequestStatus = "rejected"
)

// PendingRequest represents a song that was requested but has no lyrics yet
// This table helps prioritize which songs need lyrics creation
type PendingRequest struct {
	ID               string         `gorm:"type:uuid;primaryKey" json:"id"`
	Title            string         `gorm:"type:varchar(500);not null;uniqueIndex:idx_pending_unique,priority:1" json:"title"`
	Artist           string         `gorm:"type:varchar(500);not null;uniqueIndex:idx_pending_unique,priority:2" json:"artist"`
	Album            string         `gorm:"type:varchar(500)" json:"album,omitempty"`
	Duration         int            `gorm:"not null;uniqueIndex:idx_pending_unique,priority:3" json:"duration"` // Duration in seconds
	FileHash         string         `gorm:"type:varchar(64)" json:"file_hash,omitempty"`                        // SHA256 hash (optional)
	RequestCount     int            `gorm:"default:1;index:idx_pending_count" json:"request_count"`             // How many times this was requested
	Priority         int            `gorm:"default:0;index:idx_pending_priority" json:"priority"`               // Calculated priority (higher = more urgent)
	Status           RequestStatus  `gorm:"type:varchar(20);default:'pending';index:idx_pending_status" json:"status"`
	Metadata         datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional metadata as JSON
	FirstRequestedAt time.Time      `gorm:"autoCreateTime" json:"first_requested_at"`
	LastRequestedAt  time.Time      `gorm:"autoCreateTime;autoUpdateTime" json:"last_requested_at"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate hook to generate UUID and calculate initial priority
func (p *PendingRequest) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}

	// Calculate initial priority based on request count
	p.Priority = p.RequestCount

	return nil
}

// BeforeUpdate hook to recalculate priority when request count changes
func (p *PendingRequest) BeforeUpdate(tx *gorm.DB) error {
	// Recalculate priority (can be enhanced with more complex logic)
	// Higher request count = higher priority
	p.Priority = p.RequestCount

	return nil
}

// IncrementRequestCount increments the request count for this pending request
func (p *PendingRequest) IncrementRequestCount() {
	p.RequestCount++
	p.LastRequestedAt = time.Now()
}

// MarkInProgress changes the status to in_progress
func (p *PendingRequest) MarkInProgress() {
	p.Status = StatusInProgress
}

// MarkCompleted changes the status to completed
func (p *PendingRequest) MarkCompleted() {
	p.Status = StatusCompleted
}

// MarkRejected changes the status to rejected
func (p *PendingRequest) MarkRejected() {
	p.Status = StatusRejected
}

// TableName specifies the table name for PendingRequest model
func (PendingRequest) TableName() string {
	return "pending_requests"
}
