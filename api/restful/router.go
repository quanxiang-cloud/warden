package restful

import (
	"context"
	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/warden/pkg/configs"
	"github.com/quanxiang-cloud/warden/pkg/probe"
	"github.com/quanxiang-cloud/warden/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/quanxiang-cloud/cabin/tailormade/db/redis"
	ginlogger "github.com/quanxiang-cloud/cabin/tailormade/gin"
)

const (
	// DebugMode indicates mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates mode is release.
	ReleaseMode = "release"
)

// Router 路由
type Router struct {
	c *configs.Config

	engine *gin.Engine
}

// NewRouter 开启路由
func NewRouter(ctx context.Context, c *configs.Config, log logger.AdaptedLogger) (*Router, error) {
	engine, err := newRouter(c)
	if err != nil {
		return nil, err
	}
	redisClient, err := redis.NewClient(c.Redis)
	if err != nil {
		panic(err)
	}
	jwtAPI, err := NewJWTApi(*c, redisClient, log)
	if err != nil {
		return nil, err
	}
	newOrg, err := NewOrg(*c, redisClient)
	k := engine.Group("/api/v1/warden")
	{
		k.Any("/login", jwtAPI.LoginHandler)   //ok
		k.Any("/logout", jwtAPI.LogOutHandler) //ok
		k.Any("/refresh", jwtAPI.Refresh)      //ok

		k.Any("/auth", jwtAPI.Auth)
		k.Any("/destroy", jwtAPI.DestroyByUserID)
		k.Any("/check", jwtAPI.CheckToken)           //ok
		k.Any("/switch/tenant", jwtAPI.SwitchTenant) //ok

		k.POST("/org/m/user/update/status", newOrg.UpdateUserStatus)          //ok
		k.POST("/org/m/user/updates/status", newOrg.UpdateListUserStatus)     //ok
		k.POST("/org/m/account/reset/password", newOrg.AdminResetPassword)    //
		k.POST("/org/h/account/reset/password", newOrg.UserResetPassword)     //
		k.POST("/org/h/account/forget/reset", newOrg.UserForgetResetPassword) //

	}
	engine.Any("/authCoder", jwtAPI.AuthCoder)
	{
		probe := probe.New(util.LoggerFromContext(ctx))
		engine.GET("liveness", func(c *gin.Context) {
			probe.LivenessProbe(c.Writer, c.Request)
		})

		engine.Any("readiness", func(c *gin.Context) {
			probe.ReadinessProbe(c.Writer, c.Request)
		})

	}
	return &Router{
		c:      c,
		engine: engine,
	}, nil
}

func newRouter(c *configs.Config) (*gin.Engine, error) {
	if c.Model == "" || (c.Model != ReleaseMode && c.Model != DebugMode) {
		c.Model = ReleaseMode
	}
	gin.SetMode(c.Model)
	engine := gin.New()
	engine.Use(ginlogger.LoggerFunc(), ginlogger.RecoveryFunc())
	return engine, nil
}

// Run 启动服务
func (r *Router) Run() {
	r.engine.Run(r.c.Port)
}

// Close 关闭服务
func (r *Router) Close() {
}
