# API Stabilization (v0.7.0)

## Standard Error Response Format

All API errors follow a consistent format with error codes and details:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": "Optional additional details"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid request parameters or body |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Authenticated but not authorized |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Request conflicts with current state |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests from this IP |
| `INTERNAL_SERVER_ERROR` | 500 | Server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

### Example Error Responses

**Missing Authentication:**
```bash
curl https://api.example.com/api/v1/sections
```

Response:
```json
{
  "error": "Missing authorization header",
  "code": "UNAUTHORIZED"
}
```

**Invalid Request:**
```bash
curl -X POST https://api.example.com/api/v1/sections \
  -H "Authorization: Bearer TOKEN" \
  -d '{"name": ""}'
```

Response:
```json
{
  "error": "section name is required",
  "code": "BAD_REQUEST"
}
```

**Rate Limited:**
```bash
# After 100 requests in 1 minute
curl https://api.example.com/api/v1/sections
```

Response:
```json
{
  "error": "Too many requests. Please try again later.",
  "code": "RATE_LIMIT_EXCEEDED",
  "details": "Rate limit: 100 requests per 1m0s"
}
```

---

## Rate Limiting

### Policy
- **Limit:** 100 requests per minute per IP address
- **Scope:** All endpoints (except health check)
- **Response:** 429 Too Many Requests with error details

### Headers
The following headers indicate rate limit status:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 42
X-RateLimit-Reset: 1735689600
```

### Handling Rate Limits

When you receive a 429 response:
1. Wait before retrying (exponential backoff recommended)
2. Store the `X-RateLimit-Reset` timestamp to know when limits reset
3. Monitor `X-RateLimit-Remaining` to avoid hitting the limit

**Recommended retry strategy:**
```javascript
async function withRetry(fn, maxAttempts = 3) {
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (error) {
      if (error.status === 429) {
        const resetTime = parseInt(error.headers['X-RateLimit-Reset']);
        const waitTime = Math.max(0, resetTime * 1000 - Date.now());
        if (attempt < maxAttempts) {
          await new Promise(resolve => setTimeout(resolve, waitTime));
        }
      } else {
        throw error;
      }
    }
  }
}
```

---

## API Versioning

The API uses `/api/v1` prefix for all endpoints. This ensures:
- **Stability:** Changes to v1 are backward compatible within major version
- **Migration:** New versions (v2, v3, etc.) can coexist
- **Deprecation:** Old versions can be deprecated with notice period

### Current Version: v1

All endpoints use the `/api/v1` prefix:
```
GET /api/v1/sections
POST /api/v1/sections
PUT /api/v1/sections/:id
DELETE /api/v1/sections/:id
```

---

## Response Consistency

### Success Responses

**Single Resource:**
```json
{
  "id": 1,
  "name": "Network A",
  "created_at": "2026-05-04T12:00:00Z",
  "updated_at": "2026-05-04T12:00:00Z"
}
```

**Collection:**
```json
[
  {
    "id": 1,
    "name": "Network A",
    "created_at": "2026-05-04T12:00:00Z"
  },
  {
    "id": 2,
    "name": "Network B",
    "created_at": "2026-05-04T13:00:00Z"
  }
]
```

### Pagination

Search endpoints support pagination:
```bash
curl https://api.example.com/api/v1/sections/search \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "query": "network",
    "limit": 50,
    "offset": 0
  }'
```

Response is always an array with applied pagination limits.

---

## Status Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Successful GET/PUT/PATCH |
| 201 | Created | Successful POST (resource created) |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Invalid parameters |
| 401 | Unauthorized | Missing/invalid auth |
| 403 | Forbidden | Authenticated but not allowed |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | State conflict |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server failure |
| 503 | Service Unavailable | Database/service down |
