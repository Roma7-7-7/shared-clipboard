package log

import (
	"strings"

	"go.uber.org/zap"
)

type ZapBadgerLogger struct {
	log *zap.SugaredLogger
}

func NewZapBadger(log *zap.SugaredLogger) *ZapBadgerLogger {
	return &ZapBadgerLogger{
		log: log.WithOptions(zap.AddCallerSkip(1)),
	}
}

func (l *ZapBadgerLogger) Debugf(format string, v ...interface{}) {
	l.log.Debugf(strings.TrimSuffix(format, "\n"), v...)
}

func (l *ZapBadgerLogger) Infof(format string, v ...interface{}) {
	l.log.Infof(strings.TrimSuffix(format, "\n"), v...)
}

func (l *ZapBadgerLogger) Warningf(format string, v ...interface{}) {
	l.log.Warnf(strings.TrimSuffix(format, "\n"), v...)
}

func (l *ZapBadgerLogger) Errorf(format string, v ...interface{}) {
	l.log.Errorf(strings.TrimSuffix(format, "\n"), v...)
}
