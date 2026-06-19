package config

import (
	"errors"

	"github.com/caarlos0/env/v11"
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

// NewConfig Возвращает конфигурацию.
func NewConfig() (*Config, error) {
	var cfg = &Config{}
	_ = godotenv.Load(".env")
	var params EnvParams
	if err := env.Parse(&params); err != nil {
		return cfg, err
	}

	if isEmpty(params.DatabaseDsn) ||
		isEmpty(params.Key) ||
		isEmpty(params.TLSCert) ||
		isEmpty(params.TLSKey) ||
		isEmpty(params.MasterKey) {
		return cfg, errors.New("missing required environment variables")
	}

	cfg.DatabaseDsn = *params.DatabaseDsn
	cfg.GRPCAddress = *params.GRPCAddress
	cfg.Security.SecretKey = *params.Key
	cfg.Security.TLSCert = *params.TLSCert
	cfg.Security.TLSKey = *params.TLSKey
	cfg.Security.MasterKey = *params.MasterKey

	return cfg, nil
}

func isEmpty(s *string) bool {
	return s == nil || *s == ""
}
