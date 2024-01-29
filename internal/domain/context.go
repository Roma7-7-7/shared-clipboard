package domain

import (
	"context"
)

const RuntimeTraceID = "runtime"

type (
	Authority struct {
		UserID   uint64
		UserName string
	}

	authorityContextKey struct{}
	traceIDCtxKey       struct{}
)

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, &traceIDCtxKey{}, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(&traceIDCtxKey{}).(string); ok {
		return traceID
	}
	return "undefined"
}

func AuthorityFromContext(ctx context.Context) (*Authority, bool) {
	token, ok := ctx.Value(authorityContextKey{}).(*Authority)
	return token, ok
}

func ContextWithAuthority(ctx context.Context, authority *Authority) context.Context {
	return context.WithValue(ctx, authorityContextKey{}, authority)
}
