## ADDED Requirements

### Requirement: Prescription creation
The system SHALL allow personnel to create a recurring prescription for a patient. A prescription SHALL include: medication name, units per box, daily consumption rate (units per day), and the date the current box was started.

#### Scenario: Successful prescription creation
- **WHEN** personnel submits a prescription form with medication name, units per box, daily consumption, and box start date for an active patient
- **THEN** the system creates the prescription record linked to the patient

#### Scenario: Prescription for inactive patient
- **WHEN** personnel attempts to create a prescription for a patient without recorded consensus
- **THEN** the system displays an error requiring patient consensus before adding prescriptions

### Requirement: Estimated depletion calculation
The system SHALL calculate the estimated depletion date for each prescription based on: box start date + (units per box / daily consumption rate) days.

#### Scenario: Depletion date calculation
- **WHEN** a prescription has a box started on 2026-01-01 with 30 units and a consumption rate of 1 unit/day
- **THEN** the system calculates the estimated depletion date as 2026-01-31

#### Scenario: Fractional consumption
- **WHEN** a prescription has 100 units and a consumption rate of 3 units/day
- **THEN** the system calculates depletion at 33 days (rounded down) from the box start date

### Requirement: Prescription listing per patient
The system SHALL display all prescriptions for a given patient, showing medication name, units remaining (estimated), estimated depletion date, and status (ok, approaching depletion, depleted).

#### Scenario: View patient prescriptions
- **WHEN** personnel views a patient's detail page
- **THEN** the system displays all prescriptions with current estimated status

### Requirement: Prescription editing
The system SHALL allow personnel to edit prescription details (medication name, units per box, daily consumption, box start date).

#### Scenario: Successful prescription edit
- **WHEN** personnel submits updated prescription details
- **THEN** the system updates the prescription and recalculates the depletion date

### Requirement: Refill recording
The system SHALL allow personnel to record a refill, which resets the box start date to the refill date, starting a new depletion cycle. The previous box data SHALL be preserved for history. If an active order exists for the previous cycle, it SHALL be marked as fulfilled automatically.

#### Scenario: Record a refill
- **WHEN** personnel records a refill for a prescription
- **THEN** the system updates the box start date to the refill date and stores the previous box period in history

#### Scenario: Refill auto-fulfills active order
- **WHEN** personnel records a refill and a pending or prepared order exists for the previous cycle
- **THEN** the system marks that order as fulfilled

### Requirement: Prescription status classification
The system SHALL classify each prescription into one of three statuses based on estimated remaining days: "ok" (more than 7 days remaining), "approaching" (7 days or fewer remaining), "depleted" (0 or fewer days remaining).

#### Scenario: Prescription with 10 days remaining
- **WHEN** a prescription has an estimated 10 days of supply left
- **THEN** the system classifies it as "ok"

#### Scenario: Prescription with 5 days remaining
- **WHEN** a prescription has an estimated 5 days of supply left
- **THEN** the system classifies it as "approaching"

#### Scenario: Prescription past depletion
- **WHEN** a prescription's estimated depletion date has passed
- **THEN** the system classifies it as "depleted"
