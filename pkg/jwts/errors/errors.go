package errors

import "errors"

// New returns an error that formats as the given text.
var New = errors.New

// known errors
var (
	ErrInvalidRedirectURI    = errors.New("invalid redirect uri")
	ErrInvalidAccessToken    = errors.New("invalid access token")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrExpiredAccessToken    = errors.New("expired access token")
	ErrExpiredRefreshToken   = errors.New("expired refresh token")
	ErrUnSupportedSignMethod = errors.New("unsupported sign method")
)
