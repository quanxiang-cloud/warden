package jwts

import (
	"context"
	"time"
)

type (
	// GenerateBasic provide the basis of the generated token data
	GenerateBasic struct {
		Jti       string
		CreateAt  time.Time
		TokenInfo TokenInfo

		OtherInfo string
	}
	// AccessGenerate generate the access and refresh tokens interface
	AccessGenerate interface {
		Token(ctx context.Context, data *GenerateBasic, isGenRefresh bool) (access, refresh string, err error)
		Verify(ctx context.Context, ssoToken string) map[string]interface{}
	}
)
