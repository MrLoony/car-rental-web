# Car Rental Web

Car Rental Web is a server-rendered Go web application for a car rental platform. The project currently includes the application foundation, PostgreSQL connection, database migrations, seeded demo catalog data, Tailwind-based catalog pages, query-parameter catalog filtering, server-side pagination, a booking request flow, availability validation, protected admin authentication, admin booking management, admin fleet management, and basic car image management.

Notifications, advanced media management, and production security hardening are planned but not implemented yet.

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

## Bookings

Booking requests can be started from a car details page:

```text
http://localhost:8080/cars/toyota-corolla/book
```

The booking form is server-rendered and works without JavaScript. Backend validation is the source of truth. Rentals are billed in 24-hour periods using `ceil(duration_hours / 24)`, with a minimum of one billing day. The estimated total is calculated as `billing_days * price_per_day`, and new booking requests are saved with `pending` status.

Availability is checked before a booking request is saved. `pending` and `confirmed` bookings block overlapping requests, while `cancelled` and `completed` bookings do not. After a return time, the car remains unavailable for 4 hours to allow for possible late return, cleaning, washing, inspection, and preparation. When a selected period is unavailable, the form shows the nearest available pickup time when one can be calculated.

## Admin

The admin dashboard is available at:

```text
http://localhost:8080/admin
```

Current admin pages include a booking requests list, booking detail page, booking status update form, car list, car create form, car detail page, car edit form, and availability toggle. Supported booking statuses are `pending`, `confirmed`, `cancelled`, and `completed`. Admin routes are protected with session-based authentication.

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

Admin car management includes create/edit forms, image URL management, local image upload, image preview, and availability toggles.

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

## Authentication

Admin authentication is implemented with a login page, bcrypt password verification, and cookie-based sessions using `gorilla/sessions`. Logging out clears the admin session.

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

## Implemented

- Server-rendered Go application foundation
- Chi router setup
- Template rendering with a base layout and home page
- Tailwind CSS build/watch workflow
- Static file serving
- Environment configuration loading
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
- Admin dashboard
- Admin booking requests list
- Admin booking detail page
- Admin booking status updates
- Admin car management
- Create/edit cars
- Availability toggle
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
- Health endpoint at `/health`

## Planned

- CSRF protection
- Registration
- Password reset
- Roles and permissions
- OAuth or social login
- Production security hardening
- Advanced responsive redesign
- Infinite scroll or AJAX pagination
- Advanced analytics and reporting
- Advanced availability window search
- Email notifications
- Multiple image gallery
- Old uploaded image cleanup
- Cloud or object storage for media
- Delete or archive cars
- Advanced UI/UX polish
- Payments
