package filemgr

import "errors"

var (
	ErrInvalidURL                 = errors.New("Invalid URL")
	ErrUnableToForceDownload      = errors.New("The existing download file could not be deleted")
	ErrUnableToCreateDownloadFile = errors.New("The download file could not be created")
	ErrDownloadFileAlreadyExists  = errors.New("The download file already exists")
	ErrNotAcceptRange             = errors.New("The download URL does not accept range download")
	ErrNoContentLength            = errors.New("The URL does not provide a \"Content-Length\" header")
	ErrStatusCodeNotOK            = errors.New("Server returned an status code out of range 2XX")
	ErrUnableToRequest            = errors.New("An error occured while connecting to the provided URL")
	ErrInvalidContentLength       = errors.New("The server returned an invalid \"Content-Length\" header")
	ErrInvalidNumberOfParts       = errors.New("The number of parts should be greater than 0")
	ErrRemovingParts              = errors.New("Error removing part files")
)
