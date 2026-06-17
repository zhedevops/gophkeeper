package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"gophkeeper/internal/client"
	"gophkeeper/internal/config"
	"gophkeeper/internal/model"
	pb "gophkeeper/proto"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// getClient Возвращает клиента
func getClient() (pb.GophkeeperServiceClient, *grpc.ClientConn, error) {
	if err := config.SetConfig(); err != nil {
		return nil, nil, err
	}
	cnf := config.GetConfig()

	return client.New(cnf)
}

// CacheDBPath Получает каталог для хранения БД кеша
func CacheDBPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "cache.db"), nil
}

// SaveToken Сохраняет JWT-токен в локальном конфигурационном файле
func SaveToken(token string) error {
	configFile, configDir, err := getConfigFilePath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return err
	}

	cfg := model.Token{
		AccessToken: token,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	log.Info().Str("config file", configFile).Msg("token received")

	return os.WriteFile(configFile, data, 0600)
}

// LoadToken Получает JWT-токен из локального конфигурационного файла
func LoadToken() (string, error) {
	configFile, _, err := getConfigFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return "", err
	}

	var cfg model.Token
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}

	return cfg.AccessToken, nil
}

// getConfigFilePath Получает путь к локальному конфигурационному файлу
func getConfigFilePath() (string, string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", "", err
	}
	configFile := filepath.Join(configDir, "config.json")

	return configFile, configDir, nil
}

// getConfigDir Получает директорию локального конфигурационного файла
func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".gophkeeper")
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return "", err
	}

	return configDir, nil
}

// ParseData Парсит новые пользовательские данные
func ParseData(opts CreateOptions) (pb.DataType, []byte, error) {
	var payload []byte
	var err error
	switch strings.ToLower(opts.Type) {
	case "credentials":
		if login == "" {
			return pb.DataType_DATA_TYPE_CREDENTIALS, nil, errors.New("login is required")
		}

		if password == "" {
			return pb.DataType_DATA_TYPE_CREDENTIALS, nil, errors.New("password is required")
		}
		payload, err = json.Marshal(CredentialsData{
			Login:    login,
			Password: password,
		})
		if err != nil {
			return pb.DataType_DATA_TYPE_CREDENTIALS, nil, err
		}
		return pb.DataType_DATA_TYPE_CREDENTIALS, payload, nil
	case "text":
		if err := validateDataOrFile(data, file); err != nil {
			return pb.DataType_DATA_TYPE_TEXT, nil, err
		}
		if data != "" {
			payload = []byte(data)
		} else if file != "" {
			fileInfo, err := os.Stat(file)
			if err != nil {
				return pb.DataType_DATA_TYPE_TEXT, nil, err
			}

			if fileInfo.Size() > MaxFileSize {
				return pb.DataType_DATA_TYPE_TEXT, nil, fmt.Errorf(
					"file size %d exceeds limit %d bytes",
					fileInfo.Size(),
					MaxFileSize,
				)
			}

			payload, err = os.ReadFile(file)
			if err != nil {
				return pb.DataType_DATA_TYPE_TEXT, nil, err
			}
		}
		return pb.DataType_DATA_TYPE_TEXT, payload, nil
	case "binary":
		if err := validateDataOrFile(data, file); err != nil {
			return pb.DataType_DATA_TYPE_BINARY, nil, err
		}
		if data != "" {
			payload = []byte(data)
		} else if file != "" {
			payload, err = os.ReadFile(file)
			if err != nil {
				return pb.DataType_DATA_TYPE_BINARY, nil, err
			}
		}
		return pb.DataType_DATA_TYPE_BINARY, payload, nil
	case "card":
		if number == "" {
			return pb.DataType_DATA_TYPE_CARD, nil, errors.New("number is required")
		}

		if holder == "" {
			return pb.DataType_DATA_TYPE_CARD, nil, errors.New("holder is required")
		}

		if expiry == "" {
			return pb.DataType_DATA_TYPE_CARD, nil, errors.New("expiry is required")
		}

		if cvv == "" {
			return pb.DataType_DATA_TYPE_CARD, nil, errors.New("cvv is required")
		}

		payload, err = json.Marshal(CardData{
			Number: number,
			Holder: holder,
			Expiry: expiry,
			CVV:    cvv,
		})
		return pb.DataType_DATA_TYPE_CARD, payload, nil
	default:
		return pb.DataType_DATA_TYPE_UNSPECIFIED, nil, fmt.Errorf("unknown type %q", opts.Type)
	}
}

func validateDataOrFile(data, file string) error {
	switch {
	case data == "" && file == "":
		return errors.New("either data or file must be specified")

	case data != "" && file != "":
		return errors.New("data and file are mutually exclusive")
	}

	return nil
}

func PrintRecord(resp *pb.GetVaultResponse) {
	fmt.Printf("ID: %d\n", resp.GetId())
	fmt.Printf("Type: %s\n", getDatatype(resp.GetDatatype()))
	fmt.Printf("Meta: %s\n", resp.GetMeta())

	switch resp.GetDatatype() {
	case pb.DataType_DATA_TYPE_TEXT:
		fmt.Printf("Data: %s\n", string(resp.GetUserdata()))

	case pb.DataType_DATA_TYPE_BINARY:
		if resp.GetFilename() != "" {
			fmt.Printf("File: %s (%d bytes)\n", resp.GetFilename(), len(resp.GetUserdata()))
		} else {
			fmt.Printf("Data: %s \n", string(resp.GetUserdata()))
		}

	case pb.DataType_DATA_TYPE_CARD:
		fmt.Printf("Card: %s\n", resp.GetUserdata())
	}
}

func PrintRecordCache(v model.VaultCache) {
	fmt.Printf("ID: %d\n", v.ID)
	fmt.Printf("Type: %s\n", getDatatype(pb.DataType(v.Datatype)))
	fmt.Printf("Meta: %s\n", v.Meta)

	switch v.Datatype {
	case int32(pb.DataType_DATA_TYPE_TEXT):
		fmt.Printf("Data: %s\n", string(v.Userdata))

	case int32(pb.DataType_DATA_TYPE_BINARY):
		if v.Filename != "" {
			fmt.Printf("File: %s (%d bytes)\n", v.Filename, len(v.Userdata))
		} else {
			fmt.Printf("Data: %s \n", string(v.Userdata))
		}

	case int32(pb.DataType_DATA_TYPE_CARD):
		fmt.Printf("Card: %s\n", v.Userdata)
	}
}

func PrintRecords(resp *pb.ListVaultsResponse) {
	for _, rec := range resp.GetItems() {
		fmt.Printf("ID: %d\n", rec.GetId())
		fmt.Printf("Type: %s\n", getDatatype(rec.GetDatatype()))
		fmt.Printf("Meta: %s\n", rec.GetMeta())
	}
}

func PrintRecordsCache(vcs []model.VaultCache) {
	for _, vc := range vcs {
		fmt.Printf("ID: %d\n", vc.ID)
		fmt.Printf("Type: %s\n", getDatatype(pb.DataType(vc.Datatype)))
		fmt.Printf("Meta: %s\n", vc.Meta)
	}
}

func getDatatype(datatype pb.DataType) string {
	switch datatype {
	case pb.DataType_DATA_TYPE_CREDENTIALS:
		return "credentials"
	case pb.DataType_DATA_TYPE_TEXT:
		return "text"
	case pb.DataType_DATA_TYPE_BINARY:
		return "binary"
	case pb.DataType_DATA_TYPE_CARD:
		return "card"
	default:
		return "unknown"
	}
}

func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 12 {
		return errors.New("password is too short")
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool

	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasLower || !hasUpper || !hasDigit || !hasSpecial {
		return errors.New("password must contain at least one digit, a lowercase letter, an uppercase letter, and a symbol")
	}

	return nil
}
