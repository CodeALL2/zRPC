package model

import "errors"

var (
	ErrClientNotLine    = errors.New("服务端通信失败")
	ErrTimeout          = errors.New("操作超时")
	ErrPermissionDenied = errors.New("权限不足")
)
