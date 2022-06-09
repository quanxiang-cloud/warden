package org

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	error2 "github.com/quanxiang-cloud/cabin/error"
	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"github.com/quanxiang-cloud/warden/internal/jwtserver"
	"github.com/quanxiang-cloud/warden/pkg/configs"
	"github.com/quanxiang-cloud/warden/pkg/jwts/server"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Org org interface
type Org interface {
	UpdateUserStatus(ctx context.Context, r *http.Request, w http.ResponseWriter, data *UpdateUserStatusRequest)
	UpdateUsersStatus(ctx context.Context, r *http.Request, w http.ResponseWriter, data *UpdateListUserStatusRequest)
	AdminResetPassword(ctx context.Context, r *http.Request, w http.ResponseWriter, data *AdminResetPasswordRequest)
	UserResetPassword(ctx context.Context, r *http.Request, w http.ResponseWriter, data *UserResetPasswordRequest)
	UserForgetResetPassword(ctx context.Context, r *http.Request, w http.ResponseWriter, data *UserForgetResetRequest)
}

// NewOrg new
func NewOrg(conf configs.Config, redisClient redis.UniversalClient) Org {

	return &org{
		s:           jwtserver.NewServer(),
		client:      client.New(conf.InternalNet),
		conf:        conf,
		redisClient: redisClient,
	}
}

type org struct {
	s           *server.Server
	client      http.Client
	conf        configs.Config
	redisClient redis.UniversalClient
}

// UpdateUserStatusRequest update user status request
type UpdateUserStatusRequest struct {
	ID        string `json:"id" binding:"required"`
	UseStatus int    `json:"useStatus" binding:"required"` //状态：1正常，-2禁用，2激活==1 （与账号库相同）批量不支持删除，删除需要谨慎操作
	UpdatedBy string `json:"updatedBy"`
	TenantID  string `json:"tenantID"`
}

// UpdateUserStatusResponse update user status response
type UpdateUserStatusResponse struct {
}

// UpdateUserStatus update user status
func (o *org) UpdateUserStatus(ctx context.Context, r *http.Request, w http.ResponseWriter, data *UpdateUserStatusRequest) {
	response, err := DealRequest(o.client, o.conf.OrgAPIs.Host, r, o.conf.OrgAPIs.UpdateUserStatusURI, data)
	if err != nil {
		DealResponse(w, response)
		return
	}
	resp, err := DeserializationResp(ctx, response, nil)
	if err != nil {
		return
	}
	if resp.Code == 0 {
		jwtserver.DestroyToken(ctx, o.s, o.redisClient, data.ID)
	}
	DealResponse(w, response)
	return
}

// UpdateListUserStatusRequest update list user status request
type UpdateListUserStatusRequest struct {
	IDS       []string `json:"ids" binding:"required"`
	UseStatus int      `json:"useStatus" binding:"required"` //状态：1正常，-2禁用，2激活==1 （与账号库相同）批量不支持删除，删除需要谨慎操作
	UpdatedBy string   `json:"updatedBy"`
	TenantID  string   `json:"tenantID"`
}

// UpdateListUserStatusResponse update list user status response
type UpdateListUserStatusResponse struct {
}

// UpdateUsersStatus update users status
func (o *org) UpdateUsersStatus(ctx context.Context, r *http.Request, w http.ResponseWriter, data *UpdateListUserStatusRequest) {
	response, err := DealRequest(o.client, o.conf.OrgAPIs.Host, r, o.conf.OrgAPIs.UpdateUsersStatusURI, data)
	if err != nil {
		DealResponse(w, response)
		return
	}
	resp, err := DeserializationResp(ctx, response, nil)
	if err != nil {
		return
	}
	if resp.Code == 0 {
		jwtserver.DestroyToken(ctx, o.s, o.redisClient, data.IDS...)
	}
	DealResponse(w, response)
	return
}

// AdminResetPasswordRequest admin reset user password request
type AdminResetPasswordRequest struct {
	UserIDs     []string      `json:"userIDs"`
	CreatedBy   string        `json:"createdBy"`
	TenantID    string        `json:"tenantID"`
	SendMessage []SendMessage `json:"sendMessage"`
}

// SendMessage send message
type SendMessage struct {
	UserID      string `json:"userID"`
	SendChannel int    `json:"sendChannel"`
	SendTo      string `json:"sendTo"`
}

// AdminUpdatePasswordResponse 管理员重制密码
type AdminUpdatePasswordResponse struct {
	Users []ResetPasswordResponse `json:"Users"`
}

// ResetPasswordResponse reset password response
type ResetPasswordResponse struct {
	UserID   string `json:"userID"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
}

// AdminResetPassword admin reset password
func (o *org) AdminResetPassword(ctx context.Context, r *http.Request, w http.ResponseWriter, data *AdminResetPasswordRequest) {
	response, err := DealRequest(o.client, o.conf.OrgAPIs.Host, r, o.conf.OrgAPIs.AdminResetPasswordURI, data)
	if err != nil {
		DealResponse(w, response)
		return
	}
	resp, err := DeserializationResp(ctx, response, nil)
	if err != nil {
		return
	}
	if resp.Code == 0 {
		jwtserver.DestroyToken(ctx, o.s, o.redisClient, data.UserIDs...)
	}
	DealResponse(w, response)
	return
}

// UserResetPasswordRequest user reset self password request
type UserResetPasswordRequest struct {
	UserID      string `json:"userID"`
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
	UseStatus   int    `json:"useStatus"` //状态：1正常，-2禁用，-1删除 （与账号库相同）
	TenantID    string `json:"tenantID"`
}

// UserResetPasswordResponse user reset self password response
type UserResetPasswordResponse struct {
	UserID string `json:"userID"`
}

// UserResetPassword  user reset self password
func (o *org) UserResetPassword(ctx context.Context, r *http.Request, w http.ResponseWriter, data *UserResetPasswordRequest) {
	response, err := DealRequest(o.client, o.conf.OrgAPIs.Host, r, o.conf.OrgAPIs.UserResetPasswordURI, data)
	if err != nil {
		logger.Logger.Error(err)
		DealResponse(w, response)
		return
	}
	resp, err := DeserializationResp(ctx, response, nil)
	if err != nil {
		logger.Logger.Error(err)
		logger.Logger.Error(response)
		return
	}
	if resp.Code == 0 {
		jwtserver.DestroyToken(ctx, o.s, o.redisClient, data.UserID)
	}
	DealResponse(w, response)
	return
}

// UserForgetResetRequest user forget reset password request
type UserForgetResetRequest struct {
	UserName    string `json:"userName" binding:"required,max=60"` //多形态:邮箱、手机、其它
	NewPassword string `json:"newPassword" binding:"required"`
	Code        string `json:"code" binding:"required"`
	TenantID    string `json:"tenantID"`
}

// UserForgetResetResponse user forget reset password response
type UserForgetResetResponse struct {
	UserID string `json:"userID"`
}

// UserForgetResetPassword user forget reset password
func (o *org) UserForgetResetPassword(ctx context.Context, r *http.Request, w http.ResponseWriter, data *UserForgetResetRequest) {
	response, err := DealRequest(o.client, o.conf.OrgAPIs.Host, r, o.conf.OrgAPIs.UserForgetResetPasswordURI, data)
	if err != nil {
		return
	}
	res := &UserForgetResetResponse{}
	resp, err := DeserializationResp(ctx, response, res)
	if err != nil {
		DealResponse(w, response)
		return
	}
	if resp.Code == 0 {
		jwtserver.DestroyToken(ctx, o.s, o.redisClient, res.UserID)
	}
	DealResponse(w, response)
	return
}

//DealRequest deal request
func DealRequest(c http.Client, host string, r *http.Request, path string, data interface{}) (*http.Response, error) {
	request := r.Clone(r.Context())
	parse, err := url.ParseRequestURI(host)
	if err != nil {
		return nil, error2.New(error2.Internal)
	}
	request.URL = parse
	request.Host = parse.Host
	request.URL.Path = path
	request.RequestURI = ""
	request.URL.RawQuery = r.URL.RawQuery
	if r.Method != "GET" && data != nil {
		marshal, _ := json.Marshal(data)
		l := len(marshal)
		itoa := strconv.Itoa(l)
		request.Header.Set("Content-Length", itoa)
		request.ContentLength = int64(l)
		request.Body = io.NopCloser(bytes.NewReader(marshal))
	}

	return c.Do(request)
}

//DealResponse deal response
func DealResponse(w http.ResponseWriter, response *http.Response) {

	defer response.Body.Close()
	all, err := io.ReadAll(response.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	header := response.Header.Clone()
	for k := range header {
		for k1 := range header[k] {
			w.Header().Add(k, header[k][k1])
		}
	}
	w.WriteHeader(response.StatusCode)
	w.Write(all)
	return
}

type respData interface{}

// R response data
type R struct {
	err  error
	Code int64    `json:"code"`
	Msg  string   `json:"msg,omitempty"`
	Data respData `json:"data"`
}

//DeserializationResp marshal response body to struct
func DeserializationResp(ctx context.Context, response *http.Response, entity interface{}) (*R, error) {
	if response.StatusCode != http.StatusOK {
		return nil, error2.New(error2.Internal)
	}
	r := new(R)
	r.Data = entity
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, r)
	if err != nil {
		return nil, err
	}
	response.Body = io.NopCloser(bytes.NewReader(body))
	response.ContentLength = int64(len(body))
	return r, nil
}
