package config

import "errors"

var (
	// ErrMissingInputFile is returned when no input file is specified
	ErrMissingInputFile = errors.New("input file is required")
)
