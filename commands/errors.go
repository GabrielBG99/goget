package commands

import "errors"

var (
	ErrUnableToDetectFolder = errors.New("The user's home folder could not be detected")
	ErrUnableToCreateFolder = errors.New("The folder could not be created")
)
