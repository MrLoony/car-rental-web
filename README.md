# Car Rental Web

Car Rental Web is a server-rendered Go web application for a car rental platform. The project currently includes the application foundation, PostgreSQL connection, database migrations, seeded demo catalog data, and Tailwind-based catalog pages.

Booking flows, admin tools, authentication, and catalog filtering are planned but not implemented yet.

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
- Health endpoint at `/health`

## Planned

- Filters, search, and sorting
- Booking requests
- Admin dashboard
- Authentication
- Image upload or asset management
