-- +goose Up
CREATE TABLE refill_history (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    prescription_id BIGINT NOT NULL,
    box_start_date  DATE NOT NULL,
    box_end_date    DATE NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refill_history_prescription_id ON refill_history (prescription_id);

ALTER TABLE refill_history
    ADD CONSTRAINT fk_refill_history_prescription
    FOREIGN KEY (prescription_id) REFERENCES prescriptions (id);

-- +goose Down
ALTER TABLE refill_history DROP CONSTRAINT fk_refill_history_prescription;
DROP TABLE refill_history;
