package cache

import (
	"context"
	"database/sql"
	"gophkeeper/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func newMockRepository(t *testing.T, dbPath string) *Repository {
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)

	err = migrate(db)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
		_, _ = db.Exec(`DELETE FROM vault_cache`)
	})

	return &Repository{db: db}
}

func TestIntegration(t *testing.T) {
	repo := newMockRepository(t, "../../testdata/test_cache.db")

	ctx := context.Background()
	vc := model.VaultCache{
		ID:       1,
		Datatype: 1,
		Meta:     "meta",
		Filename: "fname",
		Userdata: []byte("userdata"),
	}
	var saved model.VaultCache

	t.Run("insert new vault ok", func(t *testing.T) {
		err := repo.Save(ctx, vc)
		require.NoError(t, err)

		row := repo.db.QueryRowContext(
			ctx,
			`SELECT id, datatype, meta, filename, userdata FROM vault_cache WHERE id = ?`,
			vc.ID,
		)

		err = row.Scan(&saved.ID, &saved.Datatype, &saved.Meta, &saved.Filename, &saved.Userdata)
		require.NoError(t, err)
		require.Equal(t, vc, saved)
	})

	t.Run("get vault", func(t *testing.T) {
		vc, err := repo.Get(ctx, saved.ID)
		require.NoError(t, err)
		require.Equal(t, vc, saved)
	})

	t.Run("get vault err", func(t *testing.T) {
		vc, err := repo.Get(ctx, 2)
		require.Error(t, err)
		require.Equal(t, model.ErrVaultNotFound, err)
		require.Equal(t, vc, model.VaultCache{})
	})

	t.Run("list vaults", func(t *testing.T) {
		vcs, err := repo.List(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, len(vcs))
		require.Equal(t, vcs[0].ID, saved.ID)
		require.Equal(t, vcs[0].Datatype, saved.Datatype)
		require.Equal(t, vcs[0].Meta, saved.Meta)
	})

	t.Run("delete vault", func(t *testing.T) {
		err := repo.Delete(ctx, saved.ID)
		require.NoError(t, err)

		vc, err := repo.Get(ctx, saved.ID)
		require.Error(t, err)
		require.Equal(t, model.ErrVaultNotFound, err)
		require.Equal(t, vc, model.VaultCache{})
	})

	t.Run("delete vault err", func(t *testing.T) {
		err := repo.Delete(ctx, 2)
		require.Error(t, err)
		require.Equal(t, model.ErrVaultNotFound, err)
	})
}
