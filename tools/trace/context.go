package trace

import "context"

const RuntimeTraceID = "runtime"

type traceIDCtxKey struct{}

func WithID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, &traceIDCtxKey{}, traceID)
}

func ID(ctx context.Context) string {
	if traceID, ok := ctx.Value(&traceIDCtxKey{}).(string); ok {
		return traceID
	}
	return "undefined"
}
