-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
  ADD COLUMN created_by BIGINT UNSIGNED NULL AFTER created_at,
  ADD COLUMN updated_by BIGINT UNSIGNED NULL AFTER updated_at,
  ADD COLUMN deleted_at DATETIME NULL AFTER updated_by,
  ADD COLUMN deleted_by BIGINT UNSIGNED NULL AFTER deleted_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
  DROP COLUMN created_by,
  DROP COLUMN updated_by,
  DROP COLUMN deleted_at,
  DROP COLUMN deleted_by;
-- +goose StatementEnd
