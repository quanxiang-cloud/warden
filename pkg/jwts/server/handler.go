package server

import (
	"github.com/quanxiang-cloud/warden/pkg/jwts"
	"net/http"
	"time"
)

type (

	// UserAuthorizationHandler get user id from request authorization
	UserAuthorizationHandler func(w http.ResponseWriter, r *http.Request) (userID string, err error)

	//RefreshingValidationHandler check if refresh_token is still valid. eg no revocation or other
	RefreshingValidationHandler func(ti jwts.TokenInfo) (allowed bool, err error)

	// AccessTokenExpHandler set expiration date for the access token
	AccessTokenExpHandler func(w http.ResponseWriter, r *http.Request) (exp time.Duration, err error)

	// ExtensionFieldsHandler in response to the access token with the extension of the field
	ExtensionFieldsHandler func(ti jwts.TokenInfo) (fieldsValue map[string]interface{})
)
