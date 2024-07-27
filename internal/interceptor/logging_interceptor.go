package interceptor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/orochi-keydream/dialogue-service/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LoggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	md, _ := metadata.FromIncomingContext(ctx)

	requestId := ""
	ok := false

	requestIdList := md.Get("x-request-id")

	if len(requestIdList) > 0 {
		requestId = requestIdList[0]
		ok = true
	}

	if !ok {
		slog.InfoContext(ctx, "No x-request-id provided so that it will be generated")
		requestId = uuid.NewString()
	}

	attrs := []slog.Attr{
		slog.String("x-request-id", requestId),
		slog.String("endpoint", info.FullMethod),
	}

	ctx = log.AddToContext(ctx, attrs)

	slog.InfoContext(ctx, fmt.Sprintf("%s endpoint called", info.FullMethod))

	return handler(ctx, req)
}
