package log

import (
	"context"

	"go.uber.org/zap"

	ac "github.com/Roma7-7-7/shared-clipboard/internal/context"
)

const (
	traceIDLogKey = "traceID"
)

type (
	TracedLogger interface {
		Debugw(ctx context.Context, msg string, keysAndValues ...interface{})
		Infow(ctx context.Context, msg string, keysAndValues ...interface{})
		Warnw(ctx context.Context, msg string, keysAndValues ...interface{})
		Errorw(ctx context.Context, msg string, keysAndValues ...interface{})
	}

	ZapTracedLogger struct {
		log *zap.SugaredLogger
	}
)

func NewZapTracedLogger(logger *zap.SugaredLogger) *ZapTracedLogger {
	return &ZapTracedLogger{
		log: logger.WithOptions(zap.AddCallerSkip(1)),
	}
}

func (l *ZapTracedLogger) Debugw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, ac.TraceIDFrom(ctx))
	if len(keysAndValues) > 0 {
		log.Debugw(msg, keysAndValues...)
	} else {
		log.Debug(msg)
	}
}

func (l *ZapTracedLogger) Infow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, ac.TraceIDFrom(ctx))
	if len(keysAndValues) > 0 {
		log.Infow(msg, keysAndValues...)
	} else {
		log.Info(msg)
	}
}

func (l *ZapTracedLogger) Warnw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, ac.TraceIDFrom(ctx))
	if len(keysAndValues) > 0 {
		log.Warnw(msg, keysAndValues...)
	} else {
		log.Warn(msg)
	}
}

func (l *ZapTracedLogger) Errorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, ac.TraceIDFrom(ctx))
	if len(keysAndValues) > 0 {
		log.Errorw(msg, keysAndValues...)
	} else {
		log.Error(msg)
	}
}
