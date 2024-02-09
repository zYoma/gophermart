-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_balance
ALTER COLUMN current TYPE FLOAT USING current::FLOAT,
ALTER COLUMN withdrawn TYPE FLOAT USING withdrawn::FLOAT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_balance
ALTER COLUMN current TYPE INTEGER USING current::INTEGER,
ALTER COLUMN withdrawn TYPE INTEGER USING withdrawn::INTEGER;
-- +goose StatementEnd
