package restful

import (
	"github.com/gin-gonic/gin"
	error2 "github.com/quanxiang-cloud/cabin/error"
	ginheader "github.com/quanxiang-cloud/cabin/tailormade/header"
	"github.com/quanxiang-cloud/cabin/tailormade/resp"
	"github.com/quanxiang-cloud/warden/internal/org"
	"github.com/quanxiang-cloud/warden/pkg/code"
	"github.com/quanxiang-cloud/warden/pkg/configs"
)

// Org org
type Org struct {
	orgs org.Org
}

// NewOrg new
func NewOrg(conf configs.Config, redisClient redis.UniversalClient) (*Org, error) {
	return &Org{
		orgs: org.NewOrg(conf, redisClient),
	}, nil
}

//UpdateUserStatus 单个修改用户状态
func (o *Org) UpdateUserStatus(c *gin.Context) {
	r := &org.UpdateUserStatusRequest{}
	err := c.ShouldBind(r)
	if err != nil {
		resp.Format(nil, error2.New(code.InvalidParams)).Context(c)
		return
	}
	o.orgs.UpdateUserStatus(ginheader.MutateContext(c), c.Request, c.Writer, r)
	return
}

//UpdateListUserStatus 批量修改用户状态
func (o *Org) UpdateListUserStatus(c *gin.Context) {
	r := &org.UpdateListUserStatusRequest{}
	err := c.ShouldBind(r)
	if err != nil {
		resp.Format(nil, error2.New(code.InvalidParams)).Context(c)
		return
	}
	o.orgs.UpdateUsersStatus(ginheader.MutateContext(c), c.Request, c.Writer, r)
	return
}

//AdminResetPassword 管理员重制密码
func (o *Org) AdminResetPassword(c *gin.Context) {
	r := &org.AdminResetPasswordRequest{}
	err := c.ShouldBind(r)
	if err != nil {
		resp.Format(nil, error2.New(code.InvalidParams)).Context(c)
		return
	}
	o.orgs.AdminResetPassword(ginheader.MutateContext(c), c.Request, c.Writer, r)
	return
}

//UserResetPassword 用户重制密码
func (o *Org) UserResetPassword(c *gin.Context) {
	r := &org.UserResetPasswordRequest{}
	err := c.ShouldBind(r)
	if err != nil {
		resp.Format(nil, error2.New(code.InvalidParams)).Context(c)
		return
	}
	userID := c.GetHeader("User-Id")
	r.UserID = userID
	o.orgs.UserResetPassword(ginheader.MutateContext(c), c.Request, c.Writer, r)
	return
}

//UserForgetResetPassword 忘记密码用户重制
func (o *Org) UserForgetResetPassword(c *gin.Context) {
	r := &org.UserForgetResetRequest{}
	err := c.ShouldBind(r)
	if err != nil {
		resp.Format(nil, error2.New(code.InvalidParams)).Context(c)
		return
	}
	o.orgs.UserForgetResetPassword(ginheader.MutateContext(c), c.Request, c.Writer, r)
	return
}
