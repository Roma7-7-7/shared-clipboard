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
	log := l.log.With(traceIDLogKey, ID(ctx))
	if len(keysAndValues) > 0 {
		log.Debugw(msg, keysAndValues...)
	} else {
		log.Debug(msg)
	}
}

func (l *SugaredLogger) Infow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, ID(ctx))
	if len(keysAndValues) > 0 {
		log.Infow(msg, keysAndValues...)
	} else {
		log.Info(msg)
	}
}

func (l *SugaredLogger) Warnw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, ID(ctx))
	if len(keysAndValues) > 0 {
		log.Warnw(msg, keysAndValues...)
	} else {
		log.Warn(msg)
	}
}

func (l *SugaredLogger) Errorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, ID(ctx))
	if len(keysAndValues) > 0 {
		log.Errorw(msg, keysAndValues...)
	} else {
		log.Error(msg)
	}
}
