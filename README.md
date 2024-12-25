# Subteacher Backend API

This is the backend API for the Subteacher application, built with Go, Gin, and GORM.

## Prerequisites

- Go 1.16 or higher
- MySQL 5.7 or higher
- Make sure you have `go` installed and your `GOPATH` is set up correctly

## Setup

1. Clone the repository
2. Navigate to the backend directory
3. Copy `.env.example` to `.env` and update the values:
   ```bash
   cp .env.example .env
   ```
4. Install dependencies:
   ```bash
   go mod download
   ```
5. Create the database:
   ```bash
   mysql -u root -p < internal/database/migrations/000001_init_schema.up.sql
   ```

## Running the Application

1. Start the server:
   ```bash
   go run cmd/api/main.go
   ```
   The server will start on the port specified in your `.env` file (default: 8080)

## API Endpoints

### Authentication

#### Sign Up
```http
POST /api/auth/signup
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123",
    "user_type": 2,
    "full_name": "John Doe",
    "phone_number": "+966501234567",
    "city_id": 1,
    "price_per_day": 200.00
}
```

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

#### Refresh Token
```http
POST /api/auth/refresh
X-Refresh-Token: your-refresh-token
```

## Development

- The project uses GORM for database operations
- JWT for authentication
- Gin for routing
- Configuration is managed through environment variables

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go
├── config/
│   └── config.go
├── internal/
│   ├── database/
│   │   └── migrations/
│   ├── handlers/
│   │   └── auth.go
│   ├── middleware/
│   └── models/
│       └── user.go
└── pkg/
    └── utils/
``` 