## ADDED Requirements

### Requirement: In-app notification generation
The system SHALL generate in-app notifications for pharmacy personnel when prescriptions enter "approaching" status. One notification SHALL be created per prescription per status transition.

#### Scenario: Prescription enters approaching status
- **WHEN** a prescription's estimated remaining days drops to 7 or below
- **THEN** the system generates a notification for all personnel in the patient's pharmacy

#### Scenario: No duplicate notifications
- **WHEN** a notification has already been generated for a prescription entering "approaching" status
- **THEN** the system does not generate another notification for the same status transition

### Requirement: Notification display
The system SHALL display a notification indicator in the navigation showing the count of unread notifications. Personnel SHALL be able to view a list of all notifications.

#### Scenario: Unread notification indicator
- **WHEN** personnel has unread notifications
- **THEN** the system displays a count badge in the navigation

#### Scenario: View notification list
- **WHEN** personnel opens the notification list
- **THEN** the system displays notifications sorted by creation date (newest first), showing patient name, medication, and estimated depletion date

### Requirement: Notification acknowledgement
The system SHALL allow personnel to mark notifications as read, individually or all at once.

#### Scenario: Mark single notification as read
- **WHEN** personnel marks a notification as read
- **THEN** the system updates the notification status and decrements the unread count

#### Scenario: Mark all notifications as read
- **WHEN** personnel clicks "mark all as read"
- **THEN** the system marks all unread notifications as read and resets the count to zero

### Requirement: Notification scoping
Notifications SHALL only be visible to personnel belonging to the same pharmacy as the patient.

#### Scenario: Cross-pharmacy notification isolation
- **WHEN** personnel from pharmacy A views notifications
- **THEN** the system displays only notifications for pharmacy A's patients
