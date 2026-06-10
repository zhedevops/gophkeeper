-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS user_vaults (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    datatype SMALLINT NOT NULL,
    meta TEXT NOT NULL,
    filename VARCHAR(255),
    encrypted_data BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );
CREATE INDEX idx_user_vaults_user_id ON user_vaults(user_id);
CREATE UNIQUE INDEX idx_user_vaults_user_id_datatype_meta ON user_vaults(user_id, datatype, meta);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS user_vaults CASCADE;
-- +goose StatementEnd
