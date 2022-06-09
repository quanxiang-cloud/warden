package manage

import (
	"context"
	"encoding/json"
	"github.com/quanxiang-cloud/warden/pkg/jwts"
	"github.com/quanxiang-cloud/warden/pkg/jwts/errors"
	"github.com/quanxiang-cloud/warden/pkg/jwts/models"
	"time"
)

// NewDefaultManager create to default authorization management instance
func NewDefaultManager() *Manager {
	m := NewManager()

	return m
}

// NewManager create to  management instance
func NewManager() *Manager {
	return &Manager{}
}

// Manager provide management
type Manager struct {
	codeExp        time.Duration
	tokenStore     jwts.TokenStore
	accessGenerate jwts.AccessGenerate
}

// SetCodeExp set the  code expiration time
func (m *Manager) SetCodeExp(exp time.Duration) {
	m.codeExp = exp
}

// MapAccessGenerate mapping the access token generate interface
func (m *Manager) MapAccessGenerate(gen jwts.AccessGenerate) {
	m.accessGenerate = gen
}

// MapTokenStorage mapping the token store interface
func (m *Manager) MapTokenStorage(stor jwts.TokenStore) {
	m.tokenStore = stor
}

// MustTokenStorage mandatory mapping the token store interface
func (m *Manager) MustTokenStorage(stor jwts.TokenStore, err error) {
	if err != nil {
		panic(err)
	}
	m.tokenStore = stor
}

// GenerateAccessToken generate the access token
func (m *Manager) GenerateAccessToken(ctx context.Context, jti string, otherInfo map[string]string) (jwts.TokenInfo, error) {

	ti := models.NewToken()
	ti.SetUserID(jti)

	createAt := time.Now()
	ti.SetAccessCreateAt(createAt)

	// set access token expires
	gcfg := DefaultTokenCfg
	aexp := gcfg.AccessTokenExp

	ti.SetAccessExpiresIn(aexp)
	if gcfg.IsGenerateRefresh {
		ti.SetRefreshCreateAt(createAt)
		ti.SetRefreshExpiresIn(gcfg.RefreshTokenExp)
	}

	td := &jwts.GenerateBasic{
		Jti:       jti,
		CreateAt:  createAt,
		TokenInfo: ti,
	}
	if otherInfo != nil {
		ti.OtherInfo = otherInfo
		bytes, _ := json.Marshal(otherInfo)
		td.OtherInfo = string(bytes)
	}

	av, rv, err := m.accessGenerate.Token(ctx, td, gcfg.IsGenerateRefresh)
	if err != nil {
		return nil, err
	}
	ti.SetAccess(av)

	if rv != "" {
		ti.SetRefresh(rv)
	}

	err = m.tokenStore.Create(ctx, ti)
	if err != nil {
		return nil, err
	}

	return ti, nil
}

// RefreshAccessToken refreshing an access token
func (m *Manager) RefreshAccessToken(ctx context.Context, refesh string) (jwts.TokenInfo, error) {

	ti, err := m.LoadRefreshToken(ctx, refesh)
	if err != nil {
		return nil, err
	}

	oldAccess, oldRefresh := ti.GetAccess(), ti.GetRefresh()

	td := &jwts.GenerateBasic{
		Jti:       ti.GetUserID(),
		CreateAt:  time.Now(),
		TokenInfo: ti,
	}

	rcfg := DefaultRefreshTokenCfg

	ti.SetAccessCreateAt(td.CreateAt)
	if v := rcfg.AccessTokenExp; v > 0 {
		ti.SetAccessExpiresIn(v)
	}

	if v := rcfg.RefreshTokenExp; v > 0 {
		ti.SetRefreshExpiresIn(v)
	}

	if rcfg.IsResetRefreshTime {
		ti.SetRefreshCreateAt(td.CreateAt)
	}

	tv, rv, err := m.accessGenerate.Token(ctx, td, rcfg.IsGenerateRefresh)
	if err != nil {
		return nil, err
	}

	ti.SetAccess(tv)
	if rv != "" {
		ti.SetRefresh(rv)
	}

	if err := m.tokenStore.Create(ctx, ti); err != nil {
		return nil, err
	}

	if rcfg.IsRemoveAccess {
		// remove the old access token
		if err := m.tokenStore.RemoveByAccess(ctx, oldAccess); err != nil {
			return nil, err
		}
	}

	if rcfg.IsRemoveRefreshing && rv != "" {
		// remove the old refresh token
		if err := m.tokenStore.RemoveByRefresh(ctx, oldRefresh); err != nil {
			return nil, err
		}
	}

	if rv == "" {
		ti.SetRefresh("")
		ti.SetRefreshCreateAt(time.Now())
		ti.SetRefreshExpiresIn(0)
	}

	return ti, nil
}

// RemoveAccessToken use the access token to delete the token information
func (m *Manager) RemoveAccessToken(ctx context.Context, access string) error {
	if access == "" {
		return errors.ErrInvalidAccessToken
	}
	return m.tokenStore.RemoveByAccess(ctx, access)
}

// RemoveRefreshToken use the refresh token to delete the token information
func (m *Manager) RemoveRefreshToken(ctx context.Context, refresh string) error {
	if refresh == "" {
		return errors.ErrInvalidAccessToken
	}
	return m.tokenStore.RemoveByRefresh(ctx, refresh)
}

// LoadAccessToken according to the access token for corresponding token information
func (m *Manager) LoadAccessToken(ctx context.Context, access string) (jwts.TokenInfo, error) {
	if access == "" {
		return nil, errors.ErrInvalidAccessToken
	}
	_, err := m.VerifyToken(ctx, access)
	if err != nil {
		return nil, errors.ErrInvalidAccessToken
	}
	ct := time.Now()
	ti, err := m.tokenStore.GetByAccess(ctx, access)
	if err != nil {
		return nil, err
	} else if ti == nil || ti.GetAccess() != access {
		return nil, errors.ErrInvalidAccessToken
	} else if ti.GetRefresh() != "" && ti.GetRefreshExpiresIn() != 0 &&
		ti.GetRefreshCreateAt().Add(ti.GetRefreshExpiresIn()).Before(ct) {
		return nil, errors.ErrExpiredRefreshToken
	} else if ti.GetAccessExpiresIn() != 0 &&
		ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn()).Before(ct) {
		return nil, errors.ErrExpiredAccessToken
	}
	return ti, nil
}

// LoadRefreshToken according to the refresh token for corresponding token information
func (m *Manager) LoadRefreshToken(ctx context.Context, refresh string) (jwts.TokenInfo, error) {
	if refresh == "" {
		return nil, errors.ErrInvalidRefreshToken
	}

	ti, err := m.tokenStore.GetByRefresh(ctx, refresh)
	if err != nil {
		return nil, err
	} else if ti == nil || ti.GetRefresh() != refresh {
		return nil, errors.ErrInvalidRefreshToken
	} else if ti.GetRefreshExpiresIn() != 0 && // refresh token set to not expire
		ti.GetRefreshCreateAt().Add(ti.GetRefreshExpiresIn()).Before(time.Now()) {
		return nil, errors.ErrExpiredRefreshToken
	}
	return ti, nil
}

// VerifyToken Verify token information
func (m *Manager) VerifyToken(ctx context.Context, ssoToken string) (map[string]interface{}, error) {
	if ssoToken == "" {
		return nil, errors.ErrInvalidAccessToken
	}
	res := m.accessGenerate.Verify(ctx, ssoToken)
	if res == nil {
		return nil, errors.ErrInvalidAccessToken
	}
	return res, nil
}

//RemoveToken use the jti  to delete the token  information
func (m *Manager) RemoveToken(c context.Context, jti string) (err error) {
	return m.tokenStore.RemoveToken(c, jti)
}
