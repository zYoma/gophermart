-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    login VARCHAR(100) PRIMARY KEY,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE orders (
    number VARCHAR(100) PRIMARY KEY,
    user_login VARCHAR(100) NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL,
    accrual INTEGER,
    FOREIGN KEY (user_login) REFERENCES users(login),
    UNIQUE (number, user_login)
);
CREATE TABLE withdrawals (
    order_number VARCHAR(100) PRIMARY KEY,
    sum INTEGER NOT NULL,
    proccesed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_number) REFERENCES orders(number)
);
CREATE TABLE user_balance (
    user_login VARCHAR(100) PRIMARY KEY,
    current INTEGER NOT NULL,
    withdrawn INTEGER NOT NULL,
    FOREIGN KEY (user_login) REFERENCES users(login)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_balance;
DROP TABLE withdrawals;
DROP TABLE orders;
DROP TABLE users;
-- +goose StatementEnd
