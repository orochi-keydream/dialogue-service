package interceptor

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	resp, err = handler(ctx, req)

	// TODO: Map different errors to different status codes.
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}
