## 1. Project Scaffolding

- [x] 1.1 Initialize Go module, set up `cmd/server/main.go` entrypoint with `net/http` server and graceful shutdown (`signal.NotifyContext` + `server.Shutdown`)
- [x] 1.2 Set up project directory structure (`internal/`, `db/migrations/`, `db/queries/`, `static/`)
- [x] 1.3 Add dependencies: pgx/v5 (pgxpool), sqlc, scs/v2, scs/pgxstore, goose/v3, templ, bcrypt, koanf/v2 (with TOML provider). Install Templ and sqlc CLIs as tool dependencies. Create `sqlc.yaml` config (engine: postgresql, queries: db/queries/, schema: db/migrations/, out: internal/db/, sql_package: pgx/v5).
- [x] 1.4 Set up Docker Compose with PostgreSQL 18
- [x] 1.5 Add oat.ink CSS to `static/` and configure `embed.FS` for static assets with an HTTP file server handler on `/static/`
- [x] 1.6 Create base Templ layout (html head with oat.ink, nav, content area) and a health check "/" page to verify the stack works
- [x] 1.7 Set up goose migrations with `embed.FS` and verify migrations run on startup
- [x] 1.8 Configure SCS session manager with pgxstore (with cleanup interval), wire into HTTP server middleware
- [x] 1.9 Wire `http.CrossOriginProtection` middleware into the HTTP server
- [x] 1.10 Set up koanf configuration loading from a TOML file (db connection, server port, session secret, lookahead window)

## 2. Login & Logout (Admin + Personnel)

- [x] 2.1 Migration: create `users` table (id, email, password_hash, name, role, pharmacy_id nullable, created_at, updated_at). Indexes: unique on email, index on pharmacy_id.
- [x] 2.2 Migration: create SCS sessions table (via pgxstore schema)
- [x] 2.3 Implement password hashing/verification with bcrypt
- [x] 2.4 Write sqlc queries for users: GetUserByEmail, GetUserByID
- [x] 2.5 Create login page Templ template: email/password form
- [x] 2.6 Implement login handler: validate credentials, create SCS session with user ID and role, redirect based on role (admin → admin dashboard, personnel → order dashboard)
- [x] 2.7 Implement auth middleware: load user from SCS session, attach to request context, redirect to login if unauthenticated
- [x] 2.8 Implement logout handler: destroy SCS session, redirect to login
- [x] 2.9 Add seed command/migration to create admin user with configurable email/password
- [x] 2.10 Create change-password page: current password + new password form
- [x] 2.11 Implement change-password handler: verify current password, hash new password, update user, redirect
- [x] 2.12 Verify end-to-end: seed admin, login, see a placeholder dashboard, change password, logout, login with new password

## 3. Admin Creates Pharmacy with Owner

- [ ] 3.1 Migration: create `pharmacies` table (id, name, address, phone, email, created_at, updated_at)
- [ ] 3.2 Write sqlc queries for pharmacies: CreatePharmacy, ListPharmacies (with personnel/patient counts), GetPharmacyByID, UpdatePharmacy
- [ ] 3.3 Write sqlc queries for users: CreateUser, ListUsersByPharmacy, UpdateUserPassword
- [ ] 3.4 Implement role-based middleware: require admin role
- [ ] 3.5 Create admin dashboard page listing all pharmacies (name, address, personnel count, patient count)
- [ ] 3.6 Create pharmacy creation form: pharmacy details + owner name, email, temporary password
- [ ] 3.7 Implement pharmacy creation handler: validate, create pharmacy + owner user in a transaction, redirect to admin dashboard
- [ ] 3.8 Create pharmacy detail/edit page for admin with edit form
- [ ] 3.9 Implement pharmacy update handler
- [ ] 3.10 Create admin view of pharmacy personnel on pharmacy detail page
- [ ] 3.11 Create add-personnel-to-pharmacy form for admin (name, email, temporary password)
- [ ] 3.12 Implement admin add-personnel handler: create user scoped to the selected pharmacy
- [ ] 3.13 Verify end-to-end: admin logs in, sees pharmacy list, creates pharmacy with owner, edits pharmacy, adds personnel, owner and personnel can log in

## 4. Owner Manages Personnel

- [ ] 4.1 Implement owner-only middleware
- [ ] 4.2 Create personnel list page for pharmacy owners (shows personnel in their pharmacy)
- [ ] 4.3 Create add-personnel form: name, email, temporary password
- [ ] 4.4 Implement add-personnel handler: validate, create user with personnel role scoped to owner's pharmacy
- [ ] 4.5 Verify end-to-end: owner logs in, sees personnel list, adds personnel, new personnel can log in

## 5. Patient Onboarding

- [ ] 5.1 Migration: create `patients` table (id, pharmacy_id, first_name, last_name, phone, email, delivery_address, fulfillment, notes, consensus, consensus_date, created_at, updated_at). Index on pharmacy_id.
- [ ] 5.2 Write sqlc queries for patients: CreatePatient, ListPatientsByPharmacy, GetPatientByID, UpdatePatient, SetPatientConsensus
- [ ] 5.3 Create patient list page (name, contact info, consensus status, prescription count)
- [ ] 5.4 Create patient creation form: first name, last name, phone, email, delivery address, fulfillment preference, notes
- [ ] 5.5 Implement patient creation handler with validation (require name + at least one contact, shipping requires address)
- [ ] 5.6 Create patient detail/edit page with edit form
- [ ] 5.7 Implement patient update handler with validation
- [ ] 5.8 Implement consensus recording on patient detail page: button to mark active with current date
- [ ] 5.9 Add pharmacy scoping to all patient queries (filter by logged-in user's pharmacy_id)
- [ ] 5.10 Verify end-to-end: personnel logs in, creates patient, records consensus, edits patient, sees patient list

## 6. Recurring Prescriptions & Refills

- [ ] 6.1 Migration: create `prescriptions` table (id, patient_id, medication_name, units_per_box, daily_consumption, box_start_date, created_at, updated_at). Index on patient_id.
- [ ] 6.2 Migration: create `refill_history` table (id, prescription_id, box_start_date, box_end_date, created_at)
- [ ] 6.3 Write sqlc queries for prescriptions: CreatePrescription, ListPrescriptionsByPatient, GetPrescriptionByID, UpdatePrescription, InsertRefillHistory
- [ ] 6.4 Implement depletion calculation: box_start_date + floor(units_per_box / daily_consumption) days
- [ ] 6.5 Implement prescription status classification: ok (>7 days), approaching (<=7 days), depleted (<=0 days)
- [ ] 6.6 Create prescription creation form on patient detail page: medication name, units per box, daily consumption, box start date
- [ ] 6.7 Implement prescription creation handler (block if patient has no consensus)
- [ ] 6.8 Create prescription edit form and update handler
- [ ] 6.9 Implement refill recording: update box_start_date, insert previous period into refill_history, auto-fulfill any active order for the previous cycle
- [ ] 6.10 Display prescription list on patient detail page with status badges, estimated depletion, and refill button
- [ ] 6.11 Verify end-to-end: add prescription to patient, see depletion calculation, edit prescription, record refill, see history

## 7. Order Dashboard

- [ ] 7.1 Migration: create `orders` table (id, prescription_id, cycle_start_date, estimated_depletion_date, status, created_at, updated_at). Indexes: on prescription_id, on status.
- [ ] 7.2 Write sqlc queries for orders: CreateOrder, GetActiveOrderByPrescription, UpdateOrderStatus, ListDashboardOrders (join prescriptions + patients, filter by pharmacy)
- [ ] 7.3 Implement automatic order creation: on dashboard load, create pending orders for prescriptions in the lookahead window that have no active (pending/prepared) order for the current cycle
- [ ] 7.4 Create dashboard page with entries grouped by prescription status (depleted first, then approaching, then ok)
- [ ] 7.5 Display per entry: patient name, medication, depletion date, days remaining, status badge, fulfillment method, order status
- [ ] 7.6 Implement prescription status filter (all, depleted, approaching, ok)
- [ ] 7.7 Implement date range filter
- [ ] 7.8 Implement order status filter (all, pending, prepared, fulfilled)
- [ ] 7.9 Implement order status advancement from dashboard: pending → prepared → fulfilled (form POST)
- [ ] 7.10 Implement configurable lookahead window (default 7 days)
- [ ] 7.11 Verify end-to-end: personnel sees dashboard, orders auto-created, filters work, advances order through full lifecycle

## 8. Personnel Notifications

- [ ] 8.1 Migration: create `notifications` table (id, pharmacy_id, prescription_id, transition_type, read, created_at). Indexes: on (pharmacy_id, read), unique on (prescription_id, transition_type).
- [ ] 8.2 Write sqlc queries for notifications: CreateNotification (with ON CONFLICT DO NOTHING), ListNotificationsByPharmacy (join prescriptions + patients), MarkNotificationRead, MarkAllNotificationsRead, CountUnreadNotifications
- [ ] 8.3 Implement notification generation: on dashboard load, detect prescriptions entering "approaching" status, create notifications if not already created for that transition
- [ ] 8.4 Add unread notification count to base layout nav (loaded via middleware or Templ helper on every page)
- [ ] 8.5 Create notification list page: sorted by date, showing patient name, medication, depletion date, read/unread state
- [ ] 8.6 Implement mark-as-read handler (single notification and mark-all)
- [ ] 8.7 Add pharmacy scoping to all notification queries
- [ ] 8.8 Verify end-to-end: prescription approaches depletion, notification appears in nav badge, view list, mark as read

## 9. Print Views & Labels

- [ ] 9.1 Create print-friendly order list page (no nav, clean table layout, print-specific CSS)
- [ ] 9.2 Include in print view: pharmacy name, print date, patient, medication, units/box, depletion date, status, fulfillment method
- [ ] 9.3 Pass dashboard filters through to print view via query parameters
- [ ] 9.4 Create order label Templ template: patient name, medication, fulfillment method, address (shipping) or contact info (pickup)
- [ ] 9.5 Implement single label print from dashboard entry
- [ ] 9.6 Implement batch label print for all visible orders, laid out for label paper
- [ ] 9.7 Add `window.print()` trigger on print pages
- [ ] 9.8 Verify end-to-end: filter dashboard, print order list, print single label, batch print labels
