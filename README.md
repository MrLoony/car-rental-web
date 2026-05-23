# Car Rental Web

Car Rental Web is a server-rendered Go web application for a car rental platform. The project currently includes the application foundation, PostgreSQL connection, database migrations, seeded demo catalog data, Tailwind-based catalog pages, query-parameter catalog filtering, and a booking request flow.

Admin tools, authentication, availability conflict checking, notifications, and image management are planned but not implemented yet.

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

Available filters include text search, category, fuel type, transmission, and sort order. The filter form works without JavaScript; JavaScript only enhances the experience with debounced search and automatic submit for select fields.

## Bookings

Booking requests can be started from a car details page:

```text
http://localhost:8080/cars/toyota-corolla/book
```

The booking form is server-rendered and works without JavaScript. Backend validation is the source of truth. Rentals are billed in 24-hour periods using `ceil(duration_hours / 24)`, with a minimum of one billing day. The estimated total is calculated as `billing_days * price_per_day`, and new booking requests are saved with `pending` status.

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
- Progressive JavaScript enhancement with no-JS fallback
- Booking request form
- Backend booking validation
- 24-hour rental billing calculation
- Estimated total calculation
- Booking persistence in PostgreSQL
- Booking success page
- JavaScript booking price preview
- Health endpoint at `/health`

## Planned

- Admin dashboard
- Authentication
- Availability conflict checking
- Email notifications
- Image upload or asset management
- Payments
- Pagination
