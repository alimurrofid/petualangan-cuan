# Petualangan Cuan - Backend (Go)

Backend service for Petualangan Cuan application, built with Go.

## Prerequisites

- [Go](https://go.dev/dl/) (version 1.20+ recommended)
- [Make](https://www.gnu.org/software/make/) (optional, for using Makefile commands)
- Database (MySQL/PostgreSQL depending on config)

## Installation

1. Navigate to the backend directory:
   ```bash
   cd backend-go
   ```
2. Install dependencies:
   ```bash
   go mod download
   ```

## Configuration

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```
2. Open `.env` and configure your database credentials and server port.

## Running the Application

You can run the application using `go run` or the provided `Makefile`.

### Using Makefile

- **Run the application:**
  ```bash
  make run
  ```
- **Reset database (Fresh):**
  ```bash
  make fresh
  ```
- **Seed database:**
  ```bash
  make seed
  ```
- **Reset and Seed:**
  ```bash
  make fresh-seed
  ```

### Manual Commands

If you don't have Make installed:

- **Run:**
  ```bash
  go run cmd/api/main.go
  ```
- **Fresh (Reset DB):**
  ```bash
  go run cmd/api/main.go -fresh
  ```
- **Seed:**
  ```bash
  go run cmd/api/main.go -seed
  ```

## Project Structure

```bash
backend-go/
├── cmd/
│   └── api/            # Application entry point
├── internal/
│   ├── config/         # Configuration (Database, Env, etc.)
│   ├── entity/         # Domain entities/models
│   ├── handler/        # HTTP handlers (Controllers)
│   ├── repository/     # Data access layer
│   ├── seeder/         # Database seeding logic
│   └── service/        # Business logic layer
├── pkg/
│   └── middleware/     # Shared middleware
├── docs/               # Documentation
├── .env                # Environment variables
├── Makefile            # Build and run commands
├── go.mod              # Go module definition
└── go.sum              # Go module checksums
```

