// Package storage Хранилище данных
package storage

import (
	"context"
	"errors"
	"fmt"
	"gophkeeper/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBStorage Хранилище данных
type DBStorage struct {
	db *pgxpool.Pool
}

// NewDBStorage Создаёт хранилище
func NewDBStorage(pool *pgxpool.Pool) *DBStorage {
	return &DBStorage{
		db: pool,
	}
}

// Ping Проверяет доступность хранилища
func (dbs *DBStorage) Ping(ctx context.Context) error {
	return dbs.db.Ping(ctx)
}

// CreateUser Сохраняет нового пользователя
func (dbs *DBStorage) CreateUser(ctx context.Context, username string, passHash string) (int32, error) {
	var ID int32
	sql := `INSERT INTO users (login, password_hash) VALUES ($1, $2) ON CONFLICT (login) DO NOTHING RETURNING id;`
	err := dbs.db.QueryRow(ctx, sql, username, passHash).Scan(&ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, model.ErrUserAlreadyExists
	}
	return ID, err
}

// LoginUser Получает авторизационные данные пользователя
func (dbs *DBStorage) LoginUser(ctx context.Context, username string) (int32, string, error) {
	var ID int32
	var passHash string
	sql := `SELECT id, password_hash FROM users WHERE login = $1;`
	err := dbs.db.QueryRow(ctx, sql, username).Scan(&ID, &passHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", model.ErrUserNotFound
		}
		return 0, "", err
	}
	return ID, passHash, err
}

// CreateVault Создаёт запись с пользовательскими данными
func (dbs *DBStorage) CreateVault(ctx context.Context, uv model.UserVault) (int32, error) {
	fmt.Println("db create vault: ", uv)
	var ID int32
	sql := `INSERT INTO user_vaults (user_id, datatype, meta, filename, encrypted_data) 
            VALUES ($1, $2, $3, $4, $5) ON CONFLICT (user_id, datatype, meta) DO NOTHING RETURNING id;`
	err := dbs.db.QueryRow(ctx, sql, uv.UserID, uv.Datatype, uv.Meta, uv.Filename, uv.Userdata).Scan(&ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ID, model.ErrVaultAlreadyExists
	}
	if err != nil {
		return ID, err
	}

	return ID, nil
}

// GetVault Получае запись с пользовательскими данными
func (dbs *DBStorage) GetVault(ctx context.Context, ID int32) (model.UserVault, error) {
	uv := model.UserVault{}
	sql := `SELECT user_id, datatype, meta, filename, encrypted_data FROM user_vaults WHERE id = $1;`
	err := dbs.db.QueryRow(ctx, sql, ID).Scan(&uv.UserID, &uv.Datatype, &uv.Meta, &uv.Filename, &uv.Userdata)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uv, model.ErrVaultNotFound
		}
		return uv, err
	}

	return uv, nil
}

// ListVaults Получает список пользовательских данных
func (dbs *DBStorage) ListVaults(ctx context.Context, userID int32) ([]model.UserVault, error) {
	var uvs []model.UserVault
	sql := `SELECT id, datatype, meta FROM user_vaults WHERE user_id = $1;`
	rows, err := dbs.db.Query(ctx, sql, userID)
	if err != nil {
		return uvs, err
	}
	defer rows.Close()
	for rows.Next() {
		uv := model.UserVault{}
		err = rows.Scan(&uv.ID, &uv.Datatype, &uv.Meta)
		if err != nil {
			return uvs, err
		}
		uvs = append(uvs, uv)
	}
	return uvs, nil
}

// DeleteVault Удаляет запись с пользовательскими данными
func (dbs *DBStorage) DeleteVault(ctx context.Context, ID int32) error {
	cmdTag, err := dbs.db.Exec(ctx, "DELETE FROM user_vaults WHERE id = $1;", ID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return model.ErrVaultNotFound
	}
	return nil
}
