package commands

import "errors"

var (
	ErrInvalidURL      = errors.New("Invalid URL")
	ErrStatusCodeNotOK = errors.New("Server returned an status code out of range 2XX")
)
