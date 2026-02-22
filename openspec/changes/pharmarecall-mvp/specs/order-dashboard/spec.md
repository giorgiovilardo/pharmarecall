## ADDED Requirements

### Requirement: Daily order dashboard
The system SHALL display a dashboard showing active orders (pending and prepared) and prescriptions approaching depletion. The dashboard SHALL show orders grouped by urgency: depleted prescriptions first, then approaching, then ok.

#### Scenario: Dashboard with mixed statuses
- **WHEN** personnel loads the dashboard and there are prescriptions in various states
- **THEN** the system displays depleted prescriptions first, then approaching, then ok

#### Scenario: Empty dashboard
- **WHEN** personnel loads the dashboard and no prescriptions exist
- **THEN** the system displays a message prompting to add patients and prescriptions

### Requirement: Automatic order creation
The system SHALL create an order for a prescription when the prescription enters the lookahead window and no active (pending or prepared) order exists for the current depletion cycle. Orders SHALL be created on-demand when the dashboard is loaded. Each order SHALL be tied to a specific depletion cycle via the cycle start date and estimated depletion date.

#### Scenario: Order created on dashboard load
- **WHEN** personnel loads the dashboard and a prescription has 5 days remaining with no active order
- **THEN** the system creates a new order with status "pending" tied to the current depletion cycle

#### Scenario: No duplicate order for same cycle
- **WHEN** personnel loads the dashboard and an active order already exists for the current depletion cycle
- **THEN** the system does not create a new order

#### Scenario: New order after refill
- **WHEN** a prescription has been refilled (new cycle started) and the new cycle enters the lookahead window
- **THEN** the system creates a new order for the new cycle, leaving the previous fulfilled order as history

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
The system SHALL allow personnel to filter the dashboard by prescription status (all, depleted, approaching, ok) and by date range.

#### Scenario: Filter by approaching status
- **WHEN** personnel selects the "approaching" filter
- **THEN** the system displays only prescriptions with "approaching" status

#### Scenario: Filter by date range
- **WHEN** personnel sets a date range filter
- **THEN** the system displays only prescriptions with estimated depletion within that range

### Requirement: Order status tracking
The system SHALL track an order status for each order through its lifecycle: "pending" (needs preparation), "prepared" (package is ready), "fulfilled" (picked up or shipped, terminal). Personnel SHALL be able to advance the status from the dashboard. Fulfilled is a terminal status — fulfilled orders remain as history.

#### Scenario: Mark order as prepared
- **WHEN** personnel clicks "mark as prepared" on a pending order
- **THEN** the system updates the order status to "prepared"

#### Scenario: Mark order as fulfilled
- **WHEN** personnel clicks "mark as fulfilled" on a prepared order
- **THEN** the system updates the order status to "fulfilled"

#### Scenario: Fulfilled order is terminal
- **WHEN** an order has been marked as fulfilled
- **THEN** the order remains in the system as history and cannot be changed

### Requirement: Dashboard filtering by order status
The system SHALL allow personnel to filter the dashboard by order status (all, pending, prepared, fulfilled) in addition to prescription status and date range filters.

#### Scenario: Filter by pending orders
- **WHEN** personnel selects the "pending" order status filter
- **THEN** the system displays only orders with "pending" status

### Requirement: Dashboard scoping
The dashboard SHALL only display orders belonging to the logged-in personnel's pharmacy.

#### Scenario: Multi-pharmacy isolation
- **WHEN** personnel from pharmacy A loads the dashboard
- **THEN** the system displays only orders for patients of pharmacy A
