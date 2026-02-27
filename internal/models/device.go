package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DevicePlatform represents the device platform (for push notifications)
type DevicePlatform string

const (
	PlatformIOS     DevicePlatform = "ios"
	PlatformAndroid DevicePlatform = "android"
	PlatformMacOS   DevicePlatform = "macos"
	PlatformWindows DevicePlatform = "windows"
	PlatformLinux   DevicePlatform = "linux"
	PlatformWeb     DevicePlatform = "web"
)

// Device represents a user device for push notifications
type Device struct {
	ID                   string         `gorm:"type:uuid;primaryKey" json:"id"`
	Token                string         `gorm:"type:varchar(500);not null;uniqueIndex:idx_device_token" json:"token"` // FCM/APNs token
	Platform             DevicePlatform `gorm:"type:varchar(20);not null" json:"platform"`
	DeviceName           string         `gorm:"type:varchar(255)" json:"device_name,omitempty"`           // e.g., "iPhone 14", "MacBook Pro"
	AppVersion           string         `gorm:"type:varchar(50)" json:"app_version,omitempty"`            // e.g., "1.0.0"
	OSVersion            string         `gorm:"type:varchar(50)" json:"os_version,omitempty"`             // e.g., "iOS 17.2"
	Language             string         `gorm:"type:varchar(10);default:'en'" json:"language"`            // User preferred language
	Timezone             string         `gorm:"type:varchar(50)" json:"timezone,omitempty"`               // e.g., "Europe/Paris"
	UserID               *string        `gorm:"type:uuid;index:idx_device_user" json:"user_id,omitempty"` // Optional: linked user (future auth)
	Active               bool           `gorm:"default:true" json:"active"`                               // Device is active
	NotificationsEnabled bool           `gorm:"default:true" json:"notifications_enabled"`                // User enabled notifications
	LastSeenAt           time.Time      `gorm:"autoUpdateTime" json:"last_seen_at"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Notification preferences (expand as needed)
	NotifyOnSongAvailable  bool `gorm:"default:true" json:"notify_on_song_available"`
	NotifyOnPendingRequest bool `gorm:"default:true" json:"notify_on_pending_request"`
	NotifyOnLyricsUpdated  bool `gorm:"default:false" json:"notify_on_lyrics_updated"`
	NotifyOnPopularSong    bool `gorm:"default:false" json:"notify_on_popular_song"`
}

// BeforeCreate hook to generate UUID before creating
func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// IsActive returns true if device is active and notifications are enabled
func (d *Device) IsActive() bool {
	return d.Active && d.NotificationsEnabled
}

// TableName specifies the table name for Device model
func (Device) TableName() string {
	return "devices"
}

// NotificationLog represents a log of sent notifications
// Useful for tracking delivery status and debugging
type NotificationLog struct {
	ID               string    `gorm:"type:uuid;primaryKey" json:"id"`
	DeviceID         string    `gorm:"type:uuid;not null;index:idx_notif_device" json:"device_id"`
	NotificationType string    `gorm:"type:varchar(50);not null" json:"notification_type"` // e.g., "song_available", "pending_request"
	Title            string    `gorm:"type:varchar(255);not null" json:"title"`
	Body             string    `gorm:"type:text;not null" json:"body"`
	DataPayload      string    `gorm:"type:jsonb" json:"data_payload,omitempty"` // Additional data (song_id, etc.)
	Status           string    `gorm:"type:varchar(20);not null" json:"status"`  // "sent", "failed", "delivered"
	ErrorMessage     string    `gorm:"type:text" json:"error_message,omitempty"`
	SentAt           time.Time `gorm:"autoCreateTime" json:"sent_at"`

	// Relationship
	Device Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
}

// BeforeCreate hook to generate UUID before creating
func (n *NotificationLog) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}

// TableName specifies the table name for NotificationLog model
func (NotificationLog) TableName() string {
	return "notification_logs"
}
