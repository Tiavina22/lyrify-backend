package errors

import (
	"errors"
	"fmt"
)

// Common application errors
var (
	// Song errors
	ErrSongNotFound    = errors.New("song not found")
	ErrDuplicateSong   = errors.New("song already exists")
	ErrInvalidDuration = errors.New("invalid song duration")
	ErrInvalidSongData = errors.New("invalid song data")

	// Lyrics errors
	ErrLyricsNotFound    = errors.New("lyrics not found")
	ErrInvalidLRCFormat  = errors.New("invalid LRC format")
	ErrDurationMismatch  = errors.New("lyrics duration does not match song duration")
	ErrInvalidLyricsData = errors.New("invalid lyrics data")

	// Artist errors
	ErrArtistNotFound    = errors.New("artist not found")
	ErrDuplicateArtist   = errors.New("artist already exists")
	ErrInvalidArtistData = errors.New("invalid artist data")

	// Request errors
	ErrInvalidRequest   = errors.New("invalid request")
	ErrMissingParameter = errors.New("missing required parameter")

	// Database errors
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrDatabaseQuery      = errors.New("database query failed")

	// Validation errors
	ErrValidationFailed = errors.New("validation failed")
)

// APIError represents a structured error response for API endpoints
type APIError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewAPIError creates a new APIError
func NewAPIError(code int, message string, details map[string]interface{}) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common HTTP error constructors
func BadRequest(message string, details map[string]interface{}) *APIError {
	return NewAPIError(400, message, details)
}

func NotFound(message string) *APIError {
	return NewAPIError(404, message, nil)
}

func InternalServerError(message string, details map[string]interface{}) *APIError {
	return NewAPIError(500, message, details)
}

func ValidationError(message string, details map[string]interface{}) *APIError {
	return NewAPIError(422, message, details)
}

func Conflict(message string, details map[string]interface{}) *APIError {
	return NewAPIError(409, message, details)
}

// WrapError wraps a standard error into an APIError based on its type
func WrapError(err error) *APIError {
	switch {
	case errors.Is(err, ErrSongNotFound), errors.Is(err, ErrLyricsNotFound), errors.Is(err, ErrArtistNotFound):
		return NotFound(err.Error())

	case errors.Is(err, ErrDuplicateSong), errors.Is(err, ErrDuplicateArtist):
		return Conflict(err.Error(), nil)

	case errors.Is(err, ErrInvalidLRCFormat), errors.Is(err, ErrDurationMismatch),
		errors.Is(err, ErrInvalidDuration), errors.Is(err, ErrValidationFailed),
		errors.Is(err, ErrInvalidSongData), errors.Is(err, ErrInvalidLyricsData),
		errors.Is(err, ErrInvalidArtistData):
		return ValidationError(err.Error(), nil)

	case errors.Is(err, ErrInvalidRequest), errors.Is(err, ErrMissingParameter):
		return BadRequest(err.Error(), nil)

	default:
		return InternalServerError("Internal server error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
