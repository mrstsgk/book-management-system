package common

import "errors"

// Sentinel errors returned by usecase / infrastructure. The presentation
// layer's HTTPErrorHandler is the only place that maps them to HTTP status.
var (
	ErrNotFound = errors.New("not found") // → 404
	ErrInvalid  = errors.New("invalid")   // → 400
	ErrConflict = errors.New("conflict")  // → 409
)
