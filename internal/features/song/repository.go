package song

import (
	"strings"

	"github.com/Tiavina22/lyrify-backend/internal/models"
	"github.com/Tiavina22/lyrify-backend/internal/utils"
	"gorm.io/gorm"
)

// Repository handles data access for song-related operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new song repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindByHash finds a song by its file hash
func (r *Repository) FindByHash(hash string) (*models.Song, error) {
	var song models.Song
	err := r.db.
		Where("file_hash = ?", hash).
		Preload("Artists").
		Preload("Lyrics", func(db *gorm.DB) *gorm.DB {
			return db.Order("version DESC, upvotes DESC").Limit(1)
		}).
		First(&song).Error

	if err != nil {
		return nil, err
	}
	return &song, nil
}

// FindByNormalizedMatch finds a song by normalized title, artist, and duration
func (r *Repository) FindByNormalizedMatch(title, artist string, duration int) (*models.Song, error) {
	normalizedTitle := utils.NormalizeString(title)
	normalizedArtist := utils.NormalizeString(artist)
	minDuration, maxDuration := utils.NormalizeDuration(duration, 2)

	var song models.Song
	err := r.db.
		Where("normalized_title = ? AND normalized_artist = ? AND duration BETWEEN ? AND ?",
			normalizedTitle, normalizedArtist, minDuration, maxDuration).
		Preload("Artists").
		Preload("Lyrics", func(db *gorm.DB) *gorm.DB {
			return db.Where("duration BETWEEN ? AND ?", minDuration, maxDuration).
				Order("version DESC, upvotes DESC").
				Limit(1)
		}).
		First(&song).Error

	if err != nil {
		return nil, err
	}
	return &song, nil
}

// FindByID retrieves a song by its ID with all relationships
func (r *Repository) FindByID(id string) (*models.Song, error) {
	var song models.Song
	err := r.db.
		Preload("Artists").
		Preload("Lyrics", func(db *gorm.DB) *gorm.DB {
			return db.Order("version DESC")
		}).
		Where("id = ?", id).
		First(&song).Error

	if err != nil {
		return nil, err
	}
	return &song, nil
}

// FindLyricsByID retrieves lyrics by ID
func (r *Repository) FindLyricsByID(id string) (*models.Lyrics, error) {
	var lyrics models.Lyrics
	err := r.db.
		Preload("Song").
		Where("id = ?", id).
		First(&lyrics).Error

	if err != nil {
		return nil, err
	}
	return &lyrics, nil
}

// Search searches for songs by title and/or artist
func (r *Repository) Search(title, artist string, limit int) ([]models.Song, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := r.db.Preload("Artists").Preload("Lyrics")

	if title != "" {
		normalizedTitle := utils.NormalizeString(title)
		query = query.Where("normalized_title LIKE ?", "%"+normalizedTitle+"%")
	}

	if artist != "" {
		normalizedArtist := utils.NormalizeString(artist)
		query = query.Where("normalized_artist LIKE ?", "%"+normalizedArtist+"%")
	}

	var songs []models.Song
	err := query.Limit(limit).Find(&songs).Error
	if err != nil {
		return nil, err
	}

	return songs, nil
}

// CreateRequestHistory logs a match request to history
func (r *Repository) CreateRequestHistory(history *models.RequestHistory) error {
	return r.db.Create(history).Error
}

// CreatePendingRequest creates or updates a pending request
func (r *Repository) CreatePendingRequest(title, artist string, duration int) error {
	// Try to find existing pending request
	var existing models.PendingRequest
	err := r.db.Where("title = ? AND artist = ? AND duration = ?", title, artist, duration).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new pending request
		pr := &models.PendingRequest{
			Title:    title,
			Artist:   artist,
			Duration: duration,
			Status:   models.StatusPending,
		}
		return r.db.Create(pr).Error
	}

	if err != nil {
		return err
	}

	// Update existing: increment request count
	existing.IncrementRequestCount()
	return r.db.Save(&existing).Error
}

// GetTopPendingRequests retrieves top pending requests by priority
func (r *Repository) GetTopPendingRequests(limit int) ([]models.PendingRequest, error) {
	var requests []models.PendingRequest
	err := r.db.
		Where("status = ?", models.StatusPending).
		Order("priority DESC, request_count DESC").
		Limit(limit).
		Find(&requests).Error

	if err != nil {
		return nil, err
	}
	return requests, nil
}

// CreateArtist creates a new artist
func (r *Repository) CreateArtist(artist *models.Artist) error {
	return r.db.Create(artist).Error
}

// CreateSong creates a new song with associated artists
func (r *Repository) CreateSong(song *models.Song, artistIDs []string) error {
	// Start transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Find artists first to build normalized name
		var artists []models.Artist
		if len(artistIDs) > 0 {
			if err := tx.Where("id IN ?", artistIDs).Find(&artists).Error; err != nil {
				return err
			}

			// Verify all artists were found
			if len(artists) != len(artistIDs) {
				return gorm.ErrRecordNotFound
			}

			// Build normalized artist name from all artists
			artistNames := make([]string, len(artists))
			for i, artist := range artists {
				artistNames[i] = artist.Name
			}
			song.NormalizedArtist = utils.NormalizeString(strings.Join(artistNames, " "))
		}

		// Create the song
		if err := tx.Create(song).Error; err != nil {
			return err
		}

		// Associate artists with song
		if len(artists) > 0 {
			if err := tx.Model(song).Association("Artists").Append(&artists); err != nil {
				return err
			}
		}

		return nil
	})
}

// CreateLyrics creates new lyrics for a song
func (r *Repository) CreateLyrics(lyrics *models.Lyrics) error {
	return r.db.Create(lyrics).Error
}

// GetArtistByID retrieves an artist by ID
func (r *Repository) GetArtistByID(id string) (*models.Artist, error) {
	var artist models.Artist
	err := r.db.
		Preload("Songs").
		Where("id = ?", id).
		First(&artist).Error
	if err != nil {
		return nil, err
	}
	return &artist, nil
}
