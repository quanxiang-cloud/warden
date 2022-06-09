package models

import (
	"github.com/quanxiang-cloud/warden/pkg/jwts"
	"time"
)

// NewToken create to token model instance
func NewToken() *Token {
	return &Token{}
}

// Token token model
type Token struct {
	UserID           string        `bson:"UserID"`
	Access           string        `bson:"Access"`
	AccessCreateAt   time.Time     `bson:"AccessCreateAt"`
	AccessExpiresIn  time.Duration `bson:"AccessExpiresIn"`
	Refresh          string        `bson:"Refresh"`
	RefreshCreateAt  time.Time     `bson:"RefreshCreateAt"`
	RefreshExpiresIn time.Duration `bson:"RefreshExpiresIn"`

	OtherInfo map[string]string `bson:"OtherInfo"`
}

// New create to token model instance
func (t *Token) New() jwts.TokenInfo {
	return NewToken()
}

// GetUserID the user id
func (t *Token) GetUserID() string {
	return t.UserID
}

// SetUserID the user id
func (t *Token) SetUserID(userID string) {
	t.UserID = userID
}

// GetAccess access Token
func (t *Token) GetAccess() string {
	return t.Access
}

// SetAccess access Token
func (t *Token) SetAccess(access string) {
	t.Access = access
}

// GetAccessCreateAt create Time
func (t *Token) GetAccessCreateAt() time.Time {
	return t.AccessCreateAt
}

// SetAccessCreateAt create Time
func (t *Token) SetAccessCreateAt(createAt time.Time) {
	t.AccessCreateAt = createAt
}

// GetAccessExpiresIn the lifetime in seconds of the access token
func (t *Token) GetAccessExpiresIn() time.Duration {
	return t.AccessExpiresIn
}

// SetAccessExpiresIn the lifetime in seconds of the access token
func (t *Token) SetAccessExpiresIn(exp time.Duration) {
	t.AccessExpiresIn = exp
}

// GetRefresh refresh Token
func (t *Token) GetRefresh() string {
	return t.Refresh
}

// SetRefresh refresh Token
func (t *Token) SetRefresh(refresh string) {
	t.Refresh = refresh
}

// GetRefreshCreateAt create Time
func (t *Token) GetRefreshCreateAt() time.Time {
	return t.RefreshCreateAt
}

// SetRefreshCreateAt create Time
func (t *Token) SetRefreshCreateAt(createAt time.Time) {
	t.RefreshCreateAt = createAt
}

// GetRefreshExpiresIn the lifetime in seconds of the refresh token
func (t *Token) GetRefreshExpiresIn() time.Duration {
	return t.RefreshExpiresIn
}

// SetRefreshExpiresIn the lifetime in seconds of the refresh token
func (t *Token) SetRefreshExpiresIn(exp time.Duration) {
	t.RefreshExpiresIn = exp
}

// GetOtherInfo GetOtherInfo
func (t *Token) GetOtherInfo() map[string]string {

	return t.OtherInfo
}

// SetOtherInfo SetOtherInfo
func (t *Token) SetOtherInfo(data map[string]string) {
	t.OtherInfo = data
}
