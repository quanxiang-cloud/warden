package jwtserver

import (
	"context"
	"encoding/json"
	"errors"
	error2 "github.com/quanxiang-cloud/cabin/error"
	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/client"

	"github.com/quanxiang-cloud/warden/pkg/code"
	"github.com/quanxiang-cloud/warden/pkg/configs"
	"github.com/quanxiang-cloud/warden/pkg/jwts"
	"github.com/quanxiang-cloud/warden/pkg/jwts/generates"
	"github.com/quanxiang-cloud/warden/pkg/jwts/manage"
	"github.com/quanxiang-cloud/warden/pkg/jwts/server"
	"github.com/quanxiang-cloud/warden/pkg/jwts/store"
	"github.com/quanxiang-cloud/warden/pkg/org"

	"net/http"
	"strings"
	"time"
)

// JWTServer interface
type JWTServer interface {
	Login(ctx context.Context, r *LoginRequst) (*LoginResponse, error)
	Logout(ctx context.Context, tokenString string) (string, error)
	Refresh(ctx context.Context, refreshToken string) (interface{}, error)
	DestroyByUserID(ctx context.Context, req *DestroyTokenRequest) (*DestroyTokenResponse, error)
	CheckToken(c context.Context, header http.Header, token string) (response *CheckTokenResponse, err error)
	Auth(c context.Context, header http.Header, token string) (interface{}, error)
	FaasCheck(c context.Context, req *FaasCheckReq) (*FaasCheckResp, error)
	SwitchTenant(c context.Context, req *SwitchTenantRequest) (*SwitchTenantResponse, error)
}

//jwtServer 登录实现结构体
type jwtServer struct {
	s      *server.Server
	client http.Client
	org    org.User
	redisc redis.UniversalClient
	conf   configs.Config
}

//LoginRequst LoginRequst
type LoginRequst struct {
	UserName  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	LoginType string `json:"login_type" binding:"required"`
}

//LoginResponse response
type LoginResponse struct {
	Token map[string]interface{}
}

//OrgCheckRequest 用于调用org服务
type OrgCheckRequest struct {
	UserName string `json:"username"` //多形态:邮箱、手机、其它
	Password string `json:"password"`
	Types    string `json:"types"` //登录模式
}

// OrgCheckResponse 返回用户信息
type OrgCheckResponse struct {
	UserID    string `json:"userID"`
	UseStatus int    `json:"useStatus"` //状态：1正常，-2禁用，-1删除 （与账号库相同）
	Code      int64  `json:"code"`      //错误码
	Msg       string `json:"msg"`       //错误信息
}

// Login Login
func (j *jwtServer) Login(ctx context.Context, r *LoginRequst) (*LoginResponse, error) {
	loginReq := OrgCheckRequest{
		UserName: r.UserName,
		Password: r.Password,
		Types:    r.LoginType,
	}
	userAccount := OrgCheckResponse{}
	err := client.POST(ctx, &j.client, j.conf.OrgAPIs.Host+j.conf.OrgAPIs.LoginURI, loginReq, &userAccount)
	if err != nil {
		return nil, err
	}
	if userAccount.UserID == "" {
		return nil, error2.NewErrorWithString(userAccount.Code, userAccount.Msg)
	}

	token, err := j.s.HandleTokenRequest(ctx, userAccount.UserID, nil)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
	j.redisc.Del(ctx, wardenUserTenantCache+userAccount.UserID)
	return &LoginResponse{
		Token: token,
	}, nil
}

// Logout LoginOut
func (j *jwtServer) Logout(c context.Context, tokenString string) (string, error) {
	tokenInfo, _ := j.s.Manager.LoadAccessToken(c, tokenString)
	if tokenInfo == nil {
		return "", errors.New("invalid accessToken")
	}
	basicID := j.redisc.Get(c, store.JWTRedis+tokenInfo.GetAccess()).Val()
	err := j.s.Manager.RemoveAccessToken(c, tokenInfo.GetAccess())
	err = j.s.Manager.RemoveRefreshToken(c, tokenInfo.GetRefresh())
	j.redisc.Del(c, store.JWTRedis+basicID)
	j.redisc.HDel(c, store.JWTRedisUsers+tokenInfo.GetUserID(), basicID)
	if err != nil {
		return "", err
	}
	return "", nil
}

// Refresh Refresh
func (j *jwtServer) Refresh(ctx context.Context, refreshToken string) (interface{}, error) {
	token, err := j.s.HandleRefreshTokenRequest(ctx, refreshToken)
	if err != nil {
		logger.Logger.Error(err)
		return nil, error2.New(code.ErrInvalidRefreshToken)
	}
	return token, nil
}

// DestroyTokenRequest 接收用户id数组
type DestroyTokenRequest struct {
	UsersID []string `json:"usersID"`
}

// DestroyTokenResponse 接收用户id数组
type DestroyTokenResponse struct {
}

// DestroyByUserID DestroyByUserID
func (j *jwtServer) DestroyByUserID(ctx context.Context, req *DestroyTokenRequest) (*DestroyTokenResponse, error) {
	DestroyToken(ctx, j.s, j.redisc, req.UsersID...)
	return nil, nil
}

// CheckTokenResponse check token response
type CheckTokenResponse struct {
	UserID   string
	Name     string
	DepID    string
	TenantID string
}

// CheckToken CheckToken
func (j *jwtServer) CheckToken(c context.Context, header http.Header, accesstoken string) (response *CheckTokenResponse, err error) {
	var tokenInfo jwts.TokenInfo

	tokenInfo, err = j.s.ValidationBearerToken(c, "", accesstoken)
	if err != nil {
		return nil, error2.New(code.ErrInvalidAccessToken)
	}
	info, depID, err := GetUserInfo(c, j.org, j.redisc, header, tokenInfo.GetUserID(), j.conf)
	if err != nil {
		return nil, error2.New(code.ErrInvalidAccessToken)
	}
	res := &CheckTokenResponse{
		UserID:   tokenInfo.GetUserID(),
		Name:     info.Name,
		DepID:    depID,
		TenantID: info.TenantID,
	}
	return res, nil
}

// Auth Auth
func (j *jwtServer) Auth(c context.Context, header http.Header, token string) (interface{}, error) {
	verifyToken, err := j.s.Manager.VerifyToken(c, token)
	if err != nil {
		return nil, error2.New(code.ErrInvalidAccessToken)
	}

	info, depID, err := GetUserInfo(c, j.org, j.redisc, header, verifyToken["jti"].(string), j.conf)
	if err != nil {
		return nil, error2.New(code.ErrInvalidAccessToken)
	}
	other := make(map[string]string)
	other["Department-Id"] = depID
	other["User-Name"] = info.Name

	ti, errData := j.s.HandleTokenRequest(c, info.ID, other)
	if errData != nil {
		logger.Logger.Error(errData)
		return nil, error2.New(code.ErrInvalidAccessToken)
	}
	return ti, nil
}

// FaasCheckReq FaasCheckReq
type FaasCheckReq struct {
	Token string
}

// FaasCheckResp FaasCheckResp
type FaasCheckResp struct {
	Code int
}

// FaasCheck FaasCheck
func (j *jwtServer) FaasCheck(c context.Context, req *FaasCheckReq) (*FaasCheckResp, error) {
	_, err := j.s.ValidationBearerToken(c, "", req.Token)
	if err != nil {
		logger.Logger.Info("Validation is fail ", err.Error())
		return nil, err
	}
	return &FaasCheckResp{
		Code: http.StatusOK,
	}, nil
}

//NewJWTImpl 初始化
func NewJWTImpl(conf configs.Config, redisClient redis.UniversalClient) (JWTServer, error) {

	return &jwtServer{
		s:      NewServer(),
		client: client.New(configs.GetConfig().InternalNet),
		org:    org.NewUser(configs.GetConfig().InternalNet),
		redisc: redisClient,
		conf:   conf,
	}, nil
}

//NewServer 初始化
func NewServer() *server.Server {
	manager := manage.NewDefaultManager()
	config := new(manage.Config)
	config.AccessTokenExp = time.Hour * configs.GetConfig().JWTConfig.AccessTokenExp
	config.RefreshTokenExp = time.Hour * configs.GetConfig().JWTConfig.RefreshTokenExp
	config.IsGenerateRefresh = true

	manager.MapTokenStorage(store.NewRedisClusterStore(&redis.ClusterOptions{
		Addrs:    configs.GetConfig().Redis.Addrs,
		Username: configs.GetConfig().Redis.Username,
		Password: configs.GetConfig().Redis.Password,
	}))

	// generate jwtServer access token
	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte(configs.GetConfig().JWTConfig.JwtKey), nil, jwt.SigningMethodHS256))

	return server.NewServer(server.NewConfig(), manager)
}

const wardenUserCache = "warden:orgs:user:"
const wardenUserTenantCache = "warden:orgs:user:tenant:"

// GetUserInfo get user info
func GetUserInfo(ctx context.Context, u org.User, redisClient redis.UniversalClient, header http.Header, userID string, conf configs.Config) (info *org.OneUserResponse, depID string, err error) {
	userData := redisClient.Get(ctx, wardenUserCache+userID).Val()
	tenantID := redisClient.Get(ctx, wardenUserTenantCache+userID).Val()
	user := &org.OneUserResponse{}
	if userData == "" {
		request := &org.OneUserRequest{
			ID: userID,
		}
		user, err = u.GetUserInfo(ctx, request)
		if err != nil {
			return nil, "", err
		}
		if tenantID != "" {
			user.TenantID = tenantID
		}
		marshal, _ := json.Marshal(user)
		redisClient.SetEX(ctx, wardenUserCache+userID, marshal, conf.OrgAPIs.Exp*time.Minute)
		redisClient.SetEX(ctx, wardenUserTenantCache+userID, user.TenantID, conf.JWTConfig.AccessTokenExp*time.Minute)

	} else {
		err := json.Unmarshal([]byte(userData), user)
		if err != nil {
			request := &org.OneUserRequest{
				ID: userID,
			}
			user, err = u.GetUserInfo(ctx, request)
			if err != nil {
				return nil, "", err
			}

			marshal, _ := json.Marshal(user)
			redisClient.SetEX(ctx, wardenUserCache+userID, marshal, conf.OrgAPIs.Exp*time.Minute)

		}
		if tenantID != "" {
			user.TenantID = tenantID
		}
		redisClient.SetEX(ctx, wardenUserTenantCache+userID, user.TenantID, conf.JWTConfig.AccessTokenExp*time.Minute)
	}
	depIDs := GetUserDEPIDs(user.Dep)
	for k := range depIDs {
		if k == 0 {
			depID = depID + strings.Join(depIDs[k], ",")
		} else {
			depID = depID + "|" + strings.Join(depIDs[k], ",")
		}
	}
	return user, depID, nil

}

//GetUserDEPIDs get org dep slice
func GetUserDEPIDs(deps [][]org.DepOneResponse) [][]string {
	if len(deps) > 0 {
		res := make([][]string, 0, len(deps))
		for k := range deps {
			depIDs := make([]string, 0, len(deps[k]))
			for k1 := range deps[k] {
				depIDs = append(depIDs, deps[k][k1].ID)
			}
			res = append(res, depIDs)
		}
		return res
	}
	return nil
}

// DestroyToken destroy token by userID
func DestroyToken(ctx context.Context, s *server.Server, redisClient redis.UniversalClient, userID ...string) {
	for k := range userID {
		_ = s.Manager.RemoveToken(ctx, userID[k])
		redisClient.Del(ctx, wardenUserCache+userID[k])
		redisClient.Del(ctx, wardenUserTenantCache+userID[k])
	}
}

// SwitchTenantRequest switch tenant request
type SwitchTenantRequest struct {
	UserID   string
	TenantID string `json:"tenantID"`
	Token    string
}

// SwitchTenantResponse  switch tenant response
type SwitchTenantResponse struct {
}

// SwitchTenant SwitchTenant
func (j *jwtServer) SwitchTenant(c context.Context, r *SwitchTenantRequest) (*SwitchTenantResponse, error) {
	//todo 这里要到租户服务验证人员和租户关系是否存在，存在就替换缓存
	token, _ := j.s.Manager.LoadAccessToken(c, r.Token)
	j.redisc.SetEX(c, wardenUserTenantCache+token.GetUserID(), r.TenantID, j.conf.JWTConfig.AccessTokenExp*time.Minute)
	return nil, nil
}
