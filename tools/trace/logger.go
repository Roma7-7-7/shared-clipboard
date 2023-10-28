package trace

import (
	"context"

	"go.uber.org/zap"
)

const traceIDLogKey = "traceID"

type Logger interface {
	Debugw(ctx context.Context, msg string, keysAndValues ...interface{})
	Infow(ctx context.Context, msg string, keysAndValues ...interface{})
	Warnw(ctx context.Context, msg string, keysAndValues ...interface{})
	Errorw(ctx context.Context, msg string, keysAndValues ...interface{})
}

type SugaredLogger struct {
	log *zap.SugaredLogger
}

func NewSugaredLogger(logger *zap.SugaredLogger) *SugaredLogger {
	return &SugaredLogger{
		log: logger,
	}
}

func (l *SugaredLogger) Debugw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log.Debugw(msg, withTraceIDFromContext(ctx, keysAndValues)...)
}

func (l *SugaredLogger) Infow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log.Infow(msg, withTraceIDFromContext(ctx, keysAndValues)...)
}

func (l *SugaredLogger) Warnw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log.Warnw(msg, withTraceIDFromContext(ctx, keysAndValues)...)
}

func (l *SugaredLogger) Errorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log.Errorw(msg, withTraceIDFromContext(ctx, keysAndValues)...)
}

func withTraceIDFromContext(ctx context.Context, keysAndValues []interface{}) []interface{} {
	return withTraceID(ID(ctx), keysAndValues)
}

func withTraceID(traceID string, keysAndValues []interface{}) []interface{} {
	if hasRequestID(keysAndValues...) {
		return keysAndValues
	}
	return append(keysAndValues, traceIDLogKey, traceID)
}

func hasRequestID(keysAndValues ...interface{}) bool {
	for i, key := range keysAndValues {
		if i%2 == 0 && key == traceIDLogKey && len(keysAndValues) > i+1 {
			if value, ok := keysAndValues[i+1].(string); ok && value != "" {
				return true
			}
		}
	}

	return false
}
