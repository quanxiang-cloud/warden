package jwts

import (
	"context"
)

// Manager authorization management interface
type Manager interface {

	// GenerateAccessToken the access token
	GenerateAccessToken(ctx context.Context, jti string, otherInfo map[string]string) (accessToken TokenInfo, err error)

	// RefreshAccessToken an access token
	RefreshAccessToken(ctx context.Context, refresh string) (accessToken TokenInfo, err error)

	// RemoveAccessToken use the access token to delete the token information
	RemoveAccessToken(ctx context.Context, access string) (err error)

	// RemoveRefreshToken use the refresh token to delete the token information
	RemoveRefreshToken(ctx context.Context, refresh string) (err error)

	// LoadAccessToken according to the access token for corresponding token information
	LoadAccessToken(ctx context.Context, access string) (ti TokenInfo, err error)

	// LoadRefreshToken according to the refresh token for corresponding token information
	LoadRefreshToken(ctx context.Context, refresh string) (ti TokenInfo, err error)

	// VerifyToken Verify token information
	VerifyToken(ctx context.Context, refresh string) (map[string]interface{}, error)
	//RemoveToken use the jti  to delete the token  information
	RemoveToken(ctx context.Context, jti string) (err error)
}
