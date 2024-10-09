package lib

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type Trace struct {
	TraceId     string
	SpanId      string
	Caller      string
	SrcMethod   string
	HintCode    int64
	HintContent string
}

type TraceContext struct {
	Trace
	CSpanId string
}

// SetGinTraceContext 用于在 Gin 的请求上下文 c 中设置追踪信息。它将 trace 存储在 Gin 的上下文中，供后续处理中使用。
func SetGinTraceContext(c *gin.Context, trace *TraceContext) error {
	if trace == nil || c == nil {
		return errors.New("context is nil")
	}
	c.Set("trace", trace)
	return nil
}

// SetTraceContext 用于在普通的 context.Context 中设置追踪信息，使用 context.WithValue 将 trace 存储在上下文中，返回新的上下文
func SetTraceContext(ctx context.Context, trace *TraceContext) context.Context {
	if trace == nil {
		return ctx
	}
	return context.WithValue(ctx, "trace", trace)
}

// GetTraceContext 用于从 Gin 或普通的 context.Context 中提取追踪信息。如果在上下文中找不到追踪信息，返回一个新的空追踪信息对象（NewTrace()）
func GetTraceContext(ctx context.Context) *TraceContext {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		traceIntraceContext, exists := ginCtx.Get("trace")
		if !exists {
			return NewTrace()
		}
		traceContext, ok := traceIntraceContext.(*TraceContext)
		if ok {
			return traceContext
		}
		return NewTrace()
	}

	if contextInterface, ok := ctx.(context.Context); ok {
		traceContext, ok := contextInterface.Value("trace").(*TraceContext)
		if ok {
			return traceContext
		}
		return NewTrace()

	}
	return NewTrace()
}
