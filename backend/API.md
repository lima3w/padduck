# Padduck API Reference - v0.1.0

## Overview

The Padduck API is RESTful and uses JSON for request/response bodies. All endpoints are prefixed with `/api/v1`.

## Base URL

```
http://localhost:8080/api/v1
```

## Health Check

Check server and database status.

```
GET /health
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "database": "connected"
}
```

**Response (503 Service Unavailable - Database Down):**
```json
{
  "status": "degraded",
  "database": "disconnected",
  "error": "connection refused"
}
```

## Sections API

Sections are top-level IP address groupings.

### List Sections

```
GET /api/v1/sections
```

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "Production",
    "description": "Production network segments",
    "created_by": 1,
    "created_at": "2026-05-04T10:00:00Z",
    "updated_at": "2026-05-04T10:00:00Z"
  }
]
```

### Create Section

```
POST /api/v1/sections
Content-Type: application/json

{
  "name": "Production",
  "description": "Production network segments",
  "created_by": 1
}
```

**Response (201 Created):**
```json
{
  "id": 1,
  "name": "Production",
  "description": "Production network segments",
  "created_by": 1,
  "created_at": "2026-05-04T10:00:00Z",
  "updated_at": "2026-05-04T10:00:00Z"
}
```

**Response (400 Bad Request):**
```json
{
  "error": "section name is required"
}
```

### Get Section

```
GET /api/v1/sections/:id
```

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "Production",
  "description": "Production network segments",
  "created_by": 1,
  "created_at": "2026-05-04T10:00:00Z",
  "updated_at": "2026-05-04T10:00:00Z"
}
```

**Response (404 Not Found):**
```json
{
  "error": "section not found"
}
```

### Update Section

```
PUT /api/v1/sections/:id
Content-Type: application/json

{
  "name": "Production Updated",
  "description": "Updated description"
}
```

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "Production Updated",
  "description": "Updated description",
  "created_by": 1,
  "created_at": "2026-05-04T10:00:00Z",
  "updated_at": "2026-05-04T14:30:00Z"
}
```

### Delete Section

```
DELETE /api/v1/sections/:id
```

**Response (204 No Content)**

## Error Responses

### 400 Bad Request
Invalid request format or parameters.

```json
{
  "error": "invalid request body"
}
```

### 404 Not Found
Resource not found.

```json
{
  "error": "section not found"
}
```

### 500 Internal Server Error
Server error occurred.

```json
{
  "error": "internal server error"
}
```

### 503 Service Unavailable
Service (database) is unavailable.

```json
{
  "error": "database connection failed"
}
```

## Data Models

### Section

```typescript
{
  id: number;           // Primary key
  name: string;         // Section name (required)
  description?: string; // Description
  created_by: number;   // User ID who created this section
  created_at: string;   // ISO 8601 timestamp
  updated_at: string;   // ISO 8601 timestamp
}
```

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200  | OK - Request successful |
| 201  | Created - Resource created |
| 204  | No Content - Successful deletion |
| 400  | Bad Request - Invalid input |
| 404  | Not Found - Resource doesn't exist |
| 500  | Internal Server Error |
| 503  | Service Unavailable - Database down |

## Logging

All requests are logged with timestamp, HTTP method, and path.

Example log:
```
2026-05-04T14:30:00Z POST /api/v1/sections
2026-05-04T14:30:01Z GET /api/v1/sections/1
```

## Middleware

- **Logging**: All requests are logged
- **JSON Content-Type**: Responses are JSON
- **Error Handling**: Consistent error format

## Future Endpoints (v0.2.0+)

- Subnets CRUD: `/api/v1/subnets`
- IP Addresses CRUD: `/api/v1/ip-addresses`
- Search: `POST /api/v1/search`
- User Management: `POST /api/v1/users`
