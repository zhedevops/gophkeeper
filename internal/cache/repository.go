// Package cache Локальное хранилище, режим read-only
package cache

import (
	"context"
	"database/sql"
	"errors"
	"gophkeeper/internal/model"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type Repository struct {
	db *sql.DB
}

func New(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	log.Info().Msg("sqlite3 database initialized")

	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) Save(ctx context.Context, vc model.VaultCache) error {
	if err := checkDB(r); err != nil {
		return err
	}

	query := `INSERT INTO vault_cache (id, datatype, meta, filename, userdata) 
            VALUES (?, ?, ?, ?, ?) ON CONFLICT (id) DO UPDATE SET
            datatype = excluded.datatype,
            meta = excluded.meta,
            filename = excluded.filename,
            userdata = excluded.userdata`
	_, err := r.db.ExecContext(ctx, query, vc.ID, vc.Datatype, vc.Meta, vc.Filename, vc.Userdata)

	return err
}

func (r *Repository) Get(ctx context.Context, vaultId int32) (model.VaultCache, error) {
	var vc model.VaultCache
	if err := checkDB(r); err != nil {
		return vc, err
	}

	query := `SELECT id, datatype, meta, filename, userdata FROM vault_cache WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, vaultId).Scan(&vc.ID, &vc.Datatype, &vc.Meta, &vc.Filename, &vc.Userdata)
	if errors.Is(err, sql.ErrNoRows) {
		return vc, model.ErrVaultNotFound
	}

	if err != nil {
		return vc, err
	}

	return vc, nil
}

func (r *Repository) List(ctx context.Context) ([]model.VaultCache, error) {
	var vcs []model.VaultCache
	if err := checkDB(r); err != nil {
		return vcs, err
	}

	query := `SELECT id, datatype, meta FROM vault_cache`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return vcs, err
	}
	defer rows.Close()
	for rows.Next() {
		var vc model.VaultCache
		if err := rows.Scan(&vc.ID, &vc.Datatype, &vc.Meta); err != nil {
			return vcs, err
		}
		vcs = append(vcs, vc)
	}

	return vcs, nil
}

func (r *Repository) Delete(ctx context.Context, vaultId int32) error {
	if err := checkDB(r); err != nil {
		return err
	}

	query := `DELETE FROM vault_cache WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, vaultId)
	if err != nil {
		return err
	}
	num, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if num == 0 {
		return model.ErrVaultNotFound
	}
	return nil
}

func checkDB(r *Repository) error {
	if r == nil {
		return errors.New("repository is nil")
	}

	if r.db == nil {
		return errors.New("db is nil")
	}

	return nil
}
