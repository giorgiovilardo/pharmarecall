## ADDED Requirements

### Requirement: Personnel login
The system SHALL allow pharmacy personnel to log in with email and password. The system SHALL create a server-side session stored in PostgreSQL and set a secure HTTP-only cookie.

#### Scenario: Successful login
- **WHEN** personnel submits valid email and password
- **THEN** the system creates a session, sets a session cookie, and redirects to the dashboard

#### Scenario: Invalid credentials
- **WHEN** personnel submits an incorrect email or password
- **THEN** the system displays a generic "invalid credentials" error without revealing which field is wrong

### Requirement: Personnel logout
The system SHALL allow logged-in personnel to log out, destroying the server-side session.

#### Scenario: Successful logout
- **WHEN** logged-in personnel clicks logout
- **THEN** the system destroys the session, clears the cookie, and redirects to the login page

### Requirement: Personnel management
The system SHALL allow the pharmacy owner to invite additional personnel by creating accounts with email and a temporary password. Each personnel account SHALL be scoped to a single pharmacy.

#### Scenario: Owner adds new personnel
- **WHEN** the pharmacy owner submits a new personnel form with name, email, and temporary password
- **THEN** the system creates the personnel account associated with the owner's pharmacy

#### Scenario: Non-owner attempts to add personnel
- **WHEN** a non-owner personnel attempts to access the personnel management page
- **THEN** the system denies access and displays an authorization error

### Requirement: Password change
The system SHALL allow any authenticated user (admin, owner, or personnel) to change their own password by providing their current password and a new password.

#### Scenario: Successful password change
- **WHEN** a user submits their correct current password and a new password
- **THEN** the system updates the password hash and redirects to the dashboard

#### Scenario: Incorrect current password
- **WHEN** a user submits an incorrect current password
- **THEN** the system displays an error and does not change the password

### Requirement: Password hashing
The system SHALL store passwords using bcrypt hashing. The system SHALL NOT store plaintext passwords.

#### Scenario: Password stored securely
- **WHEN** a personnel account is created or password is changed
- **THEN** the system stores only the bcrypt hash of the password

### Requirement: Route protection
The system SHALL require authentication for all routes except login and static assets. Unauthenticated requests SHALL be redirected to the login page.

#### Scenario: Unauthenticated access to protected route
- **WHEN** an unauthenticated user requests a protected page
- **THEN** the system redirects to the login page

#### Scenario: Authenticated access to protected route
- **WHEN** an authenticated user requests a protected page
- **THEN** the system serves the requested page
