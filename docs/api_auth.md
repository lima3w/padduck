# API Authentication (v0.5.0)

## Overview
IPAM Next uses token-based authentication via the Bearer token scheme. All API endpoints except authentication endpoints require a valid token.

## Authentication Flow

### 1. Generate API Token
**POST** `/api/v1/auth/tokens/:userID`

Generate a new API token for a user.

**Request Body:**
```json
{
  "token_name": "My API Token"
}
```

**Response (201 Created):**
```json
{
  "token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z",
  "name": "My API Token"
}
```

**Security:** Store the token securely. It will not be shown again.

---

### 2. List User Tokens
**GET** `/api/v1/auth/tokens/:userID`

List all API tokens for a user (shows token names and creation dates, not the actual tokens).

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "My API Token",
    "created_at": "2026-05-05T12:00:00Z",
    "last_used_at": "2026-05-05T13:30:00Z"
  },
  {
    "id": 2,
    "name": "Production Token",
    "created_at": "2026-05-05T10:00:00Z",
    "last_used_at": null
  }
]
```

---

### 3. Revoke Token
**DELETE** `/api/v1/auth/tokens/:tokenID`

Revoke an API token, preventing further use.

**Response (204 No Content)**

---

## Using Tokens

### Include Token in Requests
All protected endpoints require the `Authorization` header with a Bearer token:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.example.com/api/v1/sections
```

### Token Validation
- Tokens are validated against stored SHA-256 hashes
- Invalid or expired tokens return `401 Unauthorized`
- Token usage is automatically tracked (last_used_at field)

---

## Protected Endpoints

All endpoints except `/api/v1/auth/*` require authentication:

- `GET /api/v1/sections`
- `POST /api/v1/sections`
- `PUT /api/v1/sections/:id`
- `DELETE /api/v1/sections/:id`
- `GET /api/v1/subnets/:id`
- `POST /api/v1/sections/:sectionID/subnets`
- And all IP address endpoints

---

## Example: Complete Authentication Flow

```bash
# 1. Generate a token for user 1
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/tokens/1 \
  -H "Content-Type: application/json" \
  -d '{"token_name": "CLI Token"}' \
  | jq -r '.token')

# 2. Use the token to access protected endpoints
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/sections

# 3. List tokens
curl http://localhost:8080/api/v1/auth/tokens/1

# 4. Revoke a token
curl -X DELETE http://localhost:8080/api/v1/auth/tokens/1
```

---

## Error Responses

### 401 Unauthorized - Missing Header
```json
{"error": "missing authorization header"}
```

### 401 Unauthorized - Invalid Token
```json
{"error": "invalid or expired token"}
```

### 400 Bad Request - Invalid User ID
```json
{"error": "invalid user ID"}
```

---

## Security Considerations

1. **Token Storage:** Tokens are only stored as SHA-256 hashes in the database
2. **Token Transmission:** Always use HTTPS in production
3. **Token Rotation:** Implement token rotation by creating new tokens and revoking old ones
4. **Token Scope:** Tokens can be scoped to specific users (enforce user boundaries in handlers)
5. **Token Expiration:** Optional expiration can be implemented via the `expires_at` field
