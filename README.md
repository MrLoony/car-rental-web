# Car Rental Web

Car Rental Web is a server-rendered Go web application for a car rental platform. It demonstrates a layered Go backend with PostgreSQL persistence, HTML templates, admin authentication, booking management, fleet management, image handling, email notifications, dashboard reporting, and practical security hardening.

The project is built as a coursework and portfolio application for a Junior Go Backend Developer profile.

## Tech Stack

- Go
- Chi Router
- HTML Templates
- Tailwind CSS
- JavaScript
- PostgreSQL
- Docker Compose

## Architecture

The application follows a simple layered architecture:

```text
Handler
Service
Repository
PostgreSQL
```

Project layout:

- `cmd/web` - application entry point
- `internal/config` - environment configuration
- `internal/database` - PostgreSQL connection setup
- `internal/handler` - routes, HTTP handlers, middleware, rendering
- `internal/service` - application and business logic
- `internal/repository` - database access
- `internal/model` - shared data models and form models
- `web/templates` - server-rendered HTML templates
- `web/static/css` - Tailwind source and generated application CSS
- `web/static/js/app.js` - browser JavaScript entry point
- `web/static/js/modules` - small progressive-enhancement modules
- `web/static/images` - static image assets
- `web/static/uploads` - local uploaded car images
- `migrations` - database schema and seed data

## Features

### Public Catalog

- Server-rendered car catalog at `/cars`
- Car detail pages at `/cars/{slug}`
- Text search, category, fuel type, transmission, and sort filters
- Server-side pagination that preserves filter query parameters
- Public catalog only shows cars that are active, available, and not archived
- Placeholder image fallback for missing or broken car images

### Booking Flow

- Public booking form from each car detail page
- Backend validation for customer details and rental dates
- 24-hour rental billing with minimum one billing day
- Estimated total calculation from stored car price
- Booking persistence in PostgreSQL
- Booking status lifecycle: `pending`, `confirmed`, `cancelled`, `completed`
- Availability conflict checks for pending and confirmed bookings
- 4-hour return/preparation buffer after each blocking booking
- Suggested next pickup time and alternative rental windows when a selected period is unavailable
- Similar available vehicle suggestions for booking conflicts
- Server-side booking prefill tokens for alternative vehicle booking forms

### Admin Panel

- Session-based admin login/logout
- Protected admin routes
- Admin dashboard at `/admin`
- Booking list, detail view, and status updates
- Car list, create, detail, edit, availability toggle, archive, and restore
- Admin search and filtering for cars and bookings
- Server-side pagination for admin cars and bookings
- Flash messages after admin actions
- Manual cleanup action for expired booking prefill tokens

### Admin Dashboard

The admin dashboard includes operational reporting cards and recent activity:

- Total bookings
- Pending bookings
- Confirmed bookings
- Cancelled bookings
- Completed bookings
- Total potential revenue
- Pending revenue
- Confirmed revenue
- Completed revenue
- Cancelled value
- Recent booking activity with booking detail links

### Booking CSV Export

Admins can export booking data as CSV from:

```text
/admin/bookings/export.csv
```

The export:

- requires admin authentication
- uses the same search and status filters as the booking list
- ignores pagination and exports the full filtered result set
- uses Go standard library CSV generation

Exported columns:

- ID
- Status
- Customer Name
- Customer Email
- Customer Phone
- Car
- Pickup At
- Return At
- Billing Days
- Estimated Total
- Created At

### Image Management

Cars can use either an external image URL or one uploaded local image.

Supported URL prefixes:

- `http://`
- `https://`
- `/static/`

Local uploads:

- saved under `web/static/uploads/cars`
- stored in PostgreSQL as `/static/uploads/cars/<filename>`
- ignored by Git, with `.gitkeep` preserving the upload directory
- limited to JPEG, PNG, and WebP files up to 5 MB
- validated by extension, detected content type, and WebP signature checks
- saved with application-generated filenames

The current implementation does not include image resizing, compression, antivirus scanning, or cloud/object storage.

### Email Notifications

Email delivery is configurable and disabled by default for local development.

Implemented notifications:

- admin notification when a new booking request is created
- customer notification when an admin changes booking status to `confirmed`, `cancelled`, or `completed`

Email sending is best-effort. Booking creation and booking status updates continue even if email delivery fails; failures are logged server-side.

### User Experience

- Custom 404 page
- Custom 500 page
- Flash messages for admin actions
- Clearer validation and upload error messages
- Responsive public, auth, and admin layouts
- Polished catalog, car detail, booking, dashboard, table, and admin form pages
- Server-rendered forms that work without JavaScript
- Progressive JavaScript enhancement for filters, dashboard controls, booking price preview, image previews, copy actions, toasts, themes, favorites, and form helpers

### Frontend

The frontend remains server-rendered and progressively enhanced. Core navigation, forms, filtering, pagination, booking submission, and admin actions work without JavaScript.

Frontend structure:

- Tailwind CSS source lives in `web/static/css/input.css`
- Generated CSS is served from `web/static/css/app.css`
- `web/static/js/theme-init.js` applies the saved color theme before the main stylesheet loads
- `web/static/js/app.js` initializes browser behavior
- feature modules live under `web/static/js/modules`

Current JavaScript modules:

- `catalog-filters.js` - public and admin filter form enhancements
- `booking-preview.js` - live booking estimate, duration buttons, date warnings, and suggested window fill-in
- `booking-wizard.js` - progressive multi-step booking form, review summary, and session draft restore
- `dashboard.js` - collapsible dashboard sections, recent activity filtering, and metric highlighting
- `favorites.js` - frontend-only favorite vehicles, catalog filtering, nav counter, and cross-tab synchronization
- `admin-tables.js` - visible-row filtering, copy buttons, row counts, and row highlighting
- `admin-actions.js` - centralized confirmation prompts for sensitive admin actions
- `image-preview.js` - image fallback handling, admin image previews, upload drop zone, slug preview, and price helper
- `car-detail.js` - public car detail copy action and single-image lightbox
- `theme.js` - Light, Dark, and System theme switching
- `toast.js` - reusable toast notifications with dismiss controls and progress indicators
- `form-helpers.js` - submit-once and unsaved-changes helpers
- `flash.js` - flash message focus support
- `utils.js` - shared DOM, debounce, formatting, and form helpers

The visual direction is a light, modern SaaS-style interface with a premium automotive feel: spacious layouts, white cards, subtle borders, clear typography, and restrained blue accents.

Frontend state storage is intentionally browser-local:

- `localStorage.carRentalTheme` stores the selected theme mode: `light`, `dark`, or `system`
- `localStorage.carRentalFavorites` stores frontend-only favorite car slugs
- `localStorage.carRentalFavoritesOnly` stores the catalog's client-side favorites filter preference
- `sessionStorage.carRentalBookingDraft:<car-slug>` stores temporary booking form draft values for the current browser session

These features are convenience enhancements. They do not replace server-side booking validation, authentication, persistence, or authorization.

### Security

The project includes a pragmatic security baseline:

- environment-aware `APP_ENV` configuration
- Secure session cookies in production
- `HttpOnly` and `SameSite=Lax` session cookies
- CSRF protection for POST forms
- basic in-memory login brute-force protection
- global security headers
- HSTS in production
- hardened image upload validation
- no internal error details exposed on custom 500 pages

This is not a claim of full production compliance, distributed rate limiting, WAF protection, malware scanning, or penetration-tested security.

## Configuration

Application settings are loaded from environment variables. `APP_ENV` controls development and production behavior:

```text
APP_ENV=development
APP_ENV=production
```

Empty or unknown values are treated as development.

Required session setting:

```text
SESSION_SECRET=change-me
```

Email settings:

```text
EMAIL_ENABLED=false
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=
SMTP_FROM_NAME="Car Rental Web"
ADMIN_NOTIFICATION_EMAIL=
```

When `EMAIL_ENABLED=true`, the application requires:

- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_FROM`
- `ADMIN_NOTIFICATION_EMAIL`

`SMTP_USERNAME` and `SMTP_PASSWORD` are optional to support development SMTP providers that do not require authentication.

## Development Setup

Install frontend dependencies:

```bash
npm install
```

Start PostgreSQL:

```bash
docker compose up -d
```

Run database migrations:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

Start Tailwind in watch mode:

```bash
npm run dev
```

In another terminal, start the Go server:

```bash
go run ./cmd/web
```

The app runs on:

```text
http://localhost:8080
```

Build CSS once:

```bash
npm run build
```

Run tests:

```bash
go test ./...
```

## Demo Admin Account

Seeded demo credentials:

```text
Email: admin@example.com
Password: admin123
```

These are demo credentials only. Change them before any real deployment, and use a strong `SESSION_SECRET`.

## Project Status

Completed:
- Public vehicle catalog
- Booking workflow
- Admin panel
- Email notifications
- Dashboard reporting
- Booking CSV export
- Security hardening
- UX improvements
- Frontend modernization
- JavaScript-heavy frontend enhancements

The application is suitable as a portfolio-scale SSR Go project. Additional production work would still be needed before real-world deployment.

Recent completed milestones include Stage 19 email notifications, Stage 20 UX and error handling, Stage 21 dashboard reporting and CSV export, Stage 22 frontend modernization, and Stage 23 JavaScript-heavy frontend features.

## Implemented

- Server-rendered Go web application
- Chi routing
- PostgreSQL integration with migrations
- Layered handler/service/repository architecture
- Public car catalog with filtering and pagination
- Public car detail and booking pages
- Booking validation, billing, persistence, and availability conflict checks
- Alternative date and vehicle suggestions on booking conflicts
- Secure server-side booking prefill tokens
- Admin authentication and protected admin routes
- Admin booking management
- Admin car management
- Car archive/restore workflow
- Local car image uploads and image URL support
- Admin dashboard with booking, revenue, and recent activity reporting
- Admin booking CSV export
- Email notification infrastructure and booking notifications
- Flash messages and polished form/admin feedback
- Custom 404 and 500 pages
- Modular ES module JavaScript architecture
- Tailwind CSS component layer
- Responsive public catalog, car detail, booking, auth, and admin pages
- Interactive admin dashboard, admin tables, and admin car forms
- Toast notification system
- Light, Dark, and System theme switcher
- Frontend-only favorites with localStorage persistence
- Progressive booking form wizard with sessionStorage draft restore
- Progressive frontend enhancements without SPA routing or AJAX requirements
- CSRF protection
- Basic login brute-force protection
- Security headers and environment-aware session cookies
- Health endpoint at `/health`

## Planned

- User registration
- Password reset
- Roles and permissions
- Payment integration
- Calendar-style availability view
- More advanced reporting
- Multiple image gallery
- Cloud/object storage for media
- Booking archive/delete workflow
- Production monitoring and operational hardening
