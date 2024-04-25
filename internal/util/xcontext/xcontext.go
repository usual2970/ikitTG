package xcontext

import (
	"context"
	"errors"

	"github.com/labstack/echo/v5"
)

type AppKey struct{}

func SetAppKey(ctx context.Context, appKey string) context.Context {
	ctx = context.WithValue(ctx, AppKey{}, appKey)
	return ctx
}

func GetAppKey(ctx context.Context) (string, error) {
	val := ctx.Value(AppKey{})
	if val == nil {
		return "", errors.New("app key not found")
	}

	return val.(string), nil
}

type OpenapiContextKey struct{}

type OpenapiContext struct {
	Path     string                 `json:"path"`     // 路径
	Config   map[string]string      `json:"config"`   // 配置
	Param    map[string]interface{} `json:"param"`    // 参数
	Method   string                 `json:"method"`   // 方法
	Domain   string                 `json:"domain"`   // 网关
	Header   map[string]string      `json:"header"`   // 请求头
	Response interface{}            `json:"response"` // 响应
	Error    error                  `json:"error"`    // 错误
	Async    bool                   `json:"async"`    // 是否异步
	AsyncUrl string                 `json:"asyncUrl"` // 异步回调地址
}

func SetOpenapiContext(ctx context.Context, val *OpenapiContext) context.Context {
	return context.WithValue(ctx, OpenapiContextKey{}, val)
}

func GetOpenapiContext(ctx context.Context) (*OpenapiContext, error) {
	val := ctx.Value(OpenapiContextKey{})

	if val == nil {
		return nil, errors.New("openapi context is nil")
	}

	rs, ok := val.(*OpenapiContext)
	if !ok {
		return nil, errors.New("openapi context is not a *OpenapiContext")
	}

	return rs, nil
}

type NoticeContextKey struct{}

type NoticeContext struct {
	Url          string      `json:"url"`    // 路径
	Param        interface{} `json:"param"`  // 参数
	Method       string      `json:"method"` // 方法
	Echo         echo.Context
	Header       map[string]string `json:"header"`        // 请求头
	Response     interface{}       `json:"response"`      // 响应
	ResponseType string            `json:"response_type"` // 响应类型
	Error        error             `json:"error"`         // 错误
}

func SetNoticeContext(ctx context.Context, val *NoticeContext) context.Context {
	return context.WithValue(ctx, NoticeContextKey{}, val)
}

func GetNoticeContext(ctx context.Context) (*NoticeContext, error) {
	val := ctx.Value(NoticeContextKey{})

	if val == nil {
		return nil, errors.New("notice context is nil")
	}

	rs, ok := val.(*NoticeContext)
	if !ok {
		return nil, errors.New("notice context is not a *NoticeContext")
	}

	return rs, nil
}

type EchoContextKey struct{}

func SetEchoContext(ctx context.Context, e echo.Context) context.Context {
	return context.WithValue(ctx, NoticeContextKey{}, e)
}

func GetEchoContext(ctx context.Context) (echo.Context, error) {
	val := ctx.Value(EchoContextKey{})

	if val == nil {
		return nil, errors.New("echo context is nil")
	}

	rs, ok := val.(echo.Context)
	if !ok {
		return nil, errors.New("echo context is not a *=echo.Context")
	}

	return rs, nil
}
