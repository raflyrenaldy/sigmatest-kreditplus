-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS customer_limits (
    uuid UUID PRIMARY KEY,
    customer_uuid UUID REFERENCES customers(uuid) ON DELETE CASCADE,
    term INTEGER DEFAULT 0,
    status BOOLEAN DEFAULT false,
    amount_limit DECIMAL(15, 2) DEFAULT 0,
    remaining_limit DECIMAL(15, 2) DEFAULT 0,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS customer_limits;
-- +goose StatementEnd
