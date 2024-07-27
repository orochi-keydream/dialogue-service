package log

import (
	"context"
	"log/slog"
)

type ctxKey string

const logFieldsKey ctxKey = "log_fields"

type ContextHandler struct {
	slog.Handler
}

func NewContextHandler(handler slog.Handler) *ContextHandler {
	return &ContextHandler{
		Handler: handler,
	}
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs, ok := ctx.Value(logFieldsKey).([]slog.Attr)

	if ok {
		r.AddAttrs(attrs...)
	}

	return h.Handler.Handle(ctx, r)
}

func AddToContext(ctx context.Context, attrs []slog.Attr) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if curAttrs, ok := ctx.Value(logFieldsKey).([]slog.Attr); ok {
		curAttrs = append(curAttrs, attrs...)
		return context.WithValue(ctx, logFieldsKey, curAttrs)
	}

	return context.WithValue(ctx, logFieldsKey, attrs)
}
