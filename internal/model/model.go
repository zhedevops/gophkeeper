// Package model Содержит описание структуры модели ссылки, метода создания модели и основных типов.
package model

import (
	"context"
	"errors"
)

// Repository Интерфейс хранилища информации о ссылках.
type Repository interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, username string, password string) (int32, error)
	LoginUser(ctx context.Context, username string) (int32, string, error)
	CreateVault(ctx context.Context, uv UserVault) (int32, error)
	GetVault(ctx context.Context, ID int32) (UserVault, error)
	ListVaults(ctx context.Context, userID int32) ([]UserVault, error)
	DeleteVault(ctx context.Context, ID int32) error
}

// UserVault Пользовательские данные
type UserVault struct {
	ID       int32
	UserID   int32
	Datatype int32
	Meta     string
	Filename string
	Userdata []byte
}

// User Тип пользователя, содержащий идентификатор и токен доступа.
type User struct {
	ID          int32  `json:"id"`
	AccessToken string `json:"access_token"`
}

// UserJWT Тип, описывающий JWT пользователя, содержащий уникальный идентификатор и дату окончания действия JWT.
// generate:reset
type UserJWT struct {
	UID int32 `json:"uid"`
	Exp int64 `json:"exp"`
}

// Token JWT-токен
type Token struct {
	AccessToken string `json:"access_token"`
}

// VaultCache Тип данных для хранения в кеше на стороне клиента
type VaultCache struct {
	ID       int32
	Datatype int32
	Meta     string
	Filename string
	Userdata []byte
}

var (
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrVaultAlreadyExists       = errors.New("vault record already exists")
	ErrUserNotFound             = errors.New("user not found")
	ErrVaultNotFound            = errors.New("vault record not found")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrBadAuthToken             = errors.New("bad auth token")
	ErrDecodeAuthToken          = errors.New("decode auth token failed")
	ErrDecodeAuthTokenSignature = errors.New("decode auth token signature failed")
	ErrSignatureVerification    = errors.New("signature verification failed")
	ErrUnmarshal                = errors.New("unmarshal user data failed")
	ErrExpired                  = errors.New("user expired")
)
