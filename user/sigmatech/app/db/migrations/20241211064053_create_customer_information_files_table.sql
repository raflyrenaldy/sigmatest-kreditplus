-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS customer_information_files (
    uuid UUID PRIMARY KEY,
    customer_uuid UUID REFERENCES customers(uuid) ON DELETE CASCADE,
    cif_number VARCHAR(255) UNIQUE NOT NULL,
    nik VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    place_of_birth VARCHAR(255) NOT NULL,
    date_of_birth DATE NOT NULL,
    gender CHAR(1) NOT NULL,
    salary DECIMAL(15, 2) DEFAULT 0,
    card_photo VARCHAR(255) NOT NULL,
    selfie_photo VARCHAR(255) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customer_information_files_nik ON customer_information_files (nik);
CREATE INDEX IF NOT EXISTS idx_customer_information_files_cif_number ON customer_information_files (cif_number);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_customer_information_files_nik;
DROP INDEX IF EXISTS idx_customer_information_files_cif_number;

DROP TABLE IF EXISTS customer_information_files;
-- +goose StatementEnd
