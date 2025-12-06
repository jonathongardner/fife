package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

type logKey struct{}

func NewContextLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, logKey{}, logger)
}

func ContextLogger(ctx context.Context) *logrus.Entry {
	logger, ok := ctx.Value(logKey{}).(*logrus.Entry)
	if ok {
		return logger
	}

	return logrus.WithField("context", "no")
}
