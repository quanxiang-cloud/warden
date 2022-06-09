package jwts

import (
	"context"
)

// TokenStore TokenStore
type TokenStore interface {
	Create(ctx context.Context, info TokenInfo) error

	RemoveByAccess(ctx context.Context, access string) error

	RemoveByRefresh(ctx context.Context, refresh string) error

	GetByAccess(ctx context.Context, access string) (TokenInfo, error)

	GetByRefresh(ctx context.Context, refresh string) (TokenInfo, error)

	RemoveToken(ctx context.Context, jti string) error
}
