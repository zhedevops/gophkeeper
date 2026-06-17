package interceptor

import (
	"context"
	"gophkeeper/internal/model"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	ParseAuthToken(string) (model.User, error)
	SetUser(context.Context, model.User) context.Context
}

var publicMethods = map[string]struct{}{
	"/grpcserv.GophkeeperService/Login":    {},
	"/grpcserv.GophkeeperService/Register": {},
}

func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return handler(ctx, req)
	}
}

func UnaryInterceptor(service AuthService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := publicMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		const prefix = "Bearer "

		if !strings.HasPrefix(values[0], prefix) {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization scheme")
		}

		token := strings.TrimPrefix(values[0], prefix)
		user, err := service.ParseAuthToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		ctx = service.SetUser(ctx, user)

		return handler(ctx, req)
	}
}
