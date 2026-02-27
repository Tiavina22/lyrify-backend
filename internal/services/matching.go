package services

import (
	"log"

	"github.com/Tiavina22/lyrify-backend/internal/errors"
	"github.com/Tiavina22/lyrify-backend/internal/models"
	"github.com/Tiavina22/lyrify-backend/internal/utils"

	"gorm.io/gorm"
)

// MatchService handles song and lyrics matching logic
type MatchService struct {
	db *gorm.DB
}

// NewMatchService creates a new match service instance
func NewMatchService(db *gorm.DB) *MatchService {
	return &MatchService{db: db}
}

// MatchRequest represents a request to find matching song/lyrics
type MatchRequest struct {
	Title    string `json:"title" binding:"required"`
	Artist   string `json:"artist" binding:"required"`
	Duration int    `json:"duration" binding:"required"` // Duration in seconds
	Hash     string `json:"hash"`                        // Optional SHA256 file hash
}

// MatchResponse represents the response for a match request
type MatchResponse struct {
	Found    bool           `json:"found"`
	LyricsID string         `json:"lyrics_id,omitempty"`
	SongID   string         `json:"song_id,omitempty"`
	Version  int            `json:"version,omitempty"`
	Offset   int            `json:"offset,omitempty"`
	LRC      string         `json:"lrc,omitempty"`
	Song     *models.Song   `json:"song,omitempty"`
	Lyrics   *models.Lyrics `json:"lyrics,omitempty"`
}

// FindBestMatch attempts to find the best matching song and lyrics
// Matching strategy:
// 1. Try exact hash match (fastest, most accurate)
// 2. Try normalized title + artist + duration match (±2 seconds tolerance)
// 3. Try fuzzy match on title + artist with duration tolerance
func (s *MatchService) FindBestMatch(req *MatchRequest) (*MatchResponse, error) {
	// Validate request
	if err := utils.ValidateSongMetadata(req.Title, req.Artist, req.Duration); err != nil {
		return nil, err
	}

	var song *models.Song
	var lyrics *models.Lyrics
	var err error

	// Strategy 1: Try exact hash match (if hash provided)
	if req.Hash != "" {
		if err := utils.ValidateFileHash(req.Hash); err == nil {
			song, lyrics, err = s.findByHash(req.Hash)
			if err == nil && song != nil {
				log.Printf("✓ Match found by hash: %s", req.Hash)
				return s.buildResponse(song, lyrics), nil
			}
		}
	}

	// Strategy 2: Try normalized match with duration tolerance
	song, lyrics, err = s.findByNormalizedMatch(req.Title, req.Artist, req.Duration)
	if err == nil && song != nil {
		log.Printf("✓ Match found by normalized match: %s - %s", req.Title, req.Artist)
		return s.buildResponse(song, lyrics), nil
	}

	// Strategy 3: No match found
	log.Printf("✗ No match found for: %s - %s (%d sec)", req.Title, req.Artist, req.Duration)
	return &MatchResponse{Found: false}, errors.ErrSongNotFound
}

// findByHash finds a song by its file hash
func (s *MatchService) findByHash(hash string) (*models.Song, *models.Lyrics, error) {
	var song models.Song

	err := s.db.Where("file_hash = ?", hash).
		Preload("Lyrics", func(db *gorm.DB) *gorm.DB {
			// Get the lyrics with highest version or most upvotes
			return db.Order("version DESC, upvotes DESC").Limit(1)
		}).
		First(&song).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errors.ErrSongNotFound
		}
		return nil, nil, err
	}

	// Get the best lyrics for this song
	var lyrics *models.Lyrics
	if len(song.Lyrics) > 0 {
		lyrics = &song.Lyrics[0]
	}

	return &song, lyrics, nil
}

// findByNormalizedMatch finds a song by normalized title, artist, and duration
func (s *MatchService) findByNormalizedMatch(title, artist string, duration int) (*models.Song, *models.Lyrics, error) {
	normalizedTitle := utils.NormalizeString(title)
	normalizedArtist := utils.NormalizeString(artist)

	// Duration tolerance: ±2 seconds
	minDuration, maxDuration := utils.NormalizeDuration(duration, 2)

	var song models.Song

	err := s.db.Where("normalized_title = ? AND normalized_artist = ? AND duration BETWEEN ? AND ?",
		normalizedTitle, normalizedArtist, minDuration, maxDuration).
		Preload("Lyrics", func(db *gorm.DB) *gorm.DB {
			// Get lyrics matching the duration with tolerance
			return db.Where("duration BETWEEN ? AND ?", minDuration, maxDuration).
				Order("version DESC, upvotes DESC").
				Limit(1)
		}).
		First(&song).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errors.ErrSongNotFound
		}
		return nil, nil, err
	}

	// Get the best lyrics for this song
	var lyrics *models.Lyrics
	if len(song.Lyrics) > 0 {
		lyrics = &song.Lyrics[0]
	}

	return &song, lyrics, nil
}

// findByFuzzyMatch finds songs using fuzzy matching (for future implementation)
// This would use Levenshtein distance or similar algorithm
func (s *MatchService) findByFuzzyMatch(title, artist string, duration int) (*models.Song, *models.Lyrics, error) {
	// TODO: Implement fuzzy matching using a library like "github.com/agnivade/levenshtein"
	// For now, return not found
	return nil, nil, errors.ErrSongNotFound
}

// buildResponse builds a match response from song and lyrics
func (s *MatchService) buildResponse(song *models.Song, lyrics *models.Lyrics) *MatchResponse {
	response := &MatchResponse{
		Found:  true,
		SongID: song.ID,
		Song:   song,
	}

	if lyrics != nil {
		response.LyricsID = lyrics.ID
		response.Version = lyrics.Version
		response.Offset = lyrics.Offset
		response.LRC = lyrics.LRCContent
		response.Lyrics = lyrics
	}

	return response
}

// SearchSongs searches for songs by title and/or artist (for listing/browsing)
func (s *MatchService) SearchSongs(title, artist string, limit int) ([]models.Song, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := s.db.Preload("Artists").Preload("Lyrics")

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

// GetSongByID retrieves a song by its ID with all relationships
func (s *MatchService) GetSongByID(id string) (*models.Song, error) {
	var song models.Song

	err := s.db.Preload("Artists").
		Preload("Lyrics", func(db *gorm.DB) *gorm.DB {
			return db.Order("version DESC")
		}).
		Where("id = ?", id).
		First(&song).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrSongNotFound
		}
		return nil, err
	}

	return &song, nil
}

// GetLyricsByID retrieves lyrics by its ID
func (s *MatchService) GetLyricsByID(id string) (*models.Lyrics, error) {
	var lyrics models.Lyrics

	err := s.db.Preload("Song").
		Where("id = ?", id).
		First(&lyrics).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrLyricsNotFound
		}
		return nil, err
	}

	return &lyrics, nil
}
