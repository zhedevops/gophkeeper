package interceptor

import (
	"context"
	"gophkeeper/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockService struct {
	parseCalled bool
	user        model.User
	err         error
}

func (m *mockService) ParseAuthToken(token string) (model.User, error) {
	m.parseCalled = true
	return m.user, m.err
}

func (m *mockService) SetUser(ctx context.Context, user model.User) context.Context {
	return context.WithValue(ctx, "user", user)
}

func getUser(ctx context.Context) (model.User, error) {
	user, ok := ctx.Value("user").(model.User)
	if !ok {
		return model.User{}, status.Error(codes.Unauthenticated, "user not found")
	}

	return user, nil
}

func TestUnaryInterceptor_Ok(t *testing.T) {
	svc := &mockService{
		parseCalled: false,
		user:        model.User{ID: 1},
		err:         nil,
	}

	interceptor := UnaryInterceptor(svc)

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(
			"authorization",
			"Bearer token",
		),
	)

	called := false

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true

		user, err := getUser(ctx)
		require.NoError(t, err)
		require.Equal(t, int32(1), user.ID)

		return "ok", nil
	}

	resp, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/grpcserv.GophkeeperService/GetVault"},
		handler,
	)

	require.NoError(t, err)
	require.Equal(t, "ok", resp)
	require.True(t, called)
	require.True(t, svc.parseCalled)
}

func TestUnaryInterceptor_PublicMethod(t *testing.T) {
	svc := &mockService{parseCalled: false}

	interceptor := UnaryInterceptor(svc)

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(
			"authorization",
			"Bearer token",
		),
	)

	called := false

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	resp, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/grpcserv.GophkeeperService/Login"},
		handler,
	)

	require.NoError(t, err)
	require.Equal(t, "ok", resp)
	require.True(t, called)
	require.False(t, svc.parseCalled)
}

func TestUnaryInterceptor_NoMetadata(t *testing.T) {
	svc := &mockService{parseCalled: false}

	interceptor := UnaryInterceptor(svc)

	ctx := context.Background()

	called := false

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	resp, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/grpcserv.GophkeeperService/ListVaults"},
		handler,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "missing metadata")
	require.Nil(t, resp)
	require.False(t, called)
	require.False(t, svc.parseCalled)
}

func TestUnaryInterceptor_NoToken(t *testing.T) {
	svc := &mockService{parseCalled: false}

	interceptor := UnaryInterceptor(svc)

	var MD map[string][]string

	ctx := metadata.NewIncomingContext(context.Background(), MD)

	called := false

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	resp, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/grpcserv.GophkeeperService/ListVaults"},
		handler,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "missing token")
	require.Nil(t, resp)
	require.False(t, called)
	require.False(t, svc.parseCalled)
}

func TestUnaryInterceptor_InvalidScheme(t *testing.T) {
	svc := &mockService{
		parseCalled: false,
		user:        model.User{},
		err:         status.Error(codes.Unauthenticated, "invalid authorization scheme"),
	}

	interceptor := UnaryInterceptor(svc)

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(
			"authorization",
			"Bearerxxx token",
		),
	)

	called := false

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	resp, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/grpcserv.GophkeeperService/ListVaults"},
		handler,
	)

	require.Error(t, err)
	require.Equal(t, svc.err.Error(), err.Error())
	require.Nil(t, resp)
	require.False(t, called)
	require.False(t, svc.parseCalled)
}

func TestUnaryInterceptor_InvalidToken(t *testing.T) {
	svc := &mockService{
		parseCalled: true,
		user:        model.User{},
		err:         model.ErrBadAuthToken,
	}

	interceptor := UnaryInterceptor(svc)

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(
			"authorization",
			"Bearer token",
		),
	)

	called := false

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	resp, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/grpcserv.GophkeeperService/ListVaults"},
		handler,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "bad auth token")
	require.Nil(t, resp)
	require.False(t, called)
	require.True(t, svc.parseCalled)
}
