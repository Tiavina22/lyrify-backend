package song

import (
	"log"
	"time"

	"github.com/Tiavina22/lyrify-backend/internal/errors"
	"github.com/Tiavina22/lyrify-backend/internal/models"
	"github.com/Tiavina22/lyrify-backend/internal/utils"
	"gorm.io/gorm"
)

// Service handles business logic for song operations
type Service struct {
	repo *Repository
}

// NewService creates a new song service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Match finds the best matching song and lyrics
// Returns MatchResponse with found=true if match exists, or found=false if not found
func (s *Service) Match(req MatchRequest, deviceID string, ipAddress string) (*MatchResponse, error) {
	// 1. Validate request
	if err := utils.ValidateSongMetadata(req.Title, req.Artist, req.Duration); err != nil {
		return nil, err
	}

	var song *models.Song
	var err error

	// 2. Try hash match first (if hash provided)
	if req.Hash != "" {
		if err := utils.ValidateFileHash(req.Hash); err == nil {
			song, err = s.repo.FindByHash(req.Hash)
			if err == nil && song != nil {
				log.Printf("✓ Match found by hash: %s", req.Hash)
				s.logRequestHistory(req, deviceID, ipAddress, song.ID, 200, true)
				return s.buildResponse(song), nil
			}
		}
	}

	// 3. Try normalized match
	song, err = s.repo.FindByNormalizedMatch(req.Title, req.Artist, req.Duration)
	if err == nil && song != nil {
		log.Printf("✓ Match found by normalized match: %s - %s", req.Title, req.Artist)
		s.logRequestHistory(req, deviceID, ipAddress, song.ID, 200, true)
		return s.buildResponse(song), nil
	}

	// 4. Not found - log and create pending request
	log.Printf("✗ No match found for: %s - %s (%d sec)", req.Title, req.Artist, req.Duration)
	s.logRequestHistory(req, deviceID, ipAddress, "", 404, false)

	// Create or update pending request
	if err := s.repo.CreatePendingRequest(req.Title, req.Artist, req.Duration); err != nil {
		log.Printf("Warning: Failed to create pending request: %v", err)
	}

	return &MatchResponse{Found: false}, errors.ErrSongNotFound
}

// GetSongByID retrieves a song by its ID
func (s *Service) GetSongByID(id string) (*models.Song, error) {
	song, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrSongNotFound
		}
		return nil, err
	}
	return song, nil
}

// GetLyricsByID retrieves lyrics by its ID
func (s *Service) GetLyricsByID(id string) (*models.Lyrics, error) {
	lyrics, err := s.repo.FindLyricsByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrLyricsNotFound
		}
		return nil, err
	}
	return lyrics, nil
}

// Search searches for songs by title and/or artist
func (s *Service) Search(req SearchRequest) (*SearchResponse, error) {
	songs, err := s.repo.Search(req.Title, req.Artist, req.Limit)
	if err != nil {
		return nil, err
	}

	return &SearchResponse{
		Songs: songs,
		Total: len(songs),
	}, nil
}

// GetTopPendingRequests retrieves top pending requests by priority
func (s *Service) GetTopPendingRequests(limit int) (*PendingRequestsResponse, error) {
	if limit <= 0 {
		limit = 50
	}

	requests, err := s.repo.GetTopPendingRequests(limit)
	if err != nil {
		return nil, err
	}

	return &PendingRequestsResponse{
		Requests: requests,
		Total:    len(requests),
	}, nil
}

// buildResponse builds a match response from song and lyrics
func (s *Service) buildResponse(song *models.Song) *MatchResponse {
	response := &MatchResponse{
		Found:  true,
		SongID: song.ID,
		Song:   song,
	}

	// Get best lyrics (highest version or upvotes)
	if len(song.Lyrics) > 0 {
		lyrics := &song.Lyrics[0]
		response.LyricsID = lyrics.ID
		response.Version = lyrics.Version
		response.Offset = lyrics.Offset
		response.LRC = lyrics.LRCContent
		response.Lyrics = lyrics
	}

	return response
}

// logRequestHistory logs a request to the history table for analytics
func (s *Service) logRequestHistory(req MatchRequest, deviceID, ipAddress, songID string, statusCode int, found bool) {
	history := &models.RequestHistory{
		DeviceID:       deviceID,
		SearchTitle:    req.Title,
		SearchArtist:   req.Artist,
		SearchDuration: req.Duration,
		StatusCode:     statusCode,
		Found:          found,
		IPAddress:      ipAddress,
		CreatedAt:      time.Now(),
	}

	if songID != "" {
		history.SongID = &songID
	}

	if err := s.repo.CreateRequestHistory(history); err != nil {
		log.Printf("Warning: Failed to log request history: %v", err)
	}
}

// CreateArtist creates a new artist
func (s *Service) CreateArtist(req CreateArtistRequest) (*models.Artist, error) {
	// Validate artist name
	if req.Name == "" {
		return nil, errors.BadRequest("Artist name is required", nil)
	}

	artist := &models.Artist{
		Name: req.Name,
		// NormalizedName is auto-populated by GORM hooks
	}

	if err := s.repo.CreateArtist(artist); err != nil {
		return nil, errors.InternalServerError("Failed to create artist", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return artist, nil
}

// CreateSong creates a new song with associated artists
func (s *Service) CreateSong(req CreateSongRequest) (*models.Song, error) {
	// Validate request
	if err := utils.ValidateSongMetadata(req.Title, req.ArtistIDs[0], req.Duration); err != nil {
		return nil, err
	}

	// Validate file hash only if provided
	if req.FileHash != "" {
		if err := utils.ValidateFileHash(req.FileHash); err != nil {
			return nil, err
		}
	}

	if len(req.ArtistIDs) == 0 {
		return nil, errors.BadRequest("At least one artist is required", nil)
	}

	// Create song
	song := &models.Song{
		Title:    req.Title,
		Duration: req.Duration,
		FileHash: req.FileHash,
		Album:    req.Album,
		Year:     req.ReleaseYear,
		// NormalizedTitle and NormalizedArtist will be set by GORM hooks
	}

	if err := s.repo.CreateSong(song, req.ArtistIDs); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("One or more artists not found")
		}
		return nil, errors.InternalServerError("Failed to create song", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Reload song with relationships
	createdSong, err := s.repo.FindByID(song.ID)
	if err != nil {
		return nil, err
	}

	return createdSong, nil
}

// CreateLyrics creates new lyrics for a song
func (s *Service) CreateLyrics(req CreateLyricsRequest) (*models.Lyrics, error) {
	// Clean LRC content first (removes HTML/navigation text when copy-pasted from websites)
	cleanedContent := utils.CleanLRCContent(req.LRCContent)
	if cleanedContent == "" {
		return nil, errors.BadRequest("No valid LRC content found", map[string]interface{}{
			"hint": "LRC content must contain lines with timestamps like [00:12.34] Lyric text",
		})
	}

	// Validate LRC format
	if err := utils.ValidateLRCFormat(cleanedContent); err != nil {
		return nil, err
	}

	// Verify song exists
	song, err := s.repo.FindByID(req.SongID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("Song not found")
		}
		return nil, err
	}

	// Extract timestamps
	timestamps, err := utils.ExtractLRCTimestamps(cleanedContent)
	if err != nil {
		return nil, errors.BadRequest("Failed to parse LRC timestamps", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Convert timestamps to JSON string
	timestampsJSON, err := utils.TimestampsToJSON(timestamps)
	if err != nil {
		return nil, errors.InternalServerError("Failed to encode timestamps", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Set default version if not provided
	version := req.Version
	if version == 0 {
		version = 1
	}

	lyrics := &models.Lyrics{
		SongID:        req.SongID,
		LRCContent:    cleanedContent, // Use cleaned content
		LRCTimestamps: timestampsJSON,
		Duration:      song.Duration, // Use song's duration
		Version:       version,
		Offset:        req.Offset,
		Language:      req.Language,
	}

	if err := s.repo.CreateLyrics(lyrics); err != nil {
		return nil, errors.InternalServerError("Failed to create lyrics", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return lyrics, nil
}
