-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transactions (
    uuid UUID PRIMARY KEY,
    customer_uuid UUID REFERENCES customers(uuid) ON DELETE CASCADE,
    customer_limit_uuid UUID REFERENCES customer_limits(uuid) ON DELETE SET NULL,
    asset_name VARCHAR(255) NOT NULL,
    contract_number VARCHAR(255) UNIQUE NOT NULL,
    is_done BOOLEAN DEFAULT false,
    otr DECIMAL(15, 2) DEFAULT 0,
    admin_fee DECIMAL(15, 2) DEFAULT 0,
    total DECIMAL(15, 2) DEFAULT 0,
    installment_amount DECIMAL(15, 2) DEFAULT 0,
    installment_count INTEGER DEFAULT 1,
    total_interest DECIMAL(15, 2) DEFAULT 0,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transaction_installments (
    uuid UUID PRIMARY KEY,
    transaction_uuid UUID REFERENCES transactions(uuid) ON DELETE CASCADE,
    method_payment VARCHAR(255) NULL,
    term INTEGER DEFAULT 1,
    due_date DATE NOT NULL,
    payment_at DATE NULL,
    amount DECIMAL(15, 2) DEFAULT 0,
    amount_paid DECIMAL(15, 2) DEFAULT 0,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS transaction_installments;
-- +goose StatementEnd
