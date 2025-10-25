package errors

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code,omitempty"`
}

// ErrorHandler handles errors in a secure way
type ErrorHandler struct {
	// In production, you might want to configure logging levels
	debug bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(debug bool) *ErrorHandler {
	return &ErrorHandler{
		debug: debug,
	}
}

// HandleError handles an error securely
func (h *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error, statusCode int, internalMessage string) {
	// Log the internal error details
	if internalMessage != "" {
		log.Printf("Internal error for %s %s: %s - %v", r.Method, r.URL.Path, internalMessage, err)
	} else {
		log.Printf("Internal error for %s %s: %v", r.Method, r.URL.Path, err)
	}

	// Create user-facing error message
	// In production, don't expose internal error details to users
	var userMessage string
	if h.debug {
		// In debug mode, we can show more details
		if internalMessage != "" {
			userMessage = internalMessage + ": " + err.Error()
		} else {
			userMessage = err.Error()
		}
	} else {
		// In production, use generic messages
		switch statusCode {
		case http.StatusBadRequest:
			userMessage = "Invalid request"
		case http.StatusUnauthorized:
			userMessage = "Authentication required"
		case http.StatusForbidden:
			userMessage = "Access denied"
		case http.StatusNotFound:
			userMessage = "Resource not found"
		case http.StatusTooManyRequests:
			userMessage = "Rate limit exceeded"
		case http.StatusInternalServerError:
			userMessage = "Internal server error"
		default:
			userMessage = "An error occurred"
		}
	}

	// Send JSON error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: userMessage,
		Code:  statusCode,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If we can't encode the JSON response, log it and send a simple text response
		log.Printf("Failed to encode error response: %v", err)
		http.Error(w, userMessage, statusCode)
	}
}

// HandleBadRequest handles bad request errors
func (h *ErrorHandler) HandleBadRequest(w http.ResponseWriter, r *http.Request, err error, internalMessage string) {
	h.HandleError(w, r, err, http.StatusBadRequest, internalMessage)
}

// HandleUnauthorized handles unauthorized errors
func (h *ErrorHandler) HandleUnauthorized(w http.ResponseWriter, r *http.Request, err error, internalMessage string) {
	h.HandleError(w, r, err, http.StatusUnauthorized, internalMessage)
}

// HandleForbidden handles forbidden errors
func (h *ErrorHandler) HandleForbidden(w http.ResponseWriter, r *http.Request, err error, internalMessage string) {
	h.HandleError(w, r, err, http.StatusForbidden, internalMessage)
}

// HandleNotFound handles not found errors
func (h *ErrorHandler) HandleNotFound(w http.ResponseWriter, r *http.Request, err error, internalMessage string) {
	h.HandleError(w, r, err, http.StatusNotFound, internalMessage)
}

// HandleInternalServerError handles internal server errors
func (h *ErrorHandler) HandleInternalServerError(w http.ResponseWriter, r *http.Request, err error, internalMessage string) {
	h.HandleError(w, r, err, http.StatusInternalServerError, internalMessage)
}