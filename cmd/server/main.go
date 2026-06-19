package main

import (
	"github.com/rs/zerolog/log"

	"gophkeeper/internal/config"
	"gophkeeper/internal/database"
	grpcserver "gophkeeper/internal/grpc"
	"gophkeeper/internal/model"
	"gophkeeper/internal/service"
	"gophkeeper/internal/storage"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}

func run() error {
	cnf, err := config.NewConfig()
	if err != nil {
		return err
	}

	var completion func()

	var st model.Repository
	dsn := strings.TrimSpace(cnf.DatabaseDsn)
	pool, err := database.ConnectDB(dsn)
	if err != nil {
		return err
	}
	st = storage.NewDBStorage(pool)
	completion = func() {
		log.Info().Str("addr", cnf.GRPCAddress).Msg("database pool closed")
		database.CloseDB(pool)
	}

	defer completion()

	srv := service.NewService(st, cnf)

	return grpcserver.Serve(srv, cnf)
}
