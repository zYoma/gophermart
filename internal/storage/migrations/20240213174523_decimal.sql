-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_balance
ALTER COLUMN current TYPE NUMERIC USING current::NUMERIC,
ALTER COLUMN withdrawn TYPE NUMERIC USING withdrawn::NUMERIC;
ALTER TABLE withdrawals
ALTER COLUMN sum TYPE NUMERIC USING sum::NUMERIC;
ALTER TABLE orders
ALTER COLUMN accrual TYPE NUMERIC USING accrual::NUMERIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_balance
ALTER COLUMN current TYPE FLOAT USING current::FLOAT,
ALTER COLUMN withdrawn TYPE FLOAT USING withdrawn::FLOAT;
ALTER TABLE withdrawals
ALTER COLUMN sum TYPE FLOAT USING sum::FLOAT;
ALTER TABLE orders
ALTER COLUMN accrual TYPE FLOAT USING accrual::FLOAT;
-- +goose StatementEnd
