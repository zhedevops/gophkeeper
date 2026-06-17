package grpc

import (
	"context"
	"errors"
	"fmt"
	"gophkeeper/internal/config"
	"gophkeeper/internal/interceptor"
	"gophkeeper/internal/model"
	"gophkeeper/internal/service"
	"os"
	"os/signal"
	"syscall"

	pb "gophkeeper/proto"

	"net"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Service interface {
	CreateUser(ctx context.Context, username string, password string) (model.User, error)
	LoginUser(ctx context.Context, username string, password string) (string, error)
	SetVault(ctx context.Context, datatype pb.DataType, meta string, filename string, userdata []byte) (int32, error)
	GetVault(ctx context.Context, ID int32) (model.UserVault, error)
	ListVaults(ctx context.Context) ([]model.UserVault, error)
	DeleteVault(ctx context.Context, ID int32) error
}

type GophkeeperServiceServer struct {
	pb.UnimplementedGophkeeperServiceServer

	service Service
}

func Serve(service *service.Service, cnf *config.Config) error {
	// Нужно определить порт для сервера
	listen, err := net.Listen("tcp", cnf.GRPCAddress)
	if err != nil {
		return fmt.Errorf("ошибка при инициализации listener: %w", err)
	}
	defer listen.Close()
	creds, err := credentials.NewServerTLSFromFile(
		cnf.Security.TLSCert,
		cnf.Security.TLSKey,
	)
	if err != nil {
		return err
	}
	// Создаем gRPC сервер без зарегистрированной службы
	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			interceptor.TimeoutInterceptor(5*time.Second),
			interceptor.UnaryInterceptor(service),
		))
	// Регистрируем сервис-хендлер
	grpcHandler := New(service)

	pb.RegisterGophkeeperServiceServer(s, grpcHandler)

	log.Info().Str("addr", cnf.GRPCAddress).Msg("gRPC server has started")

	// Канал для получения сигналов прерывания
	signalChan := make(chan os.Signal, 1)
	// Канал для обработки ошибки
	errChan := make(chan error, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		// Получение запроса gRpc
		if err := s.Serve(listen); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	select {
	case sig := <-signalChan:
		// Когда будет получен сигнал прерывания выполнится код
		s.GracefulStop()
		<-errChan
		log.Info().Any("signal", sig).Msg("gRPC server terminated on signal")
		return nil
	case err := <-errChan:
		// Если запуск сервиса вернул ошибку
		return err
	}
}

func New(service *service.Service) *GophkeeperServiceServer {
	return &GophkeeperServiceServer{
		service: service,
	}
}

func (s *GophkeeperServiceServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	var response pb.RegisterResponse

	resp, err := s.service.CreateUser(ctx, in.GetUsername(), in.GetPassword())
	if err != nil {
		return nil, err
	}
	response.SetUserId(resp.ID)
	response.SetAccessToken(resp.AccessToken)

	return &response, nil
}

func (s *GophkeeperServiceServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	var response pb.LoginResponse

	resp, err := s.service.LoginUser(ctx, in.GetUsername(), in.GetPassword())
	if err != nil {
		return nil, err
	}
	response.SetAccessToken(resp)

	return &response, nil
}

func (s *GophkeeperServiceServer) SetVault(ctx context.Context, in *pb.SetVaultRequest) (*pb.SetVaultResponse, error) {
	var response pb.SetVaultResponse

	id, err := s.service.SetVault(ctx, in.GetDatatype(), in.GetMeta(), in.GetFilename(), in.GetUserdata())
	if err != nil {
		return nil, err
	}
	response.SetId(id)

	return &response, nil
}

func (s *GophkeeperServiceServer) GetVault(ctx context.Context, in *pb.GetVaultRequest) (*pb.GetVaultResponse, error) {
	var response pb.GetVaultResponse

	uv, err := s.service.GetVault(ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	response.SetId(in.GetId())
	response.SetDatatype(pb.DataType(uv.Datatype))
	response.SetMeta(uv.Meta)
	response.SetFilename(uv.Filename)
	response.SetUserdata(uv.Userdata)

	return &response, nil
}

func (s *GophkeeperServiceServer) ListVaults(ctx context.Context, in *pb.ListVaultsRequest) (*pb.ListVaultsResponse, error) {
	var response pb.ListVaultsResponse

	resp, err := s.service.ListVaults(ctx)
	if err != nil {
		return nil, err
	}

	var vis []*pb.VaultInfo
	for _, uv := range resp {
		vi := pb.VaultInfo{}
		vi.SetId(uv.ID)
		vi.SetDatatype(pb.DataType(uv.Datatype))
		vi.SetMeta(uv.Meta)
		vis = append(vis, &vi)
	}

	response.SetItems(vis)

	return &response, nil
}

func (s *GophkeeperServiceServer) DeleteVault(ctx context.Context, in *pb.DeleteVaultRequest) (*pb.DeleteVaultResponse, error) {
	return &pb.DeleteVaultResponse{}, s.service.DeleteVault(ctx, in.GetId())
}
