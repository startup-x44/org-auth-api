# API Specification - Multi-Tenant Authentication Service

## Overview

This document provides detailed API specifications for the Multi-Tenant Authentication Service, a comprehensive SaaS authentication system built with Go and React.

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication

All protected endpoints require:
- **Authorization**: `Bearer <access_token>`
- **X-Tenant-ID**: `<tenant_id>` (UUID or domain)

## Response Format

All responses follow this structure:
```json
{
  "success": boolean,
  "data": object | array | null,
  "message": string
}
```

## Endpoints

### Authentication

#### POST /auth/register
Register a new user account with automatic tenant creation.

**Request Body:**
```json
{
  "email": "user@company.com",
  "password": "SecurePass123!",
  "confirm_password": "SecurePass123!",
  "user_type": "Student",
  "tenant_id": "company.com",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@company.com",
      "user_type": "Student",
      "tenant_id": "uuid",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-01-15T10:30:00Z"
    },
    "token": {
      "access_token": "eyJhbGciOiJSUzI1NiIs...",
      "refresh_token": "refresh_token_here",
      "expires_in": 3600,
      "token_type": "Bearer"
    }
  },
  "message": "User registered successfully"
}
```

**Error Responses:**
- `400`: Invalid request data or validation errors
- `409`: User already exists

#### POST /auth/login
Authenticate a user and return tokens.

**Request Body:**
```json
{
  "email": "user@company.com",
  "password": "SecurePass123!",
  "tenant_id": "company.com"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@company.com",
      "user_type": "Student",
      "tenant_id": "uuid",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "last_login": "2025-01-15T10:30:00Z"
    },
    "token": {
      "access_token": "eyJhbGciOiJSUzI1NiIs...",
      "refresh_token": "refresh_token_here",
      "expires_in": 3600,
      "token_type": "Bearer"
    }
  },
  "message": "Login successful"
}
```

**Error Responses:**
- `400`: Invalid credentials or account locked
- `401`: Invalid tenant

#### POST /auth/refresh
Refresh access token using refresh token.

**Request Body:**
```json
{
  "refresh_token": "refresh_token_here"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "token": {
      "access_token": "new_access_token",
      "refresh_token": "new_refresh_token",
      "expires_in": 3600,
      "token_type": "Bearer"
    }
  },
  "message": "Token refreshed successfully"
}
```

#### POST /auth/forgot-password
Initiate password reset process.

**Request Body:**
```json
{
  "email": "user@company.com",
  "tenant_id": "company.com"
}
```

#### POST /auth/reset-password
Reset password using reset token.

**Request Body:**
```json
{
  "token": "reset_token_here",
  "new_password": "NewSecurePass123!",
  "confirm_password": "NewSecurePass123!"
}
```

### User Management

#### GET /user/profile
Get current user profile.

**Headers:**
```
Authorization: Bearer <access_token>
X-Tenant-ID: <tenant_id>
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@company.com",
    "user_type": "Student",
    "tenant_id": "uuid",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "is_active": true,
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z",
    "last_login": "2025-01-15T10:30:00Z"
  }
}
```

#### PUT /user/profile
Update user profile.

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Smith",
  "phone": "+1987654321"
}
```

#### POST /user/change-password
Change user password.

**Request Body:**
```json
{
  "current_password": "OldPass123!",
  "new_password": "NewPass123!",
  "confirm_password": "NewPass123!"
}
```

#### POST /user/logout
Logout user and invalidate session.

**Request Body:**
```json
{
  "user_id": "uuid",
  "refresh_token": "refresh_token_here"
}
```

### Admin Endpoints

#### GET /admin/users
List users with pagination.

**Query Parameters:**
- `limit` (optional): Number of users to return (default: 10, max: 100)
- `cursor` (optional): Cursor for pagination

**Response (200):**
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "uuid",
        "email": "user@company.com",
        "user_type": "Student",
        "tenant_id": "uuid",
        "first_name": "John",
        "last_name": "Doe",
        "is_active": true,
        "created_at": "2025-01-15T10:30:00Z"
      }
    ],
    "total": 25,
    "limit": 10,
    "next_cursor": "next_cursor_value"
  }
}
```

#### PUT /admin/users/{userId}/activate
Activate a user account.

#### PUT /admin/users/{userId}/deactivate
Deactivate a user account.

#### DELETE /admin/users/{userId}
Delete a user account.

#### POST /admin/tenants
Create a new tenant.

**Request Body:**
```json
{
  "name": "New Company",
  "domain": "newcompany.com"
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "New Company",
    "domain": "newcompany.com",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  },
  "message": "Tenant created successfully"
}
```

#### GET /admin/tenants
List tenants with pagination.

**Query Parameters:**
- `limit` (optional): Number of tenants to return (default: 10, max: 100)
- `offset` (optional): Offset for pagination (default: 0)

#### GET /admin/tenants/{tenantId}
Get tenant details.

#### PUT /admin/tenants/{tenantId}
Update tenant.

**Request Body:**
```json
{
  "name": "Updated Company Name",
  "domain": "updatedcompany.com"
}
```

#### DELETE /admin/tenants/{tenantId}
Delete tenant (only if no users exist).

### Health Check

#### GET /health
Check service health.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "database": "connected",
    "redis": "connected",
    "uptime": "1h 30m"
  }
}
```

## Data Types

### User Types
- `Admin` - System administrator
- `Student` - Student user
- `RTO` - Registered Training Organization
- `Issuer` - Certificate issuer
- `Validator` - Certificate validator
- `badger` - Badger user type
- `Non-partner` - Non-partner user
- `Partner` - Partner user

### User Status
- `active` - Account is active
- `suspended` - Account is suspended
- `deactivated` - Account is deactivated

### Tenant Status
- `active` - Tenant is active
- `suspended` - Tenant is suspended

## Error Codes

### Common HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (invalid credentials/token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `409` - Conflict (resource already exists)
- `422` - Unprocessable Entity (validation failed)
- `429` - Too Many Requests (rate limited)
- `500` - Internal Server Error

### Error Response Format
```json
{
  "success": false,
  "message": "Error description",
  "errors": "Detailed error information"
}
```

## Rate Limiting

- **Global Rate Limit**: 100 requests per minute per IP
- **Authentication Endpoints**: Additional rate limiting
- **Admin Endpoints**: Stricter rate limiting

## Security

### Authentication
- JWT tokens with RSA signing
- Refresh token rotation
- Automatic token refresh

### Multi-Tenant Isolation
- Database-level row security
- Tenant context injection
- Session scoping per tenant

### Input Validation
- Comprehensive input sanitization
- SQL injection prevention
- XSS protection

### CORS Configuration
- Configurable allowed origins
- Support for wildcard patterns (*.domain.com)
- Credential support for cookies

## SDKs and Examples

### cURL Examples

#### Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@company.com",
    "password": "SecurePass123!",
    "confirm_password": "SecurePass123!",
    "user_type": "Student",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@company.com",
    "password": "SecurePass123!",
    "tenant_id": "company.com"
  }'
```

#### Get Profile
```bash
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "X-Tenant-ID: company.com"
```

### JavaScript Example
```javascript
// Login
const loginResponse = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'user@company.com',
    password: 'SecurePass123!',
    tenant_id: 'company.com'
  })
});

const { data } = await loginResponse.json();
const { access_token, refresh_token } = data.token;

// Store tokens (tenant-specific)
localStorage.setItem(`access_token_company.com`, access_token);
localStorage.setItem(`refresh_token_company.com`, refresh_token);
localStorage.setItem('tenant_id', 'company.com');

// Make authenticated request
const profileResponse = await fetch('/api/v1/user/profile', {
  headers: {
    'Authorization': `Bearer ${access_token}`,
    'X-Tenant-ID': 'company.com'
  }
});
```

## Versioning

- **API Version**: v1
- **Version Header**: Not required (included in URL path)
- **Breaking Changes**: New major version
- **Backward Compatibility**: Maintained within major versions

## Changelog

### v1.0.0
- Initial release
- Multi-tenant authentication
- JWT token management
- User management
- Admin functionality
- Health checks</content>
<parameter name="filePath">/Users/niloflora/fligno/blocksure/abc/auth-service/docs/api-spec.md