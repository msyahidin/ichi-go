-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
  ADD COLUMN created_by BIGINT NULL,
  ADD COLUMN updated_by BIGINT NULL,
  ADD COLUMN deleted_at TIMESTAMPTZ NULL,
  ADD COLUMN deleted_by BIGINT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
  DROP COLUMN created_by,
  DROP COLUMN updated_by,
  DROP COLUMN deleted_at,
  DROP COLUMN deleted_by;
-- +goose StatementEnd
