package generates

import (
	"context"
	"encoding/base64"
	"github.com/quanxiang-cloud/warden/pkg/jwts"
	"github.com/quanxiang-cloud/warden/pkg/jwts/errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"strings"
	"time"
)

const (
	kid = "kid"
	es  = "ES"
	rs  = "RS"
	ps  = "PS"
	hs  = "HS"
)

// JWTAccessClaims jwt claims
type JWTAccessClaims struct {
	jwt.StandardClaims
}

// Valid claims verification
func (a *JWTAccessClaims) Valid() error {
	if time.Unix(a.ExpiresAt, 0).Before(time.Now()) {
		return errors.ErrInvalidAccessToken
	}
	return nil
}

// NewJWTAccessGenerate create to generate the jwt access token instance
func NewJWTAccessGenerate(kid string, key, pubKey []byte, method jwt.SigningMethod) *JWTAccessGenerate {
	return &JWTAccessGenerate{
		SignedKeyID:  kid,
		SignedKey:    key,
		PubKey:       pubKey,
		SignedMethod: method,
	}
}

// JWTAccessGenerate generate the jwt access token
type JWTAccessGenerate struct {
	SignedKeyID  string
	SignedKey    []byte
	PubKey       []byte
	SignedMethod jwt.SigningMethod
}

// Token based on the UUID generated token
func (a *JWTAccessGenerate) Token(ctx context.Context, data *jwts.GenerateBasic, isGenRefresh bool) (string, string, error) {
	claims := &JWTAccessClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        data.Jti,
			ExpiresAt: data.TokenInfo.GetAccessCreateAt().Add(data.TokenInfo.GetAccessExpiresIn()).Unix(),
			Subject:   data.OtherInfo,
		},
	}

	token := jwt.NewWithClaims(a.SignedMethod, claims)
	if a.SignedKeyID != "" {
		token.Header[kid] = a.SignedKeyID
	}
	var key interface{}
	if a.isEs() {
		v, err := jwt.ParseECPrivateKeyFromPEM(a.SignedKey)
		if err != nil {
			return "", "", err
		}
		key = v
	} else if a.isRsOrPS() {
		v, err := jwt.ParseRSAPrivateKeyFromPEM(a.SignedKey)
		if err != nil {
			return "", "", err
		}
		key = v
	} else if a.isHs() {
		key = a.SignedKey
	} else {
		return "", "", errors.ErrUnSupportedSignMethod
	}

	access, err := token.SignedString(key)
	if err != nil {
		return "", "", err
	}
	refresh := ""

	if isGenRefresh {
		t := uuid.NewSHA1(uuid.Must(uuid.NewRandom()), []byte(access)).String()
		refresh = base64.URLEncoding.EncodeToString([]byte(t))
		refresh = strings.ToUpper(strings.TrimRight(refresh, "="))
	}

	return access, refresh, nil
}

func (a *JWTAccessGenerate) isEs() bool {
	return strings.HasPrefix(a.SignedMethod.Alg(), es)
}

func (a *JWTAccessGenerate) isRsOrPS() bool {
	isRs := strings.HasPrefix(a.SignedMethod.Alg(), rs)
	isPs := strings.HasPrefix(a.SignedMethod.Alg(), ps)
	return isRs || isPs
}

func (a *JWTAccessGenerate) isHs() bool {
	return strings.HasPrefix(a.SignedMethod.Alg(), hs)
}

// Verify Verify token
func (a *JWTAccessGenerate) Verify(ctx context.Context, ssoToken string) map[string]interface{} {
	parts := strings.Split(ssoToken, ".")
	var key []byte = nil
	if a.PubKey != nil {
		key = a.PubKey
	} else {
		key = a.SignedKey
	}
	token, _ := jwt.Parse(ssoToken, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if token == nil {
		return nil
	}
	claims := token.Claims.(jwt.MapClaims)
	if claims.VerifyExpiresAt(time.Now().Unix(), false) == false {
		return nil
	}
	err := a.SignedMethod.Verify(strings.Join(parts[0:2], "."), parts[2], key)
	if err != nil {
		return nil
	}
	if !token.Valid {
		return nil
	}
	if token != nil {
		if !token.Valid {
			return nil
		}
		return token.Claims.(jwt.MapClaims)
	}
	return nil
}
