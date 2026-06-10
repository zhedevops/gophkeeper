package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type EnvParams struct {
	DatabaseDsn *string `env:"DATABASE_DSN"`
	Key         *string `env:"KEY" envDefault:"kjdfkklsdf932.fjs"`
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

	if params.DatabaseDsn != nil {
		cfg.DatabaseDsn = *params.DatabaseDsn
	}

	if params.GRPCAddress != nil {
		cfg.GRPCAddress = *params.GRPCAddress
	}

	if params.Key != nil {
		cfg.Security.SecretKey = *params.Key
	}

	if params.TLSCert != nil {
		cfg.Security.TLSCert = *params.TLSCert
	}

	if params.TLSKey != nil {
		cfg.Security.TLSKey = *params.TLSKey
	}

	if params.MasterKey != nil {
		cfg.Security.MasterKey = *params.MasterKey
	}

	return nil
}
