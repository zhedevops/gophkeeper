// Package service Сервис, отвечающий за обработку запросов обработчика.
package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	pb "gophkeeper/proto"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gophkeeper/internal/config"
	"gophkeeper/internal/model"
)

type contextKey string

const userContextKey contextKey = "user"

// generate:reset
type Service struct {
	repo     model.Repository
	cfg      *config.Config
	hashFunc func(string) (string, error)
}

// NewService Создаёт сервис.
func NewService(r model.Repository, cnf *config.Config) *Service {
	return &Service{
		repo:     r,
		cfg:      cnf,
		hashFunc: HashPassword,
	}
}

func (srv *Service) Ping(ctx context.Context) error {
	return srv.repo.Ping(ctx)
}

// CreateUser Создаёт нового пользователя.
func (srv *Service) CreateUser(ctx context.Context, username string, password string) (model.User, error) {
	user := model.User{}
	passHash, err := srv.hashFunc(password)
	if err != nil {
		return user, err
	}
	ID, err := srv.repo.CreateUser(ctx, username, passHash)
	if err != nil {
		return user, err
	}
	user.ID = ID
	user.AccessToken = srv.GetAuthToken(user)

	return user, nil
}

// LoginUser Авторизует нового пользователя.
func (srv *Service) LoginUser(ctx context.Context, username string, password string) (string, error) {
	passHash, err := srv.hashFunc(password)
	if err != nil {
		return "", err
	}
	ID, passHash, err := srv.repo.LoginUser(ctx, username)
	if err != nil {
		return "", err
	}

	err = CheckPassword(passHash, password)
	if err != nil {
		return "", model.ErrInvalidCredentials
	}

	user := model.User{ID: ID}
	accessToken := srv.GetAuthToken(user)

	return accessToken, nil
}

// SetVault Создаёт запись с данными пользователя.
func (srv *Service) SetVault(ctx context.Context, datatype pb.DataType, meta string, filename string, userdata []byte) (int32, error) {
	user, err := getUser(ctx)
	if err != nil {
		return 0, err
	}

	key, err := getMasterKey(srv.cfg.Security.MasterKey)
	if err != nil {
		return 0, err
	}
	ciphertext, err := Encrypt(key, userdata)
	if err != nil {
		return 0, err
	}

	uv := model.UserVault{
		UserID:   user.ID,
		Datatype: int32(datatype),
		Meta:     meta,
		Filename: filename,
		Userdata: ciphertext,
	}

	ID, err := srv.repo.CreateVault(ctx, uv)
	if err != nil {
		return 0, err
	}

	return ID, nil
}

// GetVault Получает запись по идентификатору с данными пользователя
func (srv *Service) GetVault(ctx context.Context, ID int32) (model.UserVault, error) {
	uv, err := srv.repo.GetVault(ctx, ID)
	if err != nil {
		return model.UserVault{}, err
	}

	key, err := getMasterKey(srv.cfg.Security.MasterKey)
	if err != nil {
		return model.UserVault{}, err
	}
	plaintext, err := Decrypt(key, uv.Userdata)
	if err != nil {
		return model.UserVault{}, err
	}

	uv.Userdata = plaintext

	return uv, nil
}

// ListVaults Получает список пользовательских данных
func (srv *Service) ListVaults(ctx context.Context) ([]model.UserVault, error) {
	user, err := getUser(ctx)
	if err != nil {
		return []model.UserVault{}, err
	}
	uvs, err := srv.repo.ListVaults(ctx, user.ID)
	if err != nil {
		return []model.UserVault{}, err
	}

	return uvs, nil
}

// DeleteVault Удаляет по идентификатору пользовательские данные
func (srv *Service) DeleteVault(ctx context.Context, ID int32) error {
	return srv.repo.DeleteVault(ctx, ID)
}

// HashPassword Создаёт хеш пароля пользователя
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword Проверяет пароль пользователя
func CheckPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ParseAuthToken Парсит JWT-токен пользователя
func (srv *Service) ParseAuthToken(token string) (model.User, error) {
	user := model.User{}
	ujwt := model.UserJWT{}
	values := strings.Split(token, ".")
	if len(values) != 2 {
		return model.User{}, model.ErrBadAuthToken
	}
	jwtData, err := base64.StdEncoding.DecodeString(values[0])
	if err != nil {
		return user, model.ErrDecodeAuthToken
	}
	signature, err := base64.StdEncoding.DecodeString(values[1])
	if err != nil {
		return user, model.ErrDecodeAuthTokenSignature
	}
	h := hmac.New(sha256.New, []byte(srv.cfg.Security.SecretKey))
	h.Write(jwtData)
	sign := h.Sum(nil)
	if !hmac.Equal(sign, signature) {
		return user, model.ErrSignatureVerification
	}
	if err = json.Unmarshal(jwtData, &ujwt); err != nil {
		return user, model.ErrUnmarshal
	}
	if ujwt.Exp < time.Now().Unix() {
		return user, model.ErrExpired
	}
	user.ID = ujwt.UID
	return user, nil
}

// GetAuthToken Создаёт JWT-токен.
func (srv *Service) GetAuthToken(user model.User) string {
	userJWT := model.UserJWT{
		UID: user.ID,
		Exp: time.Now().Add(time.Hour).Unix(),
	}
	userData, _ := json.Marshal(userJWT)
	h := hmac.New(sha256.New, []byte(srv.cfg.Security.SecretKey))
	h.Write(userData)
	sign := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(userData) + "." + base64.StdEncoding.EncodeToString(sign)
}

// SetUser Устанавливает пользователя в контекст
func (srv *Service) SetUser(ctx context.Context, user model.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// getUser Получает пользователя в контекста
func getUser(ctx context.Context) (model.User, error) {
	user, ok := ctx.Value(userContextKey).(model.User)
	if !ok {
		return model.User{}, status.Error(codes.Unauthenticated, "user not found")
	}

	return user, nil
}

// getMasterKey Возвращает мастер-ключ
func getMasterKey(keyHex string) ([]byte, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return []byte{}, err
	}

	if len(key) != 32 {
		return []byte{}, fmt.Errorf("master key must be 32 bytes")
	}

	return key, nil
}

// Encrypt Шифрует пользовательские данные
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return append(nonce, ciphertext...), nil
}

// Decrypt Дешифрует пользовательские данные
func Decrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()

	if len(data) < nonceSize {
		return nil, fmt.Errorf("invalid ciphertext")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}
