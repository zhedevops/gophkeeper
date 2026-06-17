package config

import (
	"errors"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type EnvParams struct {
	DatabaseDsn *string `env:"DATABASE_DSN"`
	Key         *string `env:"KEY"`
	GRPCAddress *string `env:"GRPC_ADDRESS" envDefault:":3200"`
	TLSCert     *string `env:"TLS_CERT"`
	TLSKey      *string `env:"TLS_KEY"`
	MasterKey   *string `env:"MASTER_KEY"`
}

// Config Тип конфигурации, содержащий всё необходимую информацю для работы сервиса.
type Config struct {
	GRPCAddress string
	DatabaseDsn string
	Security
}

type Security struct {
	MasterKey string
	SecretKey string
	TLSCert   string
	TLSKey    string
}

var cfg = &Config{}

// SetConfig Устанавливает конфигурацию.
func SetConfig() error {
	return parseEnvParams()
}

// GetConfig Возвращает конфигурацию.
func GetConfig() *Config {
	return cfg
}

// parseEnvParams Парсит параметры из .env
func parseEnvParams() error {
	_ = godotenv.Load(".env")
	var params EnvParams
	if err := env.Parse(&params); err != nil {
		return err
	}

	if isEmpty(params.DatabaseDsn) ||
		isEmpty(params.Key) ||
		isEmpty(params.TLSCert) ||
		isEmpty(params.TLSKey) ||
		isEmpty(params.MasterKey) {
		return errors.New("missing required environment variables")
	}

	cfg.DatabaseDsn = *params.DatabaseDsn
	cfg.GRPCAddress = *params.GRPCAddress
	cfg.Security.SecretKey = *params.Key
	cfg.Security.TLSCert = *params.TLSCert
	cfg.Security.TLSKey = *params.TLSKey
	cfg.Security.MasterKey = *params.MasterKey

	return nil
}

func isEmpty(s *string) bool {
	return s == nil || *s == ""
}
