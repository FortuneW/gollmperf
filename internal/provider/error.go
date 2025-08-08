package provider

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
}

type guessError struct {
	Error map[string]interface{} `json:"error"`
}

func NewError(code int, err error) *Error {
	e := &Error{
		Code:    code,
		Message: err.Error(),
		Type:    categorizeError(err),
	}
	return e
}

func (e *Error) String() string {
	return fmt.Sprintf("code: %d, %s", e.Code, e.Message)
}

// categorizeError categorizes errors into network errors or other errors
func categorizeError(err error) string {
	// Check for network-related errors
	if ok, str := isNetworkError(err); ok {
		return str
	}

	var guessErr guessError
	if e := json.Unmarshal([]byte(err.Error()), &guessErr); e == nil {
		if guessErr.Error != nil {
			b, _ := json.Marshal(guessErr)
			return string(b)
		}
	}

	// Default to origin errors
	return err.Error()
}

// isNetworkError checks if an error is network-related
func isNetworkError(err error) (bool, string) {
	// Check for common network error types
	// Note: We can't directly import "net" in this file as it's already imported
	// We'll check the error string for network-related keywords
	errStr := err.Error()

	// Common network error indicators
	networkIndicators := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"dial tcp",
		"network is unreachable",
		"no such host",
		"i/o timeout",
		"context deadline exceeded",
		"closed by the remote host",
	}

	for _, indicator := range networkIndicators {
		if strings.Contains(strings.ToLower(errStr), indicator) {
			return true, indicator
		}
	}

	return false, ""
}
