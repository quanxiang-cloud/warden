package server

import (
	"net/http"
	"time"
)

// Config configuration parameters
type Config struct {
}

// NewConfig create to configuration instance
func NewConfig() *Config {
	return &Config{}
}

// AuthorizeRequest authorization request
type AuthorizeRequest struct {
	UserID         string
	AccessTokenExp time.Duration
	Request        *http.Request
}
