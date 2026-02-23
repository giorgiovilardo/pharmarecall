-- +goose Up
CREATE TABLE notifications (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pharmacy_id      BIGINT NOT NULL,
    prescription_id  BIGINT NOT NULL,
    transition_type  VARCHAR(20) NOT NULL CHECK (transition_type IN ('approaching')),
    read             BOOLEAN NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_pharmacy_read ON notifications (pharmacy_id, read);
CREATE UNIQUE INDEX idx_notifications_prescription_transition
    ON notifications (prescription_id, transition_type);

ALTER TABLE notifications
    ADD CONSTRAINT fk_notifications_pharmacy
    FOREIGN KEY (pharmacy_id) REFERENCES pharmacies (id);

ALTER TABLE notifications
    ADD CONSTRAINT fk_notifications_prescription
    FOREIGN KEY (prescription_id) REFERENCES prescriptions (id);

-- +goose Down
ALTER TABLE notifications DROP CONSTRAINT fk_notifications_prescription;
ALTER TABLE notifications DROP CONSTRAINT fk_notifications_pharmacy;
DROP TABLE notifications;
