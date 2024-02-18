-- +goose Up
-- +goose StatementBegin
ALTER TABLE withdrawals
ALTER COLUMN sum TYPE FLOAT USING sum::FLOAT;
ALTER TABLE orders
ALTER COLUMN accrual TYPE FLOAT USING accrual::FLOAT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE withdrawals
ALTER COLUMN sum TYPE FLOAT USING sum::INTEGER;
ALTER TABLE orders
ALTER COLUMN accrual TYPE FLOAT USING accrual::INTEGER;
-- +goose StatementEnd
