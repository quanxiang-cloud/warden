package jwts

import (
	"time"
)

type (
	// TokenInfo the token information model interface
	TokenInfo interface {
		New() TokenInfo

		GetUserID() string
		SetUserID(string)

		GetAccess() string
		SetAccess(string)
		GetAccessCreateAt() time.Time
		SetAccessCreateAt(time.Time)
		GetAccessExpiresIn() time.Duration
		SetAccessExpiresIn(time.Duration)

		GetRefresh() string
		SetRefresh(string)
		GetRefreshCreateAt() time.Time
		SetRefreshCreateAt(time.Time)
		GetRefreshExpiresIn() time.Duration
		SetRefreshExpiresIn(time.Duration)

		GetOtherInfo() map[string]string
		SetOtherInfo(map[string]string)
	}
)
