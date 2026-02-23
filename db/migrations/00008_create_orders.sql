-- +goose Up
CREATE TABLE orders (
    id                      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    prescription_id         BIGINT NOT NULL,
    cycle_start_date        DATE NOT NULL,
    estimated_depletion_date DATE NOT NULL,
    status                  VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'prepared', 'fulfilled')),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_prescription_id ON orders (prescription_id);
CREATE INDEX idx_orders_status ON orders (status);

ALTER TABLE orders
    ADD CONSTRAINT fk_orders_prescription
    FOREIGN KEY (prescription_id) REFERENCES prescriptions (id);

-- +goose Down
ALTER TABLE orders DROP CONSTRAINT fk_orders_prescription;
DROP TABLE orders;
