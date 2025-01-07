-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    id varchar NOT NULL,
    user_id varchar REFERENCES users ON DELETE CASCADE,
    status varchar NOT NULL,
    uploaded_at timestamp with time zone NOT NULL,
    accrual int DEFAULT 0,
    CONSTRAINT orders_pk PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
