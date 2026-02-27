package song

import "github.com/Tiavina22/lyrify-backend/internal/models"

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

// SearchRequest represents a request to search for songs
type SearchRequest struct {
	Title  string `json:"title" form:"title"`
	Artist string `json:"artist" form:"artist"`
	Limit  int    `json:"limit" form:"limit"`
}

// SearchResponse represents the response for a search request
type SearchResponse struct {
	Songs []models.Song `json:"songs"`
	Total int           `json:"total"`
}

// PendingRequestsResponse represents the response for pending requests
type PendingRequestsResponse struct {
	Requests []models.PendingRequest `json:"requests"`
	Total    int                     `json:"total"`
}

// CreateArtistRequest represents a request to create a new artist
type CreateArtistRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateSongRequest represents a request to create a new song
type CreateSongRequest struct {
	Title       string   `json:"title" binding:"required"`
	ArtistIDs   []string `json:"artist_ids" binding:"required"` // UUIDs of existing artists
	Duration    int      `json:"duration" binding:"required"`   // Duration in seconds
	FileHash    string   `json:"file_hash"`                     // SHA256 hash (optional, but recommended for exact matching)
	Album       string   `json:"album"`
	ReleaseYear int      `json:"release_year"`
}

// CreateLyricsRequest represents a request to create lyrics for a song
type CreateLyricsRequest struct {
	SongID     string `json:"song_id" binding:"required"`     // UUID of the song
	LRCContent string `json:"lrc_content" binding:"required"` // LRC format lyrics
	Version    int    `json:"version"`                        // Defaults to 1
	Offset     int    `json:"offset"`                         // Offset in milliseconds
	Language   string `json:"language"`                       // e.g., "fr", "en"
}
