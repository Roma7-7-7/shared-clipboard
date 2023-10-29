package trace

import (
	"context"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"
)

const (
	traceIDLogKey = "traceID"
	badgerTraceID = "badger"
)

type (
	Logger interface {
		Debugw(ctx context.Context, msg string, keysAndValues ...interface{})
		Infow(ctx context.Context, msg string, keysAndValues ...interface{})
		Warnw(ctx context.Context, msg string, keysAndValues ...interface{})
		Errorw(ctx context.Context, msg string, keysAndValues ...interface{})
	}

	SugaredLogger struct {
		log *zap.SugaredLogger
	}

	BadgerLogger struct {
		log Logger
	}
)

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

func NewBadgerLogger(log Logger) badger.Logger {
	return &BadgerLogger{
		log: log,
	}
}

func (l *BadgerLogger) Debugf(format string, v ...interface{}) {
	l.log.Debugw(backgroundContextWithBadgerTraceID(), fmt.Sprintf(strings.TrimSuffix(format, "\n"), v...))
}

func (l *BadgerLogger) Infof(format string, v ...interface{}) {
	l.log.Infow(backgroundContextWithBadgerTraceID(), fmt.Sprintf(strings.TrimSuffix(format, "\n"), v...))
}

func (l *BadgerLogger) Warningf(format string, v ...interface{}) {
	l.log.Warnw(backgroundContextWithBadgerTraceID(), fmt.Sprintf(strings.TrimSuffix(format, "\n"), v...))
}

func (l *BadgerLogger) Errorf(format string, v ...interface{}) {
	l.log.Errorw(backgroundContextWithBadgerTraceID(), fmt.Sprintf(strings.TrimSuffix(format, "\n"), v...))
}

func backgroundContextWithBadgerTraceID() context.Context {
	return WithTraceID(context.Background(), badgerTraceID)
}
