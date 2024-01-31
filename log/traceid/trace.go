package traceid

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	TraceID contextKey = "traceID"
)

func GetPair(ctx context.Context) (string, string) {
	if ctx == nil || ctx.Value(TraceID) == nil {
		return "", ""
	}
	return string(TraceID), ctx.Value(TraceID).(string)
}

func Get(ctx context.Context) string {
	if ctx == nil || ctx.Value(TraceID) == nil {
		return ""
	}
	return ctx.Value(TraceID).(string)
}

func Generate() string {
	return uuid.New().String()
}

func (key contextKey) String() string {
	return string(key)
}
