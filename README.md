# Car Rental Web

Car Rental Web is a server-rendered Go web application for a car rental platform. The project currently has the application foundation in place: configuration loading, Chi routing, HTML templates, static assets, and a Tailwind CSS frontend setup.

Database integration and rental business logic are planned but not implemented yet.

## Tech Stack

- Go
- Chi Router
- HTML Templates
- Tailwind CSS
- JavaScript
- PostgreSQL planned

## Project Structure

- `cmd/web` - application entry point
- `internal/config` - environment configuration loading
- `internal/handler` - routes, handlers, and template rendering
- `web/templates` - server-rendered HTML templates
- `web/static` - CSS, JavaScript, and image assets

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

## Implemented

- Server-rendered Go application foundation
- Chi router setup
- Template rendering with a base layout and home page
- Tailwind CSS build/watch workflow
- Static file serving
- Environment configuration loading
- Health endpoint at `/health`

## Planned

- Cars catalog
- Booking requests
- PostgreSQL integration
- Admin dashboard
