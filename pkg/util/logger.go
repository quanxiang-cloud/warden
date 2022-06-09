package util

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

// ContextKey context key
type ContextKey struct{}

// SetCtx  set ctx
func SetCtx(ctx context.Context, key, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// LoggerFromContext logger format
func LoggerFromContext(ctx context.Context) logr.Logger {
	log, ok := ctx.Value(ContextKey{}).(logr.Logger)
	if !ok {
		zapLog, err := zap.NewDevelopment()
		if err != nil {
			panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
		}
		log = zapr.NewLogger(zapLog)
		log.Error(fmt.Errorf("the log processor has not been initialized"), "context")
	}

	return log
}
