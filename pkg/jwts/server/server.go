package server

import (
	"context"
	"github.com/quanxiang-cloud/warden/pkg/jwts"
	"github.com/quanxiang-cloud/warden/pkg/jwts/errors"
	"strings"
	"time"
)

const (
	acsToken     = "access_token"
	expiry       = "expiry"
	refreshToken = "refresh_token"

	bearer = "Bearer "
)

// NewDefaultServer create a default authorization server
func NewDefaultServer(manager jwts.Manager) *Server {
	return NewServer(NewConfig(), manager)
}

// NewServer create authorization server
func NewServer(cfg *Config, manager jwts.Manager) *Server {
	srv := &Server{
		Config:  cfg,
		Manager: manager,
	}
	return srv
}

// Server Provide authorization server
type Server struct {
	Config  *Config
	Manager jwts.Manager

	RefreshingValidationHandler RefreshingValidationHandler
	ExtensionFieldsHandler      ExtensionFieldsHandler
	AccessTokenExpHandler       AccessTokenExpHandler
}

// GetAuthorizeData get authorization response data
func (s *Server) GetAuthorizeData(ti jwts.TokenInfo) map[string]interface{} {
	return s.GetTokenData(ti)
}

// GetAccessToken access token
func (s *Server) GetAccessToken(ctx context.Context, jti string, otherInfo map[string]string) (jwts.TokenInfo, error) {

	ti, err := s.Manager.GenerateAccessToken(ctx, jti, otherInfo)
	if err != nil {
		switch err {
		default:
			return nil, err
		}
	}
	return ti, nil

}

// GetRefreshAccessToken access token
func (s *Server) GetRefreshAccessToken(ctx context.Context, refresh string) (jwts.TokenInfo, error) {

	ti, err := s.Manager.RefreshAccessToken(ctx, refresh)
	if err != nil {
		switch err {
		default:
			return nil, err
		}
	}
	return ti, nil

}

// GetTokenData token data
func (s *Server) GetTokenData(ti jwts.TokenInfo) map[string]interface{} {
	data := map[string]interface{}{
		acsToken: ti.GetAccess(),
		expiry:   time.Now().Add(ti.GetAccessExpiresIn()),
	}

	if refresh := ti.GetRefresh(); refresh != "" {
		data[refreshToken] = refresh
	}

	if fn := s.ExtensionFieldsHandler; fn != nil {
		ext := fn(ti)
		for k, v := range ext {
			if _, ok := data[k]; ok {
				continue
			}
			data[k] = v
		}
	}
	return data
}

// HandleTokenRequest token request handling
func (s *Server) HandleTokenRequest(c context.Context, jti string, otherInfo map[string]string) (token map[string]interface{}, err error) {
	ti, err := s.GetAccessToken(c, jti, otherInfo)
	if err != nil {
		return nil, err
	}

	return s.GetTokenData(ti), nil
}

// HandleRefreshTokenRequest newToken request handling
func (s *Server) HandleRefreshTokenRequest(c context.Context, refresh string) (token map[string]interface{}, err error) {

	ti, err := s.GetRefreshAccessToken(c, refresh)
	if err != nil {
		return nil, err
	}

	return s.GetTokenData(ti), nil
}

// BearerAuth parse bearer token
func (s *Server) BearerAuth(ctx context.Context, authorization, accessToken string) (string, bool) {
	auth := authorization
	prefix := bearer
	token := accessToken

	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	}

	return token, token != ""
}

// ValidationBearerToken validation the bearer tokens
func (s *Server) ValidationBearerToken(ctx context.Context, authorization, token string) (jwts.TokenInfo, error) {

	accessToken, ok := s.BearerAuth(ctx, authorization, token)
	if !ok {
		return nil, errors.ErrInvalidAccessToken
	}

	return s.Manager.LoadAccessToken(ctx, accessToken)
}

// SSOValidationBearerToken validation the bearer tokens
func (s *Server) SSOValidationBearerToken(ctx context.Context, authorization, token string) (jwts.TokenInfo, error) {
	accessToken, ok := s.BearerAuth(ctx, authorization, token)
	if !ok {
		return nil, errors.ErrInvalidAccessToken
	}
	return s.Manager.LoadAccessToken(ctx, accessToken)
}
