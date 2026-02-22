## ADDED Requirements

### Requirement: Daily order dashboard
The system SHALL display a dashboard showing all prescriptions that need attention, calculated on-demand. The dashboard SHALL show prescriptions grouped by status: "depleted" first, then "approaching", then "ok".

#### Scenario: Dashboard with mixed statuses
- **WHEN** personnel loads the dashboard and there are prescriptions in various states
- **THEN** the system displays depleted prescriptions first, then approaching, then ok

#### Scenario: Empty dashboard
- **WHEN** personnel loads the dashboard and no prescriptions exist
- **THEN** the system displays a message prompting to add patients and prescriptions

### Requirement: Lookahead window
The system SHALL use a configurable lookahead window (default: 7 days) to determine which prescriptions are "approaching" depletion. The dashboard SHALL show all prescriptions estimated to deplete within the lookahead window.

#### Scenario: Default lookahead
- **WHEN** personnel loads the dashboard with default settings
- **THEN** the system shows prescriptions depleting within the next 7 days as "approaching"

### Requirement: Dashboard data display
The system SHALL display for each dashboard entry: patient name, medication name, estimated depletion date, days remaining, prescription status, fulfillment method (pickup/shipping), and order status.

#### Scenario: Dashboard entry details
- **WHEN** personnel views a dashboard entry
- **THEN** the system shows patient name, medication, depletion date, days remaining, status badge, fulfillment method, and order status

### Requirement: Dashboard filtering
The system SHALL allow personnel to filter the dashboard by status (all, depleted, approaching, ok) and by date range.

#### Scenario: Filter by approaching status
- **WHEN** personnel selects the "approaching" filter
- **THEN** the system displays only prescriptions with "approaching" status

#### Scenario: Filter by date range
- **WHEN** personnel sets a date range filter
- **THEN** the system displays only prescriptions with estimated depletion within that range

### Requirement: Order status tracking
The system SHALL track an order status for each prescription approaching or past depletion. The statuses SHALL be: "pending" (default, needs preparation), "prepared" (package is ready), "fulfilled" (picked up by patient or shipped). Personnel SHALL be able to advance the status from the dashboard.

#### Scenario: Mark order as prepared
- **WHEN** personnel clicks "mark as prepared" on a pending dashboard entry
- **THEN** the system updates the order status to "prepared"

#### Scenario: Mark order as fulfilled
- **WHEN** personnel clicks "mark as fulfilled" on a prepared dashboard entry
- **THEN** the system updates the order status to "fulfilled"

#### Scenario: Dashboard shows order status
- **WHEN** personnel loads the dashboard
- **THEN** each entry displays its current order status (pending, prepared, fulfilled)

### Requirement: Dashboard filtering by order status
The system SHALL allow personnel to filter the dashboard by order status (all, pending, prepared, fulfilled) in addition to prescription status and date range filters.

#### Scenario: Filter by pending orders
- **WHEN** personnel selects the "pending" order status filter
- **THEN** the system displays only entries with "pending" order status

### Requirement: Dashboard scoping
The dashboard SHALL only display prescriptions belonging to the logged-in personnel's pharmacy.

#### Scenario: Multi-pharmacy isolation
- **WHEN** personnel from pharmacy A loads the dashboard
- **THEN** the system displays only prescriptions for patients of pharmacy A
