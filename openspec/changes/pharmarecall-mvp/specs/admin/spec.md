## ADDED Requirements

### Requirement: Admin role
The system SHALL have an admin role, distinct from pharmacy personnel. Admin accounts SHALL be seeded via a CLI command or database migration, not through a registration form.

#### Scenario: Admin seeded at deployment
- **WHEN** the application is deployed for the first time
- **THEN** an admin account is created via a seed command with a configurable email and password

#### Scenario: Admin login
- **WHEN** an admin logs in with valid credentials
- **THEN** the system creates a session and redirects to the admin dashboard

### Requirement: Pharmacy creation by admin
The system SHALL allow admins to create a new pharmacy by providing a pharmacy name, address, phone number, and email. The system SHALL create the first personnel account (owner) as part of pharmacy creation.

#### Scenario: Admin creates a pharmacy
- **WHEN** an admin submits a valid pharmacy creation form with pharmacy details and owner credentials (name, email, temporary password)
- **THEN** the system creates the pharmacy record and an owner personnel account

#### Scenario: Duplicate pharmacy email
- **WHEN** an admin submits a pharmacy creation form with an email already associated with an existing pharmacy
- **THEN** the system displays an error and does not create the pharmacy

### Requirement: Pharmacy listing for admin
The system SHALL display a list of all pharmacies to the admin, showing name, address, number of personnel, and number of patients.

#### Scenario: Admin views pharmacy list
- **WHEN** an admin navigates to the pharmacy list
- **THEN** the system displays all pharmacies with their details

### Requirement: Admin route protection
Admin routes SHALL only be accessible to users with the admin role. Pharmacy personnel attempting to access admin routes SHALL be denied.

#### Scenario: Personnel accessing admin route
- **WHEN** a pharmacy personnel user attempts to access an admin-only route
- **THEN** the system denies access and displays an authorization error

#### Scenario: Unauthenticated access to admin route
- **WHEN** an unauthenticated user requests an admin route
- **THEN** the system redirects to the login page
