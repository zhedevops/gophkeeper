package main

import (
	"gophkeeper/internal/config"
	"gophkeeper/internal/database"
	grpcserver "gophkeeper/internal/grpc"
	"gophkeeper/internal/model"
	"gophkeeper/internal/service"
	"gophkeeper/internal/storage"
	"log"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	if err := config.SetConfig(); err != nil {
		return err
	}
	cnf := config.GetConfig()

	var completion func()

	var st model.Repository
	dsn := strings.TrimSpace(cnf.DatabaseDsn)
	if dsn != "" {
		pool, err := database.ConnectDB(dsn)
		if err != nil {
			return err
		}
		st = storage.NewDBStorage(pool)
		completion = func() {
			log.Println("database pool closed")
			database.CloseDB(pool)
		}
	}

	defer completion()

	srv := service.NewService(st, cnf)

	return grpcserver.Serve(srv, cnf)
}
