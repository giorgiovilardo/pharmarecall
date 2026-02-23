## Why

Italian pharmacies have no way to track patients with recurring prescriptions (e.g., diabetes supplies). They cannot proactively prepare packages, monitor refill timelines, or remind patients to pick up their medication — leading to missed refills and manual overhead.

## What Changes

- Add admin role for pharmacy onboarding
- Add pharmacy and pharmacy personnel registration and authentication
- Add patient onboarding with consensus tracking
- Add recurring prescription management with consumption and quantity tracking
- Add an order dashboard that calculates which orders need to be prepared each day
- Add pharmacy personnel notifications for upcoming orders
- Add printable order views

## Capabilities

### New Capabilities

- `admin`: Admin role for creating and managing pharmacies
- `pharmacy-auth`: Pharmacy personnel authentication and multi-personnel management
- `patient-management`: Patient onboarding, consensus tracking, and patient data management
- `recurring-prescriptions`: Recurring prescription creation with consumption rate, box quantity, and refill cycle tracking
- `order-dashboard`: Daily order calculation, visualization, and filtering for pharmacy personnel
- `personnel-notifications`: Notifications to pharmacy personnel about upcoming orders that need preparation
- `print-orders`: Printable order views for pharmacy operations

### Modified Capabilities

_None — greenfield project._

## Impact

- **New Go service**: Web application with HTTP API and server-rendered views
- **Database**: PostgreSQL schema for pharmacies, personnel, patients, prescriptions, and orders
- **Infrastructure**: Dockerized PostgreSQL for development; single Go binary deployment for production
- **External dependencies**: None for MVP (no ERP integration, no payment processing, no patient-facing access)
