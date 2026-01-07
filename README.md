# boot-go-http-server

A production-ready HTTP server built in Go that implements a "Chirps" API (similar to Twitter/X). This project demonstrates modern Go web development patterns including RESTful API design, JWT authentication, database integration, middleware, and webhook handling.

Built while following Boot.dev's **Learn HTTP Servers in Go** course, which covers routing, middleware, JSON APIs, auth/JWTs, webhooks, and production-minded patterns in Go ([course link](https://www.boot.dev/courses/learn-http-servers-golang)).

## What This Project Does

This server provides a complete backend API for a social media platform where users can:

- **User Management**: Create accounts, update profiles, and manage authentication
- **Authentication**: Secure JWT-based authentication with refresh tokens
- **Chirps**: Post, read, update, and delete short messages ("chirps")
- **Content Filtering**: Automatic profanity filtering for user-generated content
- **Webhooks**: Integration with external services (Polka) for premium user upgrades
- **Admin Features**: Metrics and administrative endpoints
- **Static File Serving**: Serve frontend assets and track file server hits

The API is organized into three main route groups:

- `/api/` - RESTful API endpoints for users, chirps, and authentication
- `/app/` - Static file server for frontend assets
- `/admin/` - Administrative endpoints and metrics

## Why You Should Care

This project serves as an excellent learning resource and reference implementation for:

- **Production-Ready Patterns**: Demonstrates real-world HTTP server architecture in Go
- **Security Best Practices**: Implements secure password hashing (Argon2id), JWT authentication, and token refresh mechanisms
- **Database Integration**: Uses PostgreSQL with SQLC for type-safe database queries and migrations
- **Clean Architecture**: Well-organized code structure with separation of concerns
- **Modern Go Features**: Leverages Go 1.25+ features and standard library patterns
- **API Design**: RESTful API design with proper error handling and JSON responses

Whether you're learning Go web development, building a similar API, or looking for production patterns, this codebase provides a solid foundation.

## Installation and Setup

### Prerequisites

- **Go 1.25.5** or later (see `go.mod` for exact version)
- **PostgreSQL** database (for data persistence)
- **[just](https://github.com/casey/just)** command runner (optional but recommended)
- **goose** and **sqlc** (optional, for database migrations and code generation)

### Step 1: Clone the Repository

```bash
git clone <repository-url>
cd boot-go-http-server
```

### Step 2: Install Dependencies

```bash
go mod download
```

### Step 3: Set Up PostgreSQL Database

Create a PostgreSQL database for the project:

```bash
createdb golang  # or use your preferred database name
```

### Step 4: Configure Environment Variables

Create a `.env` file in the project root with the following required variables:

```env
# Database connection string (required)
DB_URL=postgres://username:password@localhost:5432/golang

# JWT secret for token signing (required)
JWT_SECRET=your-secret-key-here

# Polka API key for webhook authentication (required)
POLKA_KEY=your-polka-api-key

# Optional: Server port (defaults to 8080)
PORT=8080

# Optional: Environment mode (defaults to development)
ENVIRONMENT=development
```

**Note**: Make sure `.env` is in your `.gitignore` (it should be by default) to avoid committing secrets.

### Step 5: Run Database Migrations

Apply the database schema migrations:

```bash
# Using just (recommended)
just migrate

# Or manually with goose
goose -dir ./sql/schema postgres "your-db-url" up
```

### Step 6: Run the Server

```bash
# Using just (recommended)
just run

# Or directly with Go
go run .
```

The server will start on the port specified in your `PORT` environment variable (default: `8080`). You should see:

```
Listening on port 8080
```

### Step 7: Verify Installation

Test the health endpoint:

```bash
curl http://localhost:8080/api/healthz
```

You should receive `OK` as the response.

## Development Commands

The project uses `just` for common development tasks. List all available commands:

```bash
just --list
```

### Common Commands

- `just run` — Start the development server (`go run .`)
- `just test` — Run auth package tests
- `just testv` — Run tests with verbose output
- `just migrate` — Apply database migrations via goose
- `just rollback` — Roll back the last migration
- `just sqlc` — Regenerate SQLC code from SQL queries

## API Endpoints

### Authentication

- `POST /api/users` - Create a new user account
- `POST /api/login` - Authenticate and receive JWT tokens
- `PUT /api/users` - Update user account (requires authentication)
- `POST /api/refresh` - Refresh access token
- `POST /api/revoke` - Revoke refresh token

### Chirps

- `GET /api/chirps` - List all chirps (supports `author_id` and `sort` query params)
- `GET /api/chirps/{chirpID}` - Get a specific chirp
- `POST /api/chirps` - Create a new chirp (requires authentication)
- `DELETE /api/chirps/{chirpID}` - Delete a chirp (requires authentication, owner only)

### Admin

- `GET /admin/metrics` - View file server hit metrics

### Health

- `GET /api/healthz` - Health check endpoint

## Project Structure

```
.
├── api.go              # Main API route handlers
├── admin.go            # Admin endpoints
├── app.go              # Static file server
├── main.go             # Application entry point
├── middleware.go       # HTTP middleware (auth, logging, CORS)
├── helper.go           # Utility functions
├── internal/
│   ├── auth/          # Authentication utilities (JWT, password hashing)
│   └── database/      # Database queries and models (generated by SQLC)
├── sql/
│   ├── schema/        # Database migrations
│   └── queries/       # SQL queries for SQLC
└── assets/            # Static frontend assets
```

## Technologies Used

- **Go 1.25.5** - Programming language
- **PostgreSQL** - Relational database
- **SQLC** - Type-safe SQL code generation
- **Goose** - Database migrations
- **JWT** - JSON Web Tokens for authentication
- **Argon2id** - Password hashing algorithm
- **godotenv** - Environment variable management
