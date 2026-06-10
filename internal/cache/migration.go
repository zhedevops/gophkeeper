package cache

import "database/sql"

func migrate(db *sql.DB) error {
	const query = `
	CREATE TABLE IF NOT EXISTS vault_cache (
		id INTEGER PRIMARY KEY,
		datatype INTEGER NOT NULL,
		meta TEXT NOT NULL,
		filename TEXT,
		userdata BLOB NOT NULL
	);
	`

	_, err := db.Exec(query)

	return err
}
