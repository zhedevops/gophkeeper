package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"gophkeeper/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*DBStorage, func(db *pgxpool.Pool)) {
	t.Helper()

	_ = godotenv.Load("../../.env")
	dsn, _ := os.LookupEnv("DATABASE_DSN_TEST")

	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 30; i++ {
		pool, err = pgxpool.New(context.Background(), dsn)
		if err == nil {
			err = pool.Ping(context.Background())
			if err == nil {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	require.NoError(t, err, "Postgres is not ready")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	err = goose.UpContext(ctx, db, "../../migrations")
	require.NoError(t, err)

	repo := NewDBStorage(pool)

	teardown := func(db *pgxpool.Pool) {
		_, _ = db.Exec(context.Background(), `
		TRUNCATE TABLE users, user_vaults RESTART IDENTITY CASCADE
	`)
	}
	return repo, teardown
}

func TestIntegration(t *testing.T) {
	repo, teardown := setupTestDB(t)
	defer teardown(repo.db)

	ctx := context.Background()

	var userID int32
	var vaultID int32

	t.Run("insert new user ok", func(t *testing.T) {
		createdUserID, err := repo.CreateUser(ctx, "testuser", "testpassword")
		require.NoError(t, err)
		require.NotZero(t, createdUserID)
		userID = createdUserID
	})

	t.Run("insert new user err", func(t *testing.T) {
		_, err := repo.CreateUser(ctx, "testuser", "testpassword")
		require.Error(t, err)
	})

	t.Run("login user ok", func(t *testing.T) {
		ID, _, err := repo.LoginUser(ctx, "testuser")
		require.NoError(t, err)
		require.Equal(t, userID, ID)
	})

	t.Run("login user err", func(t *testing.T) {
		_, _, err := repo.LoginUser(ctx, "testuser2")
		require.Error(t, err)
		require.ErrorIs(t, err, model.ErrUserNotFound)
	})

	t.Run("insert new vault ok", func(t *testing.T) {
		uv := model.UserVault{
			UserID:   userID,
			Datatype: 2,
			Meta:     "meta",
			Filename: "",
			Userdata: []byte("ciphertext"),
		}
		ID, err := repo.CreateVault(ctx, uv)
		require.NoError(t, err)
		require.NotZero(t, ID)
		vaultID = ID
	})

	t.Run("vault already exists", func(t *testing.T) {
		uv := model.UserVault{
			UserID:   userID,
			Datatype: 2,
			Meta:     "meta",
			Filename: "",
			Userdata: []byte("ciphertext"),
		}
		ID, err := repo.CreateVault(ctx, uv)
		require.ErrorIs(t, err, model.ErrVaultAlreadyExists)
		require.Zero(t, ID)
	})

	t.Run("get vault ok", func(t *testing.T) {
		uv, err := repo.GetVault(ctx, vaultID, userID)
		require.NoError(t, err)
		require.Equal(t, "meta", uv.Meta)
		require.Equal(t, int32(2), uv.Datatype)
	})

	t.Run("get vault error", func(t *testing.T) {
		uv, err := repo.GetVault(ctx, 99, userID)
		require.ErrorIs(t, err, model.ErrVaultNotFound)
		require.Equal(t, model.UserVault{}, uv)
	})

	t.Run("list vault ok", func(t *testing.T) {
		uvs, err := repo.ListVaults(ctx, userID)
		require.NoError(t, err)
		require.Len(t, uvs, 1)
	})

	t.Run("list vault error", func(t *testing.T) {
		uvs, err := repo.ListVaults(ctx, 99)
		require.NoError(t, err)
		require.Len(t, uvs, 0)
	})

	t.Run("delete vault error", func(t *testing.T) {
		err := repo.DeleteVault(ctx, 99, userID)
		require.ErrorIs(t, err, model.ErrVaultNotFound)
	})

	t.Run("delete vault ok", func(t *testing.T) {
		err := repo.DeleteVault(ctx, vaultID, userID)
		require.NoError(t, err)
	})
}
