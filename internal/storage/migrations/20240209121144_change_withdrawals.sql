-- +goose Up
-- +goose StatementBegin
ALTER TABLE withdrawals DROP CONSTRAINT withdrawals_order_number_fkey;
ALTER TABLE withdrawals RENAME COLUMN order_number TO "order";
ALTER TABLE withdrawals
ADD COLUMN user_login character varying(100) NOT NULL;
ALTER TABLE withdrawals
ADD CONSTRAINT fk_withdrawals_user_login FOREIGN KEY (user_login) REFERENCES users(login);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE withdrawals RENAME COLUMN "order" TO order_number;
ALTER TABLE withdrawals
ADD CONSTRAINT withdrawals_order_number_fkey FOREIGN KEY (order_number) REFERENCES orders(number);
ALTER TABLE withdrawals DROP CONSTRAINT fk_withdrawals_user_login;
ALTER TABLE withdrawals DROP COLUMN user_login;
-- +goose StatementEnd
