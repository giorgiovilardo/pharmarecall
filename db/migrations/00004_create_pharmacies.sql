-- +goose Up
CREATE TABLE pharmacies (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    address    TEXT NOT NULL,
    phone      VARCHAR(50) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE users
    ADD CONSTRAINT fk_users_pharmacy
    FOREIGN KEY (pharmacy_id) REFERENCES pharmacies (id);

-- +goose Down
ALTER TABLE users DROP CONSTRAINT fk_users_pharmacy;
DROP TABLE pharmacies;
