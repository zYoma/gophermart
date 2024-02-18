-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_balance
ADD CONSTRAINT current_positive CHECK (current >= 0),
ADD CONSTRAINT withdrawn_positive CHECK (withdrawn >= 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_balance
DROP CONSTRAINT current_positive,
DROP CONSTRAINT withdrawn_positive;
-- +goose StatementEnd
