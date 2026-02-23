## ADDED Requirements

### Requirement: Print-friendly order view
The system SHALL provide a print-friendly page showing orders that need preparation. The print view SHALL use a clean layout optimized for paper printing (no navigation, no interactive elements).

#### Scenario: Print current orders
- **WHEN** personnel clicks "Print" from the dashboard
- **THEN** the system opens a print-friendly page with all currently visible dashboard entries

### Requirement: Print view content
The print view SHALL display: pharmacy name, print date, and a table with patient name, medication name, units per box, estimated depletion date, and status for each order.

#### Scenario: Print view table content
- **WHEN** the print view is rendered
- **THEN** it shows a table with columns for patient, medication, units/box, depletion date, and status

### Requirement: Print view filtering
The print view SHALL respect the same filters applied on the dashboard. If personnel filtered by "approaching" status, the print view SHALL only show approaching prescriptions.

#### Scenario: Print filtered view
- **WHEN** personnel filters the dashboard to "approaching" and clicks print
- **THEN** the print view shows only prescriptions with "approaching" status

### Requirement: Order label print
The system SHALL provide a printable label for any order, regardless of fulfillment method. The label SHALL include: patient name, medication name, and fulfillment method. For shipping orders, the label SHALL also include the delivery address. For pickup orders, the label SHALL also include the patient's phone or email for contact.

#### Scenario: Print label for a single order
- **WHEN** personnel clicks "print label" on a dashboard entry
- **THEN** the system opens a print-friendly page with a label for that order

#### Scenario: Batch print labels
- **WHEN** personnel clicks "print labels" from the dashboard
- **THEN** the system opens a print-friendly page with one label per visible order, laid out for standard label paper

#### Scenario: Shipping label content
- **WHEN** a label is printed for a shipping order
- **THEN** the label includes patient name, medication, fulfillment method, and delivery address

#### Scenario: Pickup label content
- **WHEN** a label is printed for a pickup order
- **THEN** the label includes patient name, medication, fulfillment method, and patient contact info

### Requirement: Print view includes fulfillment info
The print view SHALL display the fulfillment method (pickup/shipping) for each order entry.

#### Scenario: Print view shows fulfillment
- **WHEN** the order list print view is rendered
- **THEN** each entry shows whether it is pickup or shipping

### Requirement: Print via browser
The system SHALL trigger printing via the browser's native print dialog (`window.print()`). No server-side PDF generation is required for the MVP.

#### Scenario: Browser print dialog
- **WHEN** personnel clicks the print button
- **THEN** the browser's native print dialog opens with the print-friendly view
