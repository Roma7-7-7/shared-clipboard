package context

import (
	"context"

	"github.com/Roma7-7-7/shared-clipboard/tools"
)

type (
	Authority struct {
		UserID   uint64
		UserName string
	}

	authorityContextKey struct{}
	traceIDCtxKey       struct{}
)

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, &traceIDCtxKey{}, traceID)
}

func TraceIDFrom(ctx context.Context) string {
	if traceID, ok := ctx.Value(&traceIDCtxKey{}).(string); ok {
		return traceID
	}
	return "undefined#" + tools.RandomAlphanumericKey(8)
}

func AuthorityFrom(ctx context.Context) (*Authority, bool) {
	token, ok := ctx.Value(authorityContextKey{}).(*Authority)
	return token, ok
}

func WithAuthority(ctx context.Context, authority *Authority) context.Context {
	return context.WithValue(ctx, authorityContextKey{}, authority)
}
