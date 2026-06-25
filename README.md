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
- Catalog cards use the primary gallery image first, then the first gallery image, then a placeholder
- Equal-height responsive catalog cards with stable image areas, clamped titles, aligned specs, and aligned CTAs
- Multi-image gallery carousel on car detail pages when gallery images are configured
- Gallery carousel thumbnails, previous/next controls, image counter, keyboard navigation, and lightbox preview
- Text search, category, fuel type, transmission, and sort filters
- Server-side pagination that preserves filter query parameters
- Public catalog only shows cars that are active, available, and not archived
- Placeholder image fallback when no gallery image is configured

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
- Printable booking summaries from the admin booking detail page
- Car list, create, detail, edit, availability toggle, archive, and restore
- Gallery image management from the car edit page
- Gallery image URL add, multi-file local gallery upload, primary image selection, and gallery image deletion
- Gallery actions return directly to the gallery section after add, primary, or delete operations
- Admin search and filtering for cars and bookings
- Server-side pagination for admin cars and bookings
- Flash messages after admin actions
- Manual cleanup action for expired booking prefill tokens

### Admin Dashboard

The admin dashboard includes operational reporting cards and recent activity:

- Reporting ranges: All Time, Last 30 Days, and This Month
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

Dashboard range filtering is SSR-based through the `/admin?range=` query parameter. It filters booking statistics, revenue statistics, and recent activity by booking creation date without AJAX or chart libraries.

### Printable Booking Summary

Admins can print a clean booking summary from the admin booking detail page. The feature uses browser printing and print-specific CSS rather than PDF generation or external libraries.

The print layout focuses on the booking document and hides unrelated interface chrome such as navigation, action buttons, forms, flash messages, and toast regions.

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

Vehicle images are managed exclusively through galleries. Each car can have multiple gallery images, and the public catalog/detail pages use the gallery as the only vehicle image source.

Supported URL prefixes:

- `http://`
- `https://`
- `/static/`

Gallery uploads:

- saved under `web/static/uploads/cars`
- stored in PostgreSQL as `/static/uploads/cars/<filename>`
- ignored by Git by default, with `.gitkeep` preserving the upload directory
- limited to JPEG, PNG, and WebP files up to 5 MB
- validated by extension, detected content type, and WebP signature checks
- saved with application-generated filenames

Demo gallery assets for deployment are committed separately under `web/static/uploads/cars/demo/` and referenced as `/static/uploads/cars/demo/<filename>`.

Gallery images:

- are managed from the admin car edit page
- can be added from an external URL or multiple local uploads in one request
- include optional alt text
- can be marked as the primary gallery image
- can be deleted by admins
- use JPEG, PNG, WebP, and 5 MB validation rules for local files
- automatically make the first gallery image primary when a car has no existing gallery images
- keep the existing primary image when more images are added later
- are used first on catalog cards and public car detail pages when present

Catalog cards and public detail pages resolve images in this order:

1. primary gallery image
2. first gallery image
3. placeholder image

The current implementation does not include drag-and-drop gallery ordering, image resizing, compression, antivirus scanning, or cloud/object storage.

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
- Skip link for keyboard users
- Improved focus-visible states across navigation, forms, tables, and interactive controls
- Improved ARIA labels for theme, favorites, table actions, toasts, booking wizard controls, and image lightbox interactions
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
- `image-preview.js` - image fallback handling, admin slug preview, and price helper
- `car-gallery.js` - public car detail gallery carousel, thumbnail selection, previous/next controls, image counter, keyboard navigation, and lightbox image synchronization
- `car-detail.js` - public car detail copy action and selected-image lightbox
- `theme.js` - Light, Dark, and System theme switching
- `toast.js` - reusable toast notifications with dismiss controls and progress indicators
- `print-summary.js` - progressive browser printing for admin booking summaries
- `form-helpers.js` - submit-once and unsaved-changes helpers
- `flash.js` - flash message focus support
- `utils.js` - shared DOM, debounce, formatting, and form helpers

The visual direction is a light, modern SaaS-style interface with a premium automotive feel: spacious layouts, white cards, subtle borders, clear typography, and restrained blue accents.

Frontend state storage is intentionally browser-local:

- `localStorage.carRentalTheme` stores the selected theme mode: `light`, `dark`, or `system`
- `localStorage.carRentalFavorites` stores frontend-only favorite car slugs
- the catalog favorites-only view uses a `favorites` query parameter generated from browser-local favorite slugs
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

Core settings:

```text
PORT=10000
APP_PORT=8080
BASE_URL=https://your-app.example.com
DATABASE_URL=postgres://...
SESSION_SECRET=change-me
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=change-me
```

`PORT` is used by hosted environments such as Render. `APP_PORT` is kept as a local fallback when `PORT` is not set.

Email settings:

```text
EMAIL_ENABLED=false
EMAIL_PROVIDER=smtp
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=
SMTP_FROM_NAME="Car Rental Web"
BREVO_API_KEY=
BREVO_FROM_EMAIL=
BREVO_FROM_NAME="Car Rental Web"
ADMIN_NOTIFICATION_EMAIL=
```

When `EMAIL_ENABLED=true`, `EMAIL_PROVIDER` selects the delivery method. If it is omitted, the application defaults to `smtp`.

SMTP remains supported for local development and compatible providers. With `EMAIL_PROVIDER=smtp`, the application requires:

- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_FROM`
- `ADMIN_NOTIFICATION_EMAIL`

`SMTP_USERNAME` and `SMTP_PASSWORD` are optional to support development SMTP providers that do not require authentication.

Brevo Transactional Email over HTTPS is recommended for Render free deployments where SMTP ports can time out. With `EMAIL_PROVIDER=brevo`, the application requires:

- `BREVO_API_KEY`
- `BREVO_FROM_EMAIL`
- `BREVO_FROM_NAME`
- `ADMIN_NOTIFICATION_EMAIL`

The Brevo sender uses the HTTPS API endpoint rather than SMTP ports. Configure API keys only in the hosting provider environment.

When `APP_ENV=production`, the application fails fast unless deployment-critical values are configured safely:

- `DATABASE_URL`
- `BASE_URL`
- strong non-default `SESSION_SECRET`
- configured `ADMIN_EMAIL`
- non-demo `ADMIN_PASSWORD`

In production, the configured admin account is created or updated at startup. If the configured email is different from the seeded demo admin email, the seeded demo admin account is removed. Secrets must be configured in the hosting provider environment and must not be committed.

## Production Deployment

The intended demo deployment target is a Render Web Service with an external PostgreSQL database such as Neon.

Recommended Render settings:

```text
Build command: npm install && npm run build && go build -o app ./cmd/web
Start command: ./app
```

Run migrations against the production database before starting or as a deploy step:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

Required production environment variables:

```text
APP_ENV=production
PORT=<provided by Render>
BASE_URL=https://your-render-service.onrender.com
DATABASE_URL=<Neon PostgreSQL connection string>
SESSION_SECRET=<strong random secret>
ADMIN_EMAIL=<admin login email>
ADMIN_PASSWORD=<strong admin password>
EMAIL_ENABLED=false
EMAIL_PROVIDER=smtp
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=
SMTP_FROM_NAME="Car Rental Web"
BREVO_API_KEY=
BREVO_FROM_EMAIL=
BREVO_FROM_NAME="Car Rental Web"
ADMIN_NOTIFICATION_EMAIL=
```

For Render email notifications with Brevo:

```text
EMAIL_ENABLED=true
EMAIL_PROVIDER=brevo
BREVO_API_KEY=<Brevo API key>
BREVO_FROM_EMAIL=<verified sender email>
BREVO_FROM_NAME="Car Rental Web"
ADMIN_NOTIFICATION_EMAIL=<admin notification recipient>
```

Brevo uses HTTPS API requests, so it avoids SMTP port timeouts on hosts where outbound SMTP is blocked or unreliable.

Local gallery uploads are stored under `web/static/uploads/cars`. The repository ignores runtime uploads by default and commits only curated demo images under `web/static/uploads/cars/demo/`; on free Render services, runtime filesystem uploads are not persistent across redeploys. Committed demo images are intentional for defense/demo data, while a real production deployment should use object storage or persistent disk for media.

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
Production deployments should use `ADMIN_EMAIL` and `ADMIN_PASSWORD` environment variables so the startup bootstrap replaces the demo password or removes the seeded demo account.

## Project Status

Completed:
- Public vehicle catalog
- Booking workflow
- Admin panel
- Email notifications
- Dashboard reporting
- Booking CSV export
- Printable booking summaries
- Dashboard date range filtering
- Security hardening
- UX improvements
- Frontend modernization
- JavaScript-heavy frontend enhancements
- Multi-image vehicle galleries
- Production deployment readiness foundation
- Accessibility improvements

The application is suitable as a portfolio-scale SSR Go project. Additional production work would still be needed before real-world deployment.

Recent completed milestones include Stage 19 email notifications, Stage 20 UX and error handling, Stage 21 dashboard reporting and CSV export, Stage 22 frontend modernization, Stage 23 JavaScript-heavy frontend features, Stage 24 printing, dashboard range filters, and accessibility polish, Stage 25 multi-image vehicle galleries, and Stage 26 gallery-only image management.

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
- Gallery-only vehicle image management
- Local gallery image uploads and URL-based gallery images
- Gallery-first catalog and car detail image resolution
- Public car detail carousel with gallery and placeholder fallbacks
- Equal-height public catalog cards with clamped titles and aligned actions
- Multi-file local gallery uploads with gallery-section redirects after image actions
- Admin dashboard with booking, revenue, and recent activity reporting
- Dashboard reporting range filters for All Time, Last 30 Days, and This Month
- Admin booking CSV export
- Printable admin booking summaries using browser print CSS
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
- Practical accessibility improvements for keyboard navigation, focus states, ARIA labels, toasts, tables, wizard steps, and lightbox focus restoration
- Progressive frontend enhancements without SPA routing or AJAX requirements
- Render/Neon-oriented production configuration and deployment documentation
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
- Drag-and-drop gallery ordering
- Image resizing and compression
- Cloud/object storage for media
- Booking archive/delete workflow
- Production monitoring and operational hardening
