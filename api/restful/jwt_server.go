package restful

import (
	error2 "github.com/quanxiang-cloud/cabin/error"
	"github.com/quanxiang-cloud/cabin/logger"
	ginheader "github.com/quanxiang-cloud/cabin/tailormade/header"
	"github.com/quanxiang-cloud/cabin/tailormade/resp"
	"github.com/quanxiang-cloud/warden/pkg/configs"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/quanxiang-cloud/warden/internal/jwtserver"
	"github.com/quanxiang-cloud/warden/pkg/code"

	"net/http"

	"github.com/gin-gonic/gin"
)

//accessToken and refresh token
const (
	AccessToken  = "Access-Token"
	RefreshToken = "Refresh-Token"
)
const (
	xToken = "X-Token"
	xAuth  = "X-Auth"
	pass   = "true"
	noPass = "false"
)

//header profile
const (
	_userID       = "User-Id"
	_userName     = "User-Name"
	_departmentID = "Department-Id"
	_tenantID     = "Tenant-Id"
)

// JWTApi JWTApi
type JWTApi struct {
	repo jwtserver.JWTServer
}

// NewJWTApi NewJWTApi
func NewJWTApi(conf configs.Config, redisClient redis.UniversalClient, log logger.AdaptedLogger) (*JWTApi, error) {
	jwtImpl, err := jwtserver.NewJWTImpl(conf, redisClient)
	if err != nil {
		return nil, err
	}
	return &JWTApi{
		repo: jwtImpl,
	}, nil
}

// LoginHandler LoginHandler
func (j *JWTApi) LoginHandler(c *gin.Context) {
	r := new(jwtserver.LoginRequst)
	err := c.ShouldBind(r)
	if err != nil {
		resp.Format(nil, error2.New(code.InvalidParams)).Context(c)
		return
	}
	res, err := j.repo.Login(ginheader.MutateContext(c), r)
	if err != nil {
		resp.Format(nil, err).Context(c)
		return
	}
	resp.Format(res.Token, nil).Context(c)

}

// LogOutHandler LoginHandler
func (j *JWTApi) LogOutHandler(c *gin.Context) {
	access := c.GetHeader(AccessToken)
	if access == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	_, err := j.repo.Logout(ginheader.MutateContext(c), access)
	resp.Format(nil, err).Context(c)
	return
}

// Refresh Refresh
func (j *JWTApi) Refresh(c *gin.Context) {
	refreshToken := c.GetHeader(RefreshToken)
	if refreshToken == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	res, err := j.repo.Refresh(ginheader.MutateContext(c), refreshToken)
	if err != nil {
		resp.Format(nil, err).Context(c)
		return
	}
	resp.Format(res, nil).Context(c)
}

// CheckToken CheckToken
func (j *JWTApi) CheckToken(c *gin.Context) {
	accessToken := c.GetHeader(AccessToken)
	if accessToken == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	res, err := j.repo.CheckToken(ginheader.MutateContext(c), c.Request.Header.Clone(), accessToken)

	if err == nil {
		c.Writer.Header().Set(_userID, res.UserID)
		c.Writer.Header().Set(_userName, res.Name)
		c.Writer.Header().Set(_departmentID, res.DepID)
		c.Writer.Header().Set(_tenantID, res.TenantID)
		return
	}
	c.AbortWithStatus(http.StatusUnauthorized)
	return

}

// Auth Auth
func (j *JWTApi) Auth(c *gin.Context) {
	token := c.GetHeader(AccessToken)
	if token == "" {
		resp.Format(nil, nil).Context(c, http.StatusUnauthorized)
		return
	}
	auth, err := j.repo.Auth(ginheader.MutateContext(c), c.Request.Header.Clone(), token)
	if err != nil {
		resp.Format(nil, err).Context(c)
		return
	}
	resp.Format(auth, nil).Context(c)
	return
}

// DestroyByUserID DestroyByUserID
func (j *JWTApi) DestroyByUserID(c *gin.Context) {
	r := &jwtserver.DestroyTokenRequest{}
	err := c.ShouldBind(r)
	if err != nil {
		resp.Format(nil, error2.New(code.InvalidParams)).Context(c)
		return
	}
	resp.Format(j.repo.DestroyByUserID(ginheader.MutateContext(c), r)).Context(c)
	return
}

// AuthCoder AuthCoder
func (j *JWTApi) AuthCoder(c *gin.Context) {
	token := c.Request.Header.Get(xToken)
	header := c.Writer.Header()
	if token == "" {
		header.Set(xAuth, noPass)
		return
	}
	req := &jwtserver.FaasCheckReq{
		Token: token,
	}
	resp, err := j.repo.FaasCheck(c, req)
	if err != nil {
		header.Set(xAuth, noPass)
		return
	}
	if resp.Code == http.StatusOK {
		header.Set(xAuth, pass)
		return
	}
}

// SwitchTenant SwitchTenant
func (j *JWTApi) SwitchTenant(c *gin.Context) {
	r := &jwtserver.SwitchTenantRequest{}
	err := c.ShouldBind(r)
	if err != nil {
		resp.Format(nil, error2.New(code.InvalidParams)).Context(c)
		return
	}
	accessToken := c.GetHeader(AccessToken)
	if accessToken == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	r.Token = accessToken
	resp.Format(j.repo.SwitchTenant(ginheader.MutateContext(c), r)).Context(c)

}

// IndexHandler IndexHandler
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.Host, ".")[0]

	url := &url.URL{
		Scheme: "http",
		Host:   host + ".coder.svc.cluster.local",
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, r)

}
