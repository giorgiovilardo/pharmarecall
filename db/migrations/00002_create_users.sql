-- +goose Up
CREATE TABLE users (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email      VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    name       VARCHAR(255) NOT NULL,
    role       VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'owner', 'personnel')),
    pharmacy_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_pharmacy_id ON users (pharmacy_id);

-- +goose Down
DROP TABLE users;
