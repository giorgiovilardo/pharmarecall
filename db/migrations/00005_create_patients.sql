-- +goose Up
CREATE TABLE patients (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pharmacy_id      BIGINT NOT NULL,
    first_name       VARCHAR(255) NOT NULL,
    last_name        VARCHAR(255) NOT NULL,
    phone            VARCHAR(50) NOT NULL DEFAULT '',
    email            VARCHAR(255) NOT NULL DEFAULT '',
    delivery_address TEXT NOT NULL DEFAULT '',
    fulfillment      VARCHAR(20) NOT NULL DEFAULT 'pickup'
        CHECK (fulfillment IN ('pickup', 'shipping')),
    notes            TEXT NOT NULL DEFAULT '',
    consensus        BOOLEAN NOT NULL DEFAULT false,
    consensus_date   TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_patients_pharmacy_id ON patients (pharmacy_id);

ALTER TABLE patients
    ADD CONSTRAINT fk_patients_pharmacy
    FOREIGN KEY (pharmacy_id) REFERENCES pharmacies (id);

-- +goose Down
ALTER TABLE patients DROP CONSTRAINT fk_patients_pharmacy;
DROP TABLE patients;
