package client

import (
	"gophkeeper/internal/config"
	pb "gophkeeper/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// New Создаёт клиента
func New(cnf *config.Config) (pb.GophkeeperServiceClient, *grpc.ClientConn, error) {
	creds, err := credentials.NewClientTLSFromFile(
		cnf.Security.TLSCert,
		"",
	)
	if err != nil {
		return nil, nil, err
	}
	conn, err := grpc.NewClient(
		cnf.GRPCAddress,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, nil, err
	}

	client := pb.NewGophkeeperServiceClient(conn)

	return client, conn, nil
}
