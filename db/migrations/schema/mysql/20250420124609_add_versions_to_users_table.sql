-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN versions BIGINT NOT NULL DEFAULT 1 AFTER id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN versions;
-- +goose StatementEnd
