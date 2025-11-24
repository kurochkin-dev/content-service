# Content Service API

REST API application for managing articles, built with Go.

## Features

- **CRUD operations** for articles with **pagination support**
- **JWT authentication** with user identification
- **PostgreSQL** database with soft delete support
- **Docker** support with docker-compose
- **Database migrations** using SQL files
- **Structured logging** with zerolog (JSON in production, pretty console in development)
- **Graceful shutdown** for safe server termination
- **CORS middleware** for cross-origin requests
- **Rate limiting** to prevent abuse
- **Health check endpoint** for monitoring
- **Unit tests** for service layer

## Requirements

- Docker 20.10+
- Docker Compose 2.0+

## Quick Start with Docker

### 1. Clone the repository

```bash
git clone <repository-url>
cd content-service
```

### 2. Configure environment variables

Create a `.env` file in the root directory:

```env
# Application
PORT=8080
ENVIRONMENT=development
GIN_MODE=debug

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=content_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key-min-32-chars-for-production
```

### 3. Start services

```bash
docker-compose up -d
```

This will start:
- PostgreSQL database on port `5432`
- API server on port `8080`

### 4. Run database migrations

The application uses GORM AutoMigrate on startup to create tables automatically. For manual migrations using SQL files, you can run migrations from within the container:

```bash
# Apply migrations
docker-compose exec app ./migrate -command up

# Rollback last migration
docker-compose exec app ./migrate -command down

# Check migration version
docker-compose exec app ./migrate -command version
```

Or run migrations locally (requires Go installed):

```bash
go run cmd/migrate/main.go -command up
```

### 5. Verify the service

```bash
curl http://localhost:8080/api/articles
```

## JWT Authentication

### Token Format

JWT token must contain `user_id` in claims:

```json
{
  "user_id": 123,
  "exp": 1234567890,
  "iat": 1234567890,
  "nbf": 1234567890
}
```

Token must be sent in `Authorization` header:
```
Authorization: Bearer <token>
```

### Generating Test Tokens

Before testing protected API endpoints, you need to generate a JWT token. To generate a test JWT token for API testing:

```bash
# From container (using binary)
docker-compose exec app ./token -user-id 123

# Or using go run
docker-compose exec app go run cmd/token/main.go -user-id 123

# Locally (requires Go installed)
go run cmd/token/main.go -user-id 123
```

This will output a JWT token that you can use in the `Authorization: Bearer <token>` header for protected endpoints.

**Example:**
```bash
# Generate token with user ID 123
TOKEN=$(docker-compose exec -T app ./token -user-id 123)

# Use token in API requests
curl -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"title":"My Article","content":"Article content"}' \
     http://localhost:8080/api/articles
```

## API Endpoints

Base URL: `http://localhost:8080/api`

### Health Check

**GET** `/health`

Check service health status.

**Response:** `200 OK`
```json
{
  "status": "ok",
  "service": "content-service"
}
```

### Create Article

**POST** `/articles`

Requires JWT token in `Authorization` header.

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "Article Title",
  "content": "Article content here"
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "title": "Article Title",
  "content": "Article content here",
  "user_id": 123,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Get All Articles

**GET** `/articles?page=1&limit=10`

Supports pagination with query parameters:
- `page` - page number (default: 1)
- `limit` - items per page (default: 10, max: 100)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": 1,
      "title": "Article Title",
      "content": "Article content here",
      "user_id": 123,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 50,
    "total_pages": 5
  }
}
```

### Get Article by ID

**GET** `/articles/{id}`

**Response:** `200 OK`
```json
{
  "id": 1,
  "title": "Article Title",
  "content": "Article content here",
  "user_id": 123,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Update Article

**PUT** `/articles/{id}`

Requires JWT token in `Authorization` header. Users can only update their own articles.

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "Updated Title",
  "content": "Updated content"
}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "title": "Updated Title",
  "content": "Updated content",
  "user_id": 123,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T13:00:00Z"
}
```

### Delete Article

**DELETE** `/articles/{id}`

Requires JWT token in `Authorization` header. Users can only delete their own articles.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:** `204 No Content`

## Database Migrations

The application uses GORM AutoMigrate on startup, which automatically creates and updates database tables based on models.

### Running migrations from container

```bash
# Apply migrations
docker-compose exec app ./migrate -command up

# Rollback last migration
docker-compose exec app ./migrate -command down

# Check migration version
docker-compose exec app ./migrate -command version
```

### Running migrations locally

Requires Go installed locally:

```bash
# Apply migrations
go run cmd/migrate/main.go -command up

# Rollback last migration
go run cmd/migrate/main.go -command down

# Check migration version
go run cmd/migrate/main.go -command version
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Application port | `8080` |
| `ENVIRONMENT` | Environment (development, staging, test, production) | `development` |
| `GIN_MODE` | Gin framework mode (debug, release) | `debug` (auto) |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `content_db` |
| `DB_SSLMODE` | SSL mode (disable, require, verify-ca, verify-full) | `disable` |
| `DB_MAX_OPEN_CONNS` | Maximum number of open database connections | `25` |
| `DB_MAX_IDLE_CONNS` | Maximum number of idle database connections | `5` |
| `DB_CONN_MAX_LIFETIME_MIN` | Maximum connection lifetime in minutes | `5` |
| `DB_CONN_MAX_IDLE_TIME_MIN` | Maximum connection idle time in minutes | `2` |
| `JWT_SECRET` | JWT secret key (min 32 chars in production) | Auto-generated for dev |
| `CORS_ALLOWED_ORIGIN` | Allowed CORS origin (e.g., `http://localhost:3000`) | `*` (dev), empty (prod) |
| `AUTO_MIGRATE` | Enable/disable automatic database migrations (`true`/`false`) | `true` (dev), `false` (prod) |

## Rate Limiting

The API implements rate limiting to prevent abuse:
- **Limit:** 100 requests per window
- **Refill rate:** 10 tokens per second
- **Response:** `429 Too Many Requests` when limit is exceeded

**Example 429 Response:**
```json
{
  "error": "rate limit exceeded, please try again later"
}
```

## Error Responses

All errors follow this format:

```json
{
  "error": "Error message"
}
```

Or for validation errors:

```json
{
  "errors": [
    "title is required",
    "content is too short"
  ]
}
```

### Common Error Codes

- `400 Bad Request` - Invalid request data or validation errors
- `401 Unauthorized` - Missing or invalid JWT token
- `403 Forbidden` - User doesn't have permission (e.g., trying to update/delete someone else's article)
- `404 Not Found` - Article not found
- `500 Internal Server Error` - Server error

**Example 403 Forbidden:**
```json
{
  "error": "you can only update your own articles"
}
```

## Running Tests

The project includes unit tests for the service layer. Tests should be run **locally**, not inside the Docker container.

**Why not in container?** The production Docker image uses multi-stage build and doesn't include Go toolchain. This keeps the image small (~20MB) and secure.

### Run Tests Locally

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage report
go test ./... -cover

# Run tests in a specific package
go test ./internal/article/...

# Run tests with detailed coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Expected Output

```
?       content-service/cmd/migrate     [no test files]
?       content-service/cmd/server      [no test files]
?       content-service/cmd/token       [no test files]
ok      content-service/internal/article        0.587s
?       content-service/internal/shared/config  [no test files]
...
```

All tests should pass with `PASS` status.

## Stopping Services

```bash
docker-compose down
```

To remove volumes (database data):

```bash
docker-compose down -v
```

## Project Structure

```
.
├── cmd/
│   ├── migrate/          # Migration command
│   ├── server/           # Main application
│   └── token/            # Token generator utility
├── internal/
│   ├── article/          # Article domain
│   │   ├── constants.go  # Domain constants
│   │   ├── errors.go     # Custom errors
│   │   ├── handler.go    # HTTP handlers
│   │   ├── model.go      # Data models
│   │   ├── repository.go # Data access layer
│   │   ├── service.go    # Business logic
│   │   └── service_test.go # Unit tests
│   └── shared/           # Shared packages
│       ├── config/       # Configuration management
│       ├── database/     # Database connection
│       ├── logging/      # Structured logging
│       ├── middleware/   # HTTP middlewares (auth, rate limiting)
│       └── validation/   # Input validation
├── migrations/           # SQL migration files
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Architecture & Design

### Layered Architecture
- **Handler Layer:** HTTP request/response handling
- **Service Layer:** Business logic and validation
- **Repository Layer:** Data access abstraction

### Key Design Patterns
- **Dependency Injection:** Services receive dependencies through constructors
- **Repository Pattern:** Abstraction over database operations
- **Middleware Pattern:** Cross-cutting concerns (auth, CORS, rate limiting)
- **Error Wrapping:** Context-aware error handling with custom error types

### Production-Ready Features
- ✅ **Graceful Shutdown:** Safe server termination without dropping requests
- ✅ **Structured Logging:** JSON logs in production, pretty console in development
- ✅ **Request Validation:** Input validation with detailed error messages
- ✅ **Pagination:** Efficient data retrieval for large datasets
- ✅ **Rate Limiting:** Protection against abuse and DoS attacks
- ✅ **Health Checks:** Easy service monitoring
- ✅ **CORS Support:** Ready for frontend integration
- ✅ **Soft Delete:** Data preservation with DeletedAt timestamps
- ✅ **Connection Pooling:** Optimized database connections
- ✅ **Unit Tests:** Service layer test coverage

