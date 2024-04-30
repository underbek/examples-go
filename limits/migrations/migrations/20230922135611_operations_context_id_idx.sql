-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS operations_context_id_idx ON operations USING hash(context_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS operations_context_id_idx;
-- +goose StatementEnd
