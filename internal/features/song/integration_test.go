package song

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tiavina22/lyrify-backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Artist{},
		&models.Song{},
		&models.Lyrics{},
		&models.RequestHistory{},
		&models.PendingRequest{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// setupTestRouter creates a test router with all routes
func setupTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	apiV1 := router.Group("/api/v1")
	handler := NewHandler(db)
	handler.RegisterRoutes(apiV1)

	return router
}

func TestIntegration_CreateArtistFlow(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Test: Create artist
	reqBody := CreateArtistRequest{
		Name: "Damso",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/artists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var artist models.Artist
	if err := json.Unmarshal(w.Body.Bytes(), &artist); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if artist.Name != "Damso" {
		t.Errorf("Expected artist name 'Damso', got '%s'", artist.Name)
	}

	if artist.NormalizedName != "damso" {
		t.Errorf("Expected normalized name 'damso', got '%s'", artist.NormalizedName)
	}

	// Verify artist exists in database
	var dbArtist models.Artist
	if err := db.First(&dbArtist, "id = ?", artist.ID).Error; err != nil {
		t.Errorf("Artist not found in database: %v", err)
	}
}

func TestIntegration_CreateSongFlow(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Step 1: Create artist
	artist := models.Artist{Name: "Kalash"}
	if err := db.Create(&artist).Error; err != nil {
		t.Fatalf("Failed to create test artist: %v", err)
	}

	// Step 2: Create song
	reqBody := CreateSongRequest{
		Title:     "Moment gachés",
		ArtistIDs: []string{artist.ID},
		Duration:  345,
		Album:     "Test Album",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/songs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var song models.Song
	if err := json.Unmarshal(w.Body.Bytes(), &song); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if song.Title != "Moment gachés" {
		t.Errorf("Expected song title 'Moment gachés', got '%s'", song.Title)
	}

	if song.NormalizedTitle != "moment gaches" {
		t.Errorf("Expected normalized title 'moment gaches', got '%s'", song.NormalizedTitle)
	}

	// Step 3: Verify song has artist relationship
	var dbSong models.Song
	if err := db.Preload("Artists").First(&dbSong, "id = ?", song.ID).Error; err != nil {
		t.Fatalf("Failed to load song from database: %v", err)
	}

	if len(dbSong.Artists) != 1 {
		t.Errorf("Expected 1 artist, got %d", len(dbSong.Artists))
	}

	if dbSong.Artists[0].Name != "Kalash" {
		t.Errorf("Expected artist 'Kalash', got '%s'", dbSong.Artists[0].Name)
	}
}

func TestIntegration_CreateLyricsFlow(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Setup: Create artist and song
	artist := models.Artist{Name: "Test Artist"}
	db.Create(&artist)

	song := models.Song{
		Title:    "Test Song",
		Duration: 180,
	}
	db.Create(&song)
	db.Model(&song).Association("Artists").Append(&artist)

	// Test: Create lyrics
	lrcContent := `[00:12.50]First line
[00:18.30]Second line
[00:24.10]Third line`

	reqBody := CreateLyricsRequest{
		SongID:     song.ID,
		LRCContent: lrcContent,
		Language:   "en",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/lyrics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var lyrics models.Lyrics
	if err := json.Unmarshal(w.Body.Bytes(), &lyrics); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if lyrics.SongID != song.ID {
		t.Errorf("Expected song_id '%s', got '%s'", song.ID, lyrics.SongID)
	}

	if lyrics.Duration != song.Duration {
		t.Errorf("Expected duration %d, got %d", song.Duration, lyrics.Duration)
	}

	if lyrics.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", lyrics.Language)
	}

	// Verify timestamps were extracted
	if lyrics.LRCTimestamps == "" {
		t.Error("LRC timestamps should not be empty")
	}
}

func TestIntegration_CreateLyricsWithGarbage(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Setup: Create artist and song
	artist := models.Artist{Name: "Test Artist"}
	db.Create(&artist)

	song := models.Song{
		Title:    "Test Song",
		Duration: 180,
	}
	db.Create(&song)
	db.Model(&song).Association("Artists").Append(&artist)

	// Test: Create lyrics with HTML garbage (should be cleaned)
	lrcContent := `BACK TO HOME
Navigation menu
Preview
[00:12.50]First line
Some random text
[00:18.30]Second line
Footer content
[00:24.10]Third line`

	reqBody := CreateLyricsRequest{
		SongID:     song.ID,
		LRCContent: lrcContent,
		Language:   "en",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/lyrics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var lyrics models.Lyrics
	if err := json.Unmarshal(w.Body.Bytes(), &lyrics); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify garbage was removed
	if bytes.Contains([]byte(lyrics.LRCContent), []byte("BACK TO HOME")) {
		t.Error("LRC content should not contain 'BACK TO HOME'")
	}
	if bytes.Contains([]byte(lyrics.LRCContent), []byte("Navigation menu")) {
		t.Error("LRC content should not contain 'Navigation menu'")
	}

	// Verify valid lines were kept
	if !bytes.Contains([]byte(lyrics.LRCContent), []byte("[00:12.50]First line")) {
		t.Error("LRC content should contain '[00:12.50]First line'")
	}
}

func TestIntegration_MatchFlow_Found(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Setup: Create complete song with lyrics
	artist := models.Artist{Name: "Damso"}
	db.Create(&artist)

	song := models.Song{
		Title:            "Débrouillard",
		Duration:         243,
		FileHash:         "abc123def456",
		NormalizedTitle:  "debrouillard",
		NormalizedArtist: "damso",
	}
	db.Create(&song)
	db.Model(&song).Association("Artists").Append(&artist)

	lyrics := models.Lyrics{
		SongID:     song.ID,
		LRCContent: "[00:16.46] Lyrics here",
		Duration:   243,
		Version:    1,
	}
	db.Create(&lyrics)

	// Test: Match by hash
	reqBody := MatchRequest{
		Title:    "Débrouillard",
		Artist:   "Damso",
		Duration: 243,
		Hash:     "abc123def456",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/match", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-ID", "test-device-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response MatchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Found {
		t.Error("Expected Found to be true")
	}

	if response.SongID != song.ID {
		t.Errorf("Expected song_id '%s', got '%s'", song.ID, response.SongID)
	}

	if response.LyricsID != lyrics.ID {
		t.Errorf("Expected lyrics_id '%s', got '%s'", lyrics.ID, response.LyricsID)
	}

	// Verify request history was logged
	var history models.RequestHistory
	if err := db.First(&history, "search_title = ?", "Débrouillard").Error; err != nil {
		t.Errorf("Request history not found: %v", err)
	}

	if !history.Found {
		t.Error("Request history should show Found = true")
	}
}

func TestIntegration_MatchFlow_NotFound(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Test: Match non-existent song
	reqBody := MatchRequest{
		Title:    "Nonexistent Song",
		Artist:   "Unknown Artist",
		Duration: 180,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/match", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-ID", "test-device-456")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify request history was logged
	var history models.RequestHistory
	if err := db.First(&history, "search_title = ?", "Nonexistent Song").Error; err != nil {
		t.Errorf("Request history not found: %v", err)
	}

	if history.Found {
		t.Error("Request history should show Found = false")
	}

	// Verify pending request was created
	var pending models.PendingRequest
	if err := db.First(&pending, "title = ? AND artist = ? AND duration = ?",
		"Nonexistent Song", "Unknown Artist", 180).Error; err != nil {
		t.Errorf("Pending request not found: %v", err)
	}

	if pending.RequestCount != 1 {
		t.Errorf("Expected request_count 1, got %d", pending.RequestCount)
	}
}

func TestIntegration_SearchFlow(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Setup: Create multiple songs
	artist := models.Artist{Name: "Test Artist"}
	db.Create(&artist)

	songs := []models.Song{
		{Title: "Song One", Duration: 180, NormalizedTitle: "song one"},
		{Title: "Song Two", Duration: 200, NormalizedTitle: "song two"},
		{Title: "Another Song", Duration: 220, NormalizedTitle: "another song"},
	}
	for i := range songs {
		db.Create(&songs[i])
		db.Model(&songs[i]).Association("Artists").Append(&artist)
	}

	// Test: Search by title
	req, _ := http.NewRequest("GET", "/api/v1/songs?title=Song&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response SearchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// All 3 songs contain "song" in their title
	if len(response.Songs) != 3 {
		t.Errorf("Expected 3 songs (all contain 'song'), got %d", len(response.Songs))
	}
}

func TestIntegration_GetSongByID(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Setup: Create song with artist and lyrics
	artist := models.Artist{Name: "Test Artist"}
	db.Create(&artist)

	song := models.Song{Title: "Test Song", Duration: 180}
	db.Create(&song)
	db.Model(&song).Association("Artists").Append(&artist)

	lyrics := models.Lyrics{
		SongID:     song.ID,
		LRCContent: "[00:12.50]Line",
		Duration:   180,
	}
	db.Create(&lyrics)

	// Test: Get song by ID
	req, _ := http.NewRequest("GET", "/api/v1/songs/"+song.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resultSong models.Song
	if err := json.Unmarshal(w.Body.Bytes(), &resultSong); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resultSong.ID != song.ID {
		t.Errorf("Expected song ID '%s', got '%s'", song.ID, resultSong.ID)
	}

	// Verify relationships are loaded
	if len(resultSong.Artists) == 0 {
		t.Error("Expected artists to be loaded")
	}

	if len(resultSong.Lyrics) == 0 {
		t.Error("Expected lyrics to be loaded")
	}
}

func TestIntegration_GetLyricsByID(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Setup: Create song and lyrics
	song := models.Song{Title: "Test Song", Duration: 180}
	db.Create(&song)

	lyrics := models.Lyrics{
		SongID:     song.ID,
		LRCContent: "[00:12.50]Test line",
		Duration:   180,
		Language:   "en",
	}
	db.Create(&lyrics)

	// Test: Get lyrics by ID
	req, _ := http.NewRequest("GET", "/api/v1/lyrics/"+lyrics.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resultLyrics models.Lyrics
	if err := json.Unmarshal(w.Body.Bytes(), &resultLyrics); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resultLyrics.ID != lyrics.ID {
		t.Errorf("Expected lyrics ID '%s', got '%s'", lyrics.ID, resultLyrics.ID)
	}

	if resultLyrics.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", resultLyrics.Language)
	}
}

func TestIntegration_GetPendingRequests(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Setup: Create pending requests
	requests := []models.PendingRequest{
		{Title: "Song A", Artist: "Artist A", Duration: 180, RequestCount: 5, Priority: 5},
		{Title: "Song B", Artist: "Artist B", Duration: 200, RequestCount: 3, Priority: 3},
		{Title: "Song C", Artist: "Artist C", Duration: 220, RequestCount: 10, Priority: 10},
	}
	for i := range requests {
		db.Create(&requests[i])
	}

	// Test: Get pending requests (should be ordered by priority)
	req, _ := http.NewRequest("GET", "/api/v1/pending-requests?limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response PendingRequestsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response.Requests) != 3 {
		t.Errorf("Expected 3 requests, got %d", len(response.Requests))
	}

	// Verify ordering (highest priority first)
	if response.Requests[0].Priority != 10 {
		t.Errorf("First request should have priority 10, got %d", response.Requests[0].Priority)
	}
}
