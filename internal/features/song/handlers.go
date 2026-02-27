package song

import (
	"fmt"
	"net/http"

	"github.com/Tiavina22/lyrify-backend/internal/errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler handles HTTP requests for song operations
type Handler struct {
	service *Service
}

// NewHandler creates a new song handler
func NewHandler(db *gorm.DB) *Handler {
	repo := NewRepository(db)
	service := NewService(repo)
	return &Handler{service: service}
}

// RegisterRoutes registers all song-related routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/match", h.Match)
	router.GET("/songs/:id", h.GetSong)
	router.GET("/lyrics/:id", h.GetLyrics)
	router.GET("/songs", h.Search)
	router.GET("/pending-requests", h.GetPendingRequests)

	// Creation endpoints
	router.POST("/artists", h.CreateArtist)
	router.POST("/songs", h.CreateSong)
	router.POST("/lyrics", h.CreateLyrics)
}

// Match handles POST /match - Find matching song and lyrics
// @Summary Match song with lyrics
// @Description Find the best matching song and lyrics based on title, artist, and duration
// @Tags songs
// @Accept json
// @Produce json
// @Param request body MatchRequest true "Match Request"
// @Success 200 {object} MatchResponse "Match found"
// @Failure 404 {object} errors.APIError "Song not found"
// @Failure 400 {object} errors.APIError "Bad request"
// @Router /api/v1/match [post]
func (h *Handler) Match(c *gin.Context) {
	var req MatchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := errors.BadRequest("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(apiErr.Code, apiErr)
		return
	}

	// Extract device_id and IP address from request
	deviceID := c.GetHeader("X-Device-ID")
	ipAddress := c.ClientIP()

	response, err := h.service.Match(req, deviceID, ipAddress)
	if err != nil {
		apiErr := errors.WrapError(err)
		c.JSON(apiErr.Code, apiErr)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSong handles GET /songs/:id - Retrieve song by ID
// @Summary Get song by ID
// @Description Retrieve a song with all its relationships (artists, lyrics)
// @Tags songs
// @Produce json
// @Param id path string true "Song ID"
// @Success 200 {object} models.Song "Song found"
// @Failure 404 {object} errors.APIError "Song not found"
// @Router /api/v1/songs/{id} [get]
func (h *Handler) GetSong(c *gin.Context) {
	id := c.Param("id")

	song, err := h.service.GetSongByID(id)
	if err != nil {
		apiErr := errors.WrapError(err)
		c.JSON(apiErr.Code, apiErr)
		return
	}

	c.JSON(http.StatusOK, song)
}

// GetLyrics handles GET /lyrics/:id - Retrieve lyrics by ID
// @Summary Get lyrics by ID
// @Description Retrieve lyrics with its associated song
// @Tags lyrics
// @Produce json
// @Param id path string true "Lyrics ID"
// @Success 200 {object} models.Lyrics "Lyrics found"
// @Failure 404 {object} errors.APIError "Lyrics not found"
// @Router /api/v1/lyrics/{id} [get]
func (h *Handler) GetLyrics(c *gin.Context) {
	id := c.Param("id")

	lyrics, err := h.service.GetLyricsByID(id)
	if err != nil {
		apiErr := errors.WrapError(err)
		c.JSON(apiErr.Code, apiErr)
		return
	}

	c.JSON(http.StatusOK, lyrics)
}

// Search handles GET /songs - Search for songs
// @Summary Search songs
// @Description Search for songs by title and/or artist
// @Tags songs
// @Produce json
// @Param title query string false "Song title"
// @Param artist query string false "Artist name"
// @Param limit query int false "Result limit (max 100)" default(20)
// @Success 200 {object} SearchResponse "Songs found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Router /api/v1/songs [get]
func (h *Handler) Search(c *gin.Context) {
	var req SearchRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		apiErr := errors.BadRequest("Invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(apiErr.Code, apiErr)
		return
	}

	response, err := h.service.Search(req)
	if err != nil {
		apiErr := errors.WrapError(err)
		c.JSON(apiErr.Code, apiErr)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPendingRequests handles GET /pending-requests - Get top pending requests
// @Summary Get pending requests
// @Description Get top pending requests ordered by priority (for admin/contributors)
// @Tags pending-requests
// @Produce json
// @Param limit query int false "Result limit" default(50)
// @Success 200 {object} PendingRequestsResponse "Pending requests"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Router /api/v1/pending-requests [get]
func (h *Handler) GetPendingRequests(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		// Parse limit from query string
		var parsed int
		if _, err := fmt.Sscanf(l, "%d", &parsed); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	response, err := h.service.GetTopPendingRequests(limit)
	if err != nil {
		apiErr := errors.WrapError(err)
		c.JSON(apiErr.Code, apiErr)
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateArtist handles POST /artists - Create a new artist
// @Summary Create a new artist
// @Description Create a new artist in the database
// @Tags artists
// @Accept json
// @Produce json
// @Param request body CreateArtistRequest true "Artist details"
// @Success 201 {object} models.Artist "Artist created"
// @Failure 400 {object} errors.APIError "Bad request"
// @Router /api/v1/artists [post]
func (h *Handler) CreateArtist(c *gin.Context) {
	var req CreateArtistRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := errors.BadRequest("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(apiErr.Code, apiErr)
		return
	}

	artist, err := h.service.CreateArtist(req)
	if err != nil {
		apiErr := errors.WrapError(err)
		c.JSON(apiErr.Code, apiErr)
		return
	}

	c.JSON(http.StatusCreated, artist)
}

// CreateSong handles POST /songs - Create a new song
// @Summary Create a new song
// @Description Create a new song with associated artists
// @Tags songs
// @Accept json
// @Produce json
// @Param request body CreateSongRequest true "Song details"
// @Success 201 {object} models.Song "Song created"
// @Failure 400 {object} errors.APIError "Bad request"
// @Failure 404 {object} errors.APIError "Artist not found"
// @Router /api/v1/songs [post]
func (h *Handler) CreateSong(c *gin.Context) {
	var req CreateSongRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := errors.BadRequest("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(apiErr.Code, apiErr)
		return
	}

	song, err := h.service.CreateSong(req)
	if err != nil {
		apiErr := errors.WrapError(err)
		c.JSON(apiErr.Code, apiErr)
		return
	}

	c.JSON(http.StatusCreated, song)
}

// CreateLyrics handles POST /lyrics - Create lyrics for a song
// @Summary Create lyrics for a song
// @Description Create synchronized lyrics in LRC format for an existing song
// @Tags lyrics
// @Accept json
// @Produce json
// @Param request body CreateLyricsRequest true "Lyrics details"
// @Success 201 {object} models.Lyrics "Lyrics created"
// @Failure 400 {object} errors.APIError "Bad request"
// @Failure 404 {object} errors.APIError "Song not found"
// @Router /api/v1/lyrics [post]
func (h *Handler) CreateLyrics(c *gin.Context) {
	var req CreateLyricsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := errors.BadRequest("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(apiErr.Code, apiErr)
		return
	}

	lyrics, err := h.service.CreateLyrics(req)
	if err != nil {
		apiErr := errors.WrapError(err)
		c.JSON(apiErr.Code, apiErr)
		return
	}

	c.JSON(http.StatusCreated, lyrics)
}
