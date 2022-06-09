package org

import (
	"context"
	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"net/http"
)

const (
	host = "http://org/api/v1/org"

	othAddUsersURI  = "/o/user/add"
	othAddDepsURI   = "/o/department/add"
	oneUserURI      = "/o/user/info"
	usersByIDsURI   = "/o/user/ids"
	depByIDsURI     = "/o/dep/ids"
	usersByDepIDURI = "/o/user/dep/id"
	depMaxGradeURI  = "/o/dep/max/grade"
)

// User 人员组织服务提供
type User interface {
	OthAddUsers(ctx context.Context, r *AddUsersRequest) (*AddListResponse, error)
	OthAddDeps(ctx context.Context, r *AddDepartmentRequest) (*AddListResponse, error)
	GetUserInfo(ctx context.Context, r *OneUserRequest) (*OneUserResponse, error)
	GetUserByIDs(ctx context.Context, r *GetUserByIDsRequest) (*GetUserByIDsResponse, error)
	GetDepByIDs(ctx context.Context, r *GetDepByIDsRequest) (*GetDepByIDsResponse, error)
	GetUsersByDepID(ctx context.Context, r *GetUsersByDepIDRequest) (*GetUsersByDepIDResponse, error)
	GetDepMaxGrade(ctx context.Context, r *GetDepMaxGradeRequest) (*GetDepMaxGradeResponse, error)
}
type user struct {
	client http.Client
}

// NewUser 初始化对象
func NewUser(conf client.Config) User {
	return &user{
		client: client.New(conf),
	}
}

//AddUsersRequest other server add user request
type AddUsersRequest struct {
	Users      []AddUser `json:"users"`
	IsUpdate   int       `json:"isUpdate"`   //是否更新已有数据，1更新，-1不更新
	SyncID     string    `json:"syncID"`     //同步中心id
	SyncSource string    `json:"syncSource"` //同步来源
}

//AddUser other server add user to org
type AddUser struct {
	ID        string   `json:"id,omitempty"`
	Name      string   `json:"name,omitempty"`
	Phone     string   `json:"phone,omitempty"`
	Email     string   `json:"email,omitempty"`
	AccountID string   `json:"-"`
	SelfEmail string   `json:"selfEmail,omitempty"`
	IDCard    string   `json:"idCard,omitempty"`
	Address   string   `json:"address,omitempty"`
	UseStatus int      `json:"useStatus,omitempty"` //状态：1正常，-2禁用，-1删除，2激活==1 ，-3离职（与账号库相同）
	Gender    int      `json:"gender,omitempty"`    //状态：0无，1男，2女
	CompanyID string   `json:"companyID,omitempty"` //所属公司id
	Position  string   `json:"position,omitempty"`  //职位
	Avatar    string   `json:"avatar,omitempty"`    //头像
	Remark    string   `json:"remark,omitempty"`    //备注
	JobNumber string   `json:"jobNumber,omitempty"` //工号
	DepIDs    []string `json:"depIDs,omitempty"`
	EntryTime int64    `json:"entryTime,omitempty" ` //入职时间
	Source    string   `json:"source,omitempty" `    //信息来源
	SourceID  string   `json:"sourceID,omitempty" `  //信息来源的ID,用于返回给服务做业务
}

//AddListResponse other server add user or dep to org response
type AddListResponse struct {
	Result map[int]*Result `json:"result"`
}

//Result list add response
type Result struct {
	ID     string `json:"id"`
	Remark string `json:"remark"`
	Attr   int    `json:"attr"` //11 add ok,0fail,12, update ok
}

//OthAddUsers 实际请求
func (u *user) OthAddUsers(ctx context.Context, r *AddUsersRequest) (*AddListResponse, error) {
	response := &AddListResponse{}
	err := client.POST(ctx, &u.client, host+othAddUsersURI, r, response)
	if err != nil {
		return nil, err
	}
	return response, err
}

//AddDepartmentRequest other server add  department to org request
type AddDepartmentRequest struct {
	Deps       []AddDep `json:"deps"`
	SyncDep    int      `json:"syncDep"`    //是否同步部门1同步，-1不同步
	IsUpdate   int      `json:"isUpdate"`   //是否更新已有数据，1更新，-1不更新
	SyncID     string   `json:"syncID"`     //同步中心id
	SyncSource string   `json:"syncSource"` //同步来源
}

//AddDep other server add department to org
type AddDep struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	UseStatus int    `json:"useStatus"` //1正常，-1真删除，-2禁用
	Attr      int    `json:"attr"`      //1公司，2部门
	PID       string `json:"pid"`       //上层ID
	SuperPID  string `json:"superID"`   //最顶层父级ID
	CompanyID string `json:"companyID"` //所属公司id
	Grade     int    `json:"grade"`     //部门等级
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	CreatedBy string `json:"createdBy"`        //创建者
	UpdatedBy string `json:"updatedBy"`        //创建者
	Remark    string `json:"remark,omitempty"` //备注
}

//OthAddDeps 实际请求
func (u *user) OthAddDeps(ctx context.Context, r *AddDepartmentRequest) (*AddListResponse, error) {
	response := &AddListResponse{}
	err := client.POST(ctx, &u.client, host+othAddDepsURI, r, response)
	if err != nil {
		return nil, err
	}
	return response, err
}

// OneUserRequest 查询一个
type OneUserRequest struct {
	ID string `json:"id" form:"id"  binding:"required,max=64"`
}

// OneUserResponse 查询一个
type OneUserResponse struct {
	ID        string              `json:"id,omitempty" `
	Name      string              `json:"name,omitempty" `
	Phone     string              `json:"phone,omitempty" `
	Email     string              `json:"email,omitempty" `
	SelfEmail string              `json:"selfEmail,omitempty" `
	UseStatus int                 `json:"useStatus,omitempty" ` //状态：1正常，-2禁用，-3离职，-1删除，2激活==1 （与账号库相同）
	TenantID  string              `json:"tenantID,omitempty" `  //租户id
	Position  string              `json:"position,omitempty" `  //职位
	Avatar    string              `json:"avatar,omitempty" `    //头像
	JobNumber string              `json:"jobNumber,omitempty" ` //工号
	Status    int                 `json:"status"`               //第一位：密码是否需要重置
	Dep       [][]DepOneResponse  `json:"deps,omitempty"`       //用户所在部门
	Leader    [][]OneUserResponse `json:"leaders,omitempty"`    //用户所在部门
}

// DepOneResponse 用于用户部门层级线索
type DepOneResponse struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	LeaderID  string `json:"leaderID"`
	UseStatus int    `json:"useStatus,omitempty"`
	PID       string `json:"pid"`               //上层ID
	SuperPID  string `json:"superID,omitempty"` //最顶层父级ID
	Grade     int    `json:"grade,omitempty"`   //部门等级
	Attr      int    `json:"attr"`              //1公司，2部门
}

//GetUserInfo 实际请求
func (u *user) GetUserInfo(ctx context.Context, r *OneUserRequest) (*OneUserResponse, error) {
	response := &OneUserResponse{}
	err := client.POST(ctx, &u.client, host+oneUserURI, r, response)
	if err != nil {
		return nil, err
	}
	return response, err
}

//GetUserByIDsRequest get user by ids request
type GetUserByIDsRequest struct {
	IDs []string `json:"ids"`
}

// GetUserByIDsResponse get user by ids response
type GetUserByIDsResponse struct {
	Users []OneUserResponse `json:"users"`
}

//GetUserByIDs 实际请求
func (u *user) GetUserByIDs(ctx context.Context, r *GetUserByIDsRequest) (*GetUserByIDsResponse, error) {
	response := &GetUserByIDsResponse{}
	err := client.POST(ctx, &u.client, host+usersByIDsURI, r, response)
	if err != nil {
		return nil, err
	}
	return response, err
}

// GetDepByIDsRequest 批量查询
type GetDepByIDsRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// GetDepByIDsResponse 批量查询
type GetDepByIDsResponse struct {
	Deps []DepOneResponse `json:"deps"`
}

//GetDepByIDs 实际请求
func (u *user) GetDepByIDs(ctx context.Context, r *GetDepByIDsRequest) (*GetDepByIDsResponse, error) {
	response := &GetDepByIDsResponse{}
	err := client.POST(ctx, &u.client, host+depByIDsURI, r, response)
	if err != nil {
		return nil, err
	}
	return response, err
}

// GetUsersByDepIDRequest get users by id request
type GetUsersByDepIDRequest struct {
	DepID          string `json:"depID"`
	IsIncludeChild int    `json:"isIncludeChild"` //1包含
}

// GetUsersByDepIDResponse get users by id response
type GetUsersByDepIDResponse struct {
	Users []OneUserResponse `json:"users"`
}

// GetUsersByDepID 实际请求
func (u *user) GetUsersByDepID(ctx context.Context, r *GetUsersByDepIDRequest) (*GetUsersByDepIDResponse, error) {
	response := &GetUsersByDepIDResponse{}
	err := client.POST(ctx, &u.client, host+usersByDepIDURI, r, response)
	if err != nil {
		return nil, err
	}
	return response, err
}

// GetDepMaxGradeRequest request
type GetDepMaxGradeRequest struct {
}

// GetDepMaxGradeResponse response
type GetDepMaxGradeResponse struct {
	Grade int64 `json:"grade"`
}

//GetDepMaxGrade 实际请求
func (u *user) GetDepMaxGrade(ctx context.Context, r *GetDepMaxGradeRequest) (*GetDepMaxGradeResponse, error) {
	response := &GetDepMaxGradeResponse{}
	err := client.POST(ctx, &u.client, host+depMaxGradeURI, r, response)
	if err != nil {
		return nil, err
	}
	return response, err
}
