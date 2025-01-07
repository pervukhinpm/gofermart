-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id varchar UNIQUE NOT NULL,
    login varchar UNIQUE NOT NULL,
    password varchar NOT NULL,
    balance int DEFAULT 0,
    withdrawn int DEFAULT 0,
    CONSTRAINT users_pk PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
