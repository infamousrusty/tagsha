package api

import (
	"encoding/json"
	"net/http"
)

// ErrorCode is a machine-readable error identifier.
type ErrorCode string

const (
	ErrInvalidInput   ErrorCode = "INVALID_INPUT"
	ErrRepoNotFound   ErrorCode = "REPOSITORY_NOT_FOUND"
	ErrRateLimited    ErrorCode = "RATE_LIMITED"
	ErrGitHubAPIError ErrorCode = "GITHUB_API_ERROR"
	ErrInternalError  ErrorCode = "INTERNAL_ERROR"
)

// APIError is the structured error response body.
type APIError struct {
	Error     ErrorCode `json:"error"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id"`
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// At this point the header is already sent — log only.
		_ = err
	}
}

func writeError(w http.ResponseWriter, status int, code ErrorCode, message, requestID string) {
	writeJSON(w, status, APIError{
		Error:     code,
		Message:   message,
		RequestID: requestID,
	})
}
