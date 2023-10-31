package log

import (
	"go.uber.org/zap"
)

const (
	traceIDLogKey = "traceID"
)

type (
	TracedLogger interface {
		Debugw(tid string, msg string, keysAndValues ...interface{})
		Infow(tid string, msg string, keysAndValues ...interface{})
		Warnw(tid string, msg string, keysAndValues ...interface{})
		Errorw(tid string, msg string, keysAndValues ...interface{})
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

func (l *ZapTracedLogger) Debugw(tid string, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, tid)
	if len(keysAndValues) > 0 {
		log.Debugw(msg, keysAndValues...)
	} else {
		log.Debug(msg)
	}
}

func (l *ZapTracedLogger) Infow(tid string, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, tid)
	if len(keysAndValues) > 0 {
		log.Infow(msg, keysAndValues...)
	} else {
		log.Info(msg)
	}
}

func (l *ZapTracedLogger) Warnw(tid string, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, tid)
	if len(keysAndValues) > 0 {
		log.Warnw(msg, keysAndValues...)
	} else {
		log.Warn(msg)
	}
}

func (l *ZapTracedLogger) Errorw(tid string, msg string, keysAndValues ...interface{}) {
	log := l.log.With(traceIDLogKey, tid)
	if len(keysAndValues) > 0 {
		log.Errorw(msg, keysAndValues...)
	} else {
		log.Error(msg)
	}
}
