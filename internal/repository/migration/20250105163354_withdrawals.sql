-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS withdrawals (
    user_id varchar REFERENCES users ON DELETE CASCADE,
    order_id varchar NOT NULL,
    processed_at timestamp with time zone NOT NULL,
    sum int,
    CONSTRAINT withdrawals_pk PRIMARY KEY (order_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
-- +goose StatementEnd
