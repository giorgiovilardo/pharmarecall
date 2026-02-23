-- +goose Up
CREATE TABLE prescriptions (
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    patient_id        BIGINT NOT NULL,
    medication_name   VARCHAR(255) NOT NULL,
    units_per_box     INTEGER NOT NULL,
    daily_consumption NUMERIC(10, 2) NOT NULL,
    box_start_date    DATE NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_prescriptions_patient_id ON prescriptions (patient_id);

ALTER TABLE prescriptions
    ADD CONSTRAINT fk_prescriptions_patient
    FOREIGN KEY (patient_id) REFERENCES patients (id);

-- +goose Down
ALTER TABLE prescriptions DROP CONSTRAINT fk_prescriptions_patient;
DROP TABLE prescriptions;
