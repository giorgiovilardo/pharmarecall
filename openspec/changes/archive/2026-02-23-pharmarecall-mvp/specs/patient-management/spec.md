## ADDED Requirements

### Requirement: Patient creation
The system SHALL allow pharmacy personnel to create a patient record with first name, last name, at least one contact method (phone or email), an optional delivery address, and optional notes. Each patient SHALL be scoped to the pharmacy of the creating personnel.

#### Scenario: Successful patient creation with phone
- **WHEN** personnel submits a valid patient form with first name, last name, and phone number
- **THEN** the system creates the patient record associated with the personnel's pharmacy

#### Scenario: Successful patient creation with email
- **WHEN** personnel submits a valid patient form with first name, last name, and email
- **THEN** the system creates the patient record associated with the personnel's pharmacy

#### Scenario: Successful patient creation with delivery address
- **WHEN** personnel submits a valid patient form with first name, last name, phone or email, and a delivery address
- **THEN** the system creates the patient record with the delivery address stored

#### Scenario: Missing required fields
- **WHEN** personnel submits a patient form without first name, last name, or without at least one of phone/email
- **THEN** the system displays validation errors and does not create the patient

### Requirement: Fulfillment preference
The system SHALL allow personnel to set a fulfillment preference for each patient: "pickup" (patient comes to the pharmacy) or "shipping" (pharmacy sends the order to the patient's delivery address). The default SHALL be "pickup". If fulfillment is set to "shipping", a delivery address SHALL be required.

#### Scenario: Patient set to pickup
- **WHEN** personnel sets a patient's fulfillment preference to "pickup"
- **THEN** the system stores the preference and does not require a delivery address

#### Scenario: Patient set to shipping
- **WHEN** personnel sets a patient's fulfillment preference to "shipping" and a delivery address is present
- **THEN** the system stores the preference

#### Scenario: Shipping without delivery address
- **WHEN** personnel sets fulfillment to "shipping" for a patient without a delivery address
- **THEN** the system displays a validation error requiring a delivery address

### Requirement: Consensus tracking
The system SHALL record whether the patient has given consensus to be tracked in the system. A patient SHALL NOT be considered active until consensus is recorded. The consensus date SHALL be stored.

#### Scenario: Patient gives consensus
- **WHEN** personnel marks a patient as having given consensus
- **THEN** the system records the consensus with the current date and marks the patient as active

#### Scenario: Patient without consensus
- **WHEN** personnel views a patient who has not given consensus
- **THEN** the system displays the patient as inactive and shows a prompt to record consensus

### Requirement: Patient listing
The system SHALL display a list of all patients belonging to the personnel's pharmacy, showing name, contact info, consensus status, and number of active prescriptions.

#### Scenario: Personnel views patient list
- **WHEN** personnel navigates to the patient list page
- **THEN** the system displays all patients for their pharmacy sorted by last name

#### Scenario: Empty patient list
- **WHEN** personnel views the patient list and no patients exist
- **THEN** the system displays a message prompting to add the first patient

### Requirement: Patient editing
The system SHALL allow personnel to edit patient details (name, phone, email, delivery address, notes). Personnel SHALL only be able to edit patients belonging to their pharmacy.

#### Scenario: Successful patient edit
- **WHEN** personnel submits updated patient details
- **THEN** the system updates the patient record

#### Scenario: Edit patient from different pharmacy
- **WHEN** personnel attempts to edit a patient not belonging to their pharmacy
- **THEN** the system denies access
