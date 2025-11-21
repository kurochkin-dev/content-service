# Content Service API

REST API application for managing articles, built with Go.

## Features

- **CRUD operations** for articles
- **JWT authentication** with user identification
- **PostgreSQL** database
- **Docker** support with docker-compose
- **Database migrations** using SQL files

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

**GET** `/articles`

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "title": "Article Title",
    "content": "Article content here",
    "user_id": 123,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
]
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
| `JWT_SECRET` | JWT secret key (min 32 chars in production) | Auto-generated for dev |

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
│   ├── article/          # Article domain (handler, service, repository)
│   └── shared/           # Shared packages (config, database, middleware)
├── migrations/           # SQL migration files
├── Dockerfile
├── docker-compose.yml
└── README.md
```

