package code

import error2 "github.com/quanxiang-cloud/cabin/error"

func init() {
	error2.CodeTable = codeTable
}

const (
	// InvalidURI 无效的URI
	InvalidURI = 20014000000
	// InvalidParams 无效的参数
	InvalidParams = 200140000001
	// InvalidTimestamp 无效的时间格式
	InvalidTimestamp = 200140000002

	// ErrInvalidAccessToken 无效的token
	ErrInvalidAccessToken = 200140000003

	// ErrInvalidRefreshToken 无效的refresh_token
	ErrInvalidRefreshToken = 20014000004
	// ErrExpiredAccessToken ErrExpiredAccessToken
	ErrExpiredAccessToken = 20014000005
	// ErrExpiredRefreshToken ErrExpiredRefreshToken
	ErrExpiredRefreshToken = 20014000006
)

// codeTable 码表
var codeTable = map[int64]string{
	InvalidURI:             "无效的URI.",
	InvalidParams:          "无效的参数.",
	InvalidTimestamp:       "无效的时间格式.",
	ErrInvalidAccessToken:  "无效的token.",
	ErrInvalidRefreshToken: "无效的刷新token.",
	ErrExpiredAccessToken:  "token已经失效.",
	ErrExpiredRefreshToken: "刷新token已经失效.",
}
