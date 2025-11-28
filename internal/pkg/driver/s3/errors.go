package s3

import "errors"

var (
	// ErrInvalidCredentials is returned when authentication fails.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrTLSNotConfigured is returned when TLS is requested but not configured.
	ErrTLSNotConfigured = errors.New("TLS not configured")
)
