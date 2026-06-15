# Car Rental Web

Car Rental Web is a server-rendered Go web application for a car rental platform. The project currently includes the application foundation, PostgreSQL connection, database migrations, seeded demo catalog data, Tailwind-based catalog pages, query-parameter catalog filtering, server-side pagination, a booking request flow, availability validation, protected admin authentication, admin booking management, admin fleet management, basic car image management, and several security hardening measures.

Notifications, advanced media management, and broader production operations hardening are planned but not implemented yet.

## Tech Stack

- Go
- Chi Router
- HTML Templates
- Tailwind CSS
- JavaScript
- PostgreSQL

## Project Structure

- `cmd/web` - application entry point
- `internal/config` - environment configuration loading
- `internal/database` - PostgreSQL connection setup
- `internal/handler` - routes, handlers, and template rendering
- `internal/repository` - database access layer
- `internal/service` - application service layer
- `web/templates` - server-rendered HTML templates
- `web/static` - CSS, JavaScript, and image assets
- `migrations` - database schema and demo data migrations

## Development Setup

Install frontend dependencies:

```bash
npm install
```

Start Tailwind in watch mode:

```bash
npm run dev
```

In another terminal, start the Go server:

```bash
go run ./cmd/web
```

The app runs on `http://localhost:8080` by default.

To build CSS once:

```bash
npm run build
```

## Configuration

Application configuration is loaded from environment variables, with local defaults for development. `APP_ENV` controls the environment mode:

```text
APP_ENV=development
APP_ENV=production
```

Empty or unknown `APP_ENV` values are treated as `development`.

### Email Configuration

Email notification settings are available in configuration. Notifications are disabled by default:

```text
EMAIL_ENABLED=false
```

`EMAIL_ENABLED=false` is suitable for local development and allows SMTP fields to stay empty. When `EMAIL_ENABLED=true`, the application validates that the required email settings are present:

- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_FROM`
- `ADMIN_NOTIFICATION_EMAIL`

`SMTP_PORT` defaults to `587`, and `SMTP_FROM_NAME` defaults to `Car Rental Web`. `SMTP_USERNAME` and `SMTP_PASSWORD` are optional for now so local or development SMTP providers without authentication can be configured later.

The email sender foundation includes a no-op sender for local development and an SMTP sender for notification stages. Reusable admin booking-created and customer booking-status email templates also exist. New booking requests can notify the administrator by email when email is enabled. When an admin changes a booking status to `confirmed`, `cancelled`, or `completed`, the customer can be notified by email. Status changes back to `pending` do not send a customer email. Booking creation and admin status updates do not fail if email sending fails; the error is logged and the normal flow continues.

### Email Notifications Manual Test

For normal local development, keep email disabled:

```text
EMAIL_ENABLED=false
```

To manually verify SMTP delivery, use a sandbox SMTP provider such as Mailtrap or a similar testing service, or provider-specific SMTP credentials such as a Gmail app password if appropriate. Never commit real SMTP credentials.

Set the required values in your local `.env`:

```text
EMAIL_ENABLED=true
SMTP_HOST=smtp.example.test
SMTP_PORT=587
SMTP_USERNAME=your-smtp-username
SMTP_PASSWORD=your-smtp-password
SMTP_FROM=no-reply@example.test
SMTP_FROM_NAME="Car Rental Web"
ADMIN_NOTIFICATION_EMAIL=admin@example.test
```

Admin notification smoke test:

1. Start the app with SMTP enabled.
2. Create a public booking request from a car booking page.
3. Confirm the booking request succeeds and redirects normally.
4. Confirm the administrator receives the new booking email.
5. Check the server logs for SMTP errors.

Customer status notification smoke test:

1. Log in as admin.
2. Open an existing booking request.
3. Change the booking status to `confirmed`, `cancelled`, or `completed`.
4. Confirm the status update succeeds.
5. Confirm the customer receives the status-change email.
6. Temporarily disable or break SMTP settings and repeat a status update to confirm the status still changes while the email failure is only logged.

Email delivery is best-effort. A failed admin notification does not fail booking creation, and a failed customer notification does not rollback a booking status update. Failures are logged server-side for manual inspection.

## Database Setup

Start PostgreSQL with Docker Compose:

```bash
docker compose up -d
```

Stop PostgreSQL:

```bash
docker compose down
```

## Migrations

Start PostgreSQL before running migrations:

```bash
docker compose up -d
```

Run migrations using `DATABASE_URL` from `.env`:

```bash
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" version
```

## Catalog

After starting PostgreSQL and running migrations, open the demo catalog:

```text
http://localhost:8080/cars
```

The catalog supports server-rendered filtering through URL query parameters:

```text
http://localhost:8080/cars?category=suv&sort=price_desc
```

Available filters include text search, category, fuel type, transmission, and sort order. The catalog uses server-side pagination, and pagination links preserve the active filter and sort query parameters. Invalid or out-of-range page values fall back safely instead of breaking the page.

The filter form works without JavaScript; JavaScript only enhances the experience with debounced search and automatic submit for select fields.

Cars archived by an admin are removed from the public customer-facing fleet. Archived cars do not appear in the public catalog, public car detail pages return not found, and archived cars cannot be booked from public booking URLs.

## Bookings

Booking requests can be started from a car details page:

```text
http://localhost:8080/cars/toyota-corolla/book
```

The booking form is server-rendered and works without JavaScript. Backend validation is the source of truth. Rentals are billed in 24-hour periods using `ceil(duration_hours / 24)`, with a minimum of one billing day. The estimated total is calculated as `billing_days * price_per_day`, and new booking requests are saved with `pending` status.

Availability is checked before a booking request is saved. `pending` and `confirmed` bookings block overlapping requests, while `cancelled` and `completed` bookings do not. After a return time, the car remains unavailable for 4 hours to allow for possible late return, cleaning, washing, inspection, and preparation.

When a selected period is unavailable, the form shows the nearest available pickup time when one can be calculated. It can also show up to three alternative rental windows that fit the same requested rental duration. Suggested windows respect blocking bookings, non-blocking statuses, and the 4-hour return/preparation buffer. Each suggestion shows the start time, end time, billing days, and estimated total.

The form can also suggest similar available vehicles when the selected car is unavailable. Alternative vehicles are selected from the same category, within a similar price range of roughly 20% above or below the selected car, and only when they are available for the selected pickup and return period. Pending and confirmed bookings block alternatives, cancelled and completed bookings do not, archived cars are excluded, and the 4-hour return/preparation buffer is respected.

Suggested vehicle cards show car information, price per day, billing days, estimated total, a link to view the car, and a `Book this car` link. When the user clicks `Book this car`, the alternative vehicle booking form opens with the previously entered name, email, phone, pickup time, return time, and message restored automatically.

This carry-over uses a server-side prefill token. Alternative booking URLs use the format `/cars/{slug}/book?prefill=<token>`, so customer details and selected times are not exposed in the URL. Prefill tokens are URL-safe, expire after 30 minutes, and invalid or expired tokens safely fall back to a normal empty booking form. Suggestions are not AJAX/live or a full recommendation engine.

Booking records are currently managed through statuses. Cancelled and completed bookings remain as history. Physical deletion or archive handling for booking history is not implemented yet.

## Admin

The admin dashboard is available at:

```text
http://localhost:8080/admin
```

Current admin pages include a booking requests list, booking detail page, booking status update form, car list, car create form, car detail page, car edit form, availability toggle, and car archive/restore actions. Supported booking statuses are `pending`, `confirmed`, `cancelled`, and `completed`. Admin routes are protected with session-based authentication.

The admin cars and admin bookings lists use server-side pagination. Pagination preserves active filters, so filtered URLs remain shareable and bookmarkable. Admin tables also include horizontal overflow handling, compact spacing, status badges, and truncation for long customer, email, slug, and car text so the pages remain usable on narrower screens.

Admin cars can be searched by brand, model, slug, category, fuel type, and transmission. They can also be filtered by availability: `all`, `available`, or `unavailable`.

Admin bookings can be searched by customer name, customer email, customer phone, car brand, car model, and car slug. They can also be filtered by status: `all`, `pending`, `confirmed`, `cancelled`, or `completed`.

Admin filter forms are SSR-first GET forms and work without JavaScript. JavaScript only improves the experience with debounced search input and automatic select submission; filtering is not AJAX or SPA-based.

Admin routes require login:

- `/admin`
- `/admin/bookings`
- `/admin/bookings/{id}`
- `/admin/cars`
- `/admin/cars/new`
- `/admin/cars/{id}`
- `/admin/cars/{id}/edit`

Public customer pages remain accessible without login:

- `/`
- `/cars`
- `/cars/{slug}`
- `/cars/{slug}/book`

Admin car management is backed by PostgreSQL. Created and edited cars are reflected in the public catalog when they are marked available. Cars marked unavailable remain visible in admin but are hidden from public catalog results.

Admin car management includes create/edit forms, image URL management, local image upload, image preview, availability toggles, and archive/restore actions. Admin pages display active/archived status. Archived cars remain visible and manageable in admin, but they are removed from public catalog, public detail, public booking, and alternative vehicle suggestion flows. Archiving a car sets `archived_at` and marks it unavailable. Restoring a car clears `archived_at` but does not automatically make it publicly available; an admin must explicitly enable availability after restore. Availability controls are hidden for archived cars until they are restored.

The admin dashboard includes a manual maintenance action for expired booking prefill tokens. `POST /admin/cleanup/prefills` requires admin authentication and CSRF protection, deletes expired temporary booking form state records, and redirects back to the dashboard. There is no background scheduler, cron job, or automatic startup cleanup.

## Image Management

Cars can use an external image URL or a local uploaded image. Image URLs must start with one of:

- `http://`
- `https://`
- `/static/`

Admin car forms show a live image preview while editing the image URL. Public catalog and car detail pages display car images with a placeholder fallback when no image is set or an image cannot be loaded.

Uploaded car images are saved locally under:

```text
web/static/uploads/cars
```

The public path stored in PostgreSQL looks like:

```text
/static/uploads/cars/<filename>
```

Uploaded files are ignored by Git. The `.gitkeep` file keeps the upload folder structure in the repository.

Upload validation is intentionally conservative. Local car image uploads allow JPEG, PNG, and WebP files up to 5 MB. The application checks the filename extension, sniffs the uploaded content, verifies that the extension and detected content type match, rejects empty files, and performs a WebP signature check. Uploaded filenames are generated by the application, and upload paths are constrained to the local car upload directory to prevent path traversal.

The current media implementation does not include antivirus scanning, image re-encoding/compression, malware scanning, or cloud/object storage.

## Authentication

Admin authentication is implemented with a login page, bcrypt password verification, and cookie-based sessions using `gorilla/sessions`. Logging out clears the admin session.

Session cookies are configured with `HttpOnly=true` and `SameSite=Lax`. In development, cookies use `Secure=false` so local HTTP works. In production mode, cookies use `Secure=true`.

Admin login attempts have a basic in-memory per-email limiter. After 5 failed attempts within a 15-minute window, that email is temporarily locked for 10 minutes. This is suitable for a simple single-instance application, but it is not distributed or Redis-backed.

The required session secret is configured with:

```text
SESSION_SECRET
```

Do not use the default development value in production.

Demo admin credentials:

```text
Email: admin@example.com
Password: admin123
```

These are demo credentials only. Change them before any production use, and also change `SESSION_SECRET`.

## Security

The application includes a pragmatic security baseline for the current SSR implementation:

- `APP_ENV` controls development/production behavior.
- Production mode enables Secure session cookies.
- Session cookies use `HttpOnly` and `SameSite=Lax`.
- CSRF protection is enabled for POST forms with hidden `csrf_token` fields.
- Missing or invalid CSRF tokens return `403 Forbidden`.
- Admin login has basic in-memory brute-force protection.
- Security headers are applied globally.
- HSTS is enabled only in production.
- Car image uploads are validated with both extension and content checks.

Security headers include `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, and a Content Security Policy designed to work with the server-rendered pages, local static assets, uploaded images, and HTTPS image URLs.

This is not intended to claim complete production compliance, distributed rate limiting, WAF protection, malware scanning, or penetration-tested security.

## Implemented

- Server-rendered Go application foundation
- Chi router setup
- Template rendering with a base layout and home page
- Tailwind CSS build/watch workflow
- Static file serving
- Environment configuration loading
- Environment-aware security configuration
- Email notification configuration foundation
- Email sender foundation with no-op and SMTP implementations
- Email notification template/content foundation
- Admin email notification attempt for new booking requests
- Customer email notification attempt for booking status changes
- Email notification manual SMTP smoke-test documentation
- PostgreSQL integration with `pgxpool`
- Docker Compose database setup
- Database migrations
- Seeded demo car catalog data
- Repository, service, and handler flow for cars
- Cars catalog page at `/cars`
- Car details page at `/cars/{slug}`
- Catalog text search
- Category, fuel, and transmission filters
- Price sorting
- Query-parameter based catalog filtering
- Public catalog pagination
- Admin cars pagination
- Admin bookings pagination
- Admin car search
- Admin car availability filtering
- Admin booking search
- Admin booking status filtering
- Responsive admin table overflow handling
- Progressive JavaScript enhancement for admin filters
- Progressive JavaScript enhancement with no-JS fallback
- Booking request form
- Backend booking validation
- 24-hour rental billing calculation
- Estimated total calculation
- Booking persistence in PostgreSQL
- Booking success page
- JavaScript booking price preview
- Availability conflict checking
- 4-hour return/preparation buffer
- Nearest available pickup suggestion
- Advanced availability window suggestions
- Suggested rental alternatives on booking conflicts
- Alternative vehicle suggestions on booking conflicts
- Similar price/category vehicle recommendations
- Available alternative car cards on booking form
- Booking form carry-over between alternative vehicle suggestions
- Secure server-side booking form state transfer
- Expiring booking prefill tokens
- Prefilled alternative vehicle booking forms
- Admin dashboard
- Admin cleanup for expired booking prefill tokens
- Admin booking requests list
- Admin booking detail page
- Admin booking status updates
- Admin car management
- Create/edit cars
- Availability toggle
- Car archive/restore flow
- Archived car exclusion from public catalog and booking
- Archived car exclusion from alternative vehicle suggestions
- PostgreSQL-backed fleet management
- Car image URL validation
- Admin image preview
- Local car image uploads
- Public catalog image display
- Placeholder fallback for missing images
- Admin authentication
- Login/logout flow
- Session-based admin protection
- Protected admin routes
- Bcrypt password verification
- Secure session cookie settings
- CSRF protection for POST forms
- Login brute-force protection
- Security headers middleware
- Hardened image upload validation
- Health endpoint at `/health`

## Planned

- Registration
- Password reset
- Roles and permissions
- OAuth or social login
- Production security hardening
- Advanced responsive redesign
- Infinite scroll or AJAX pagination
- Advanced analytics and reporting
- Calendar UI
- Click-to-fill suggested windows
- Live availability checks
- Multi-car alternatives and more advanced recommendation logic
- Multiple image gallery
- Old uploaded image cleanup
- Cloud or object storage for media
- Delete or archive bookings
- Hard delete cars when safe
- Broader admin cleanup tools
- Advanced UI/UX polish
- Payments
