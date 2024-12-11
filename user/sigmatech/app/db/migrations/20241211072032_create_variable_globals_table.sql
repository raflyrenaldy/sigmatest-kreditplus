-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS variable_globals (
    uuid UUID PRIMARY KEY,
    code VARCHAR(255) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description TEXT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_variable_globals_code ON variable_globals (code);

INSERT INTO variable_globals (uuid, code, value, description)
values (gen_random_uuid(), 'ADM', '2.95', 'Admin fee @ percentage')
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_variable_globals_code;

DROP TABLE IF EXISTS variable_globals;
-- +goose StatementEnd
