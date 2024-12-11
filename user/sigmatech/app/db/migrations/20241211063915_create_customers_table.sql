-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS customers (
    uuid UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    is_active boolean NOT NULL DEFAULT 'false',
    last_login timestamp without time zone,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customers_email ON customers (email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_customers_email;

DROP TABLE IF EXISTS customers;
-- +goose StatementEnd
