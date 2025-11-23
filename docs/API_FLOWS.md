# API Workflows & Authentication Flows

**Version**: 1.0.0  
**Last Updated**: November 23, 2025  
**OpenAPI Spec**: See [OPENAPI_SPEC.yaml](./OPENAPI_SPEC.yaml) for complete API documentation

## Quick Links
- [OpenAPI 3.0 Specification](./OPENAPI_SPEC.yaml)
- [Authentication Flow](#authentication-flow-diagrams)
- [API Endpoint Groups](#api-endpoint-groups)
- [Sample Responses](#sample-api-responses)

---

## Authentication Flow Diagrams

### 1. User Registration Flow

```
┌──────┐                ┌─────────┐              ┌──────────┐
│Client│                │ Backend │              │ Database │
└──┬───┘                └────┬────┘              └────┬─────┘
   │                         │                        │
   │ POST /auth/register     │                        │
   ├────────────────────────>│                        │
   │ { email, password }     │                        │
   │                         │                        │
   │                         │ Check duplicate email  │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Hash password (Argon2) │
   │                         │                        │
   │                         │ Create user            │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Generate verification  │
   │                         │ token (UUID)           │
   │                         │                        │
   │                         │ Store token            │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Send verification email│
   │                         ├───────────────────────>│
   │                         │                        │
   │<────────────────────────┤                        │
   │ 201 Created             │                        │
   │ { success, message }    │                        │
   │                         │                        │
```

**Steps:**
1. Client sends email + password
2. Backend validates email format
3. Check for duplicate email
4. Hash password with Argon2
5. Create user record (status: pending)
6. Generate verification token
7. Send verification email
8. Return success response

**Security:**
- ✅ Email uniqueness enforced at DB level
- ✅ Password hashed (never stored plaintext)
- ✅ Token expires in 24 hours
- ✅ Rate limiting: 10 registrations per hour per IP

---

### 2. Email Verification Flow

```
┌──────┐                ┌─────────┐              ┌──────────┐
│Client│                │ Backend │              │ Database │
└──┬───┘                └────┬────┘              └────┬─────┘
   │                         │                        │
   │ Click email link        │                        │
   │ GET /verify?token=xxx   │                        │
   ├────────────────────────>│                        │
   │                         │                        │
   │                         │ Find token             │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Check expiration       │
   │                         │                        │
   │                         │ Mark email_verified_at │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Delete token           │
   │                         ├───────────────────────>│
   │                         │                        │
   │<────────────────────────┤                        │
   │ 200 OK                  │                        │
   │ Redirect to login       │                        │
```

---

### 3. Login Flow (Multi-Organization)

```
┌──────┐                ┌─────────┐              ┌──────────┐
│Client│                │ Backend │              │ Database │
└──┬───┘                └────┬────┘              └────┬─────┘
   │                         │                        │
   │ POST /auth/login        │                        │
   ├────────────────────────>│                        │
   │ { email, password }     │                        │
   │                         │                        │
   │                         │ Find user by email     │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Check failed attempts  │
   │                         │ (account lockout)      │
   │                         │                        │
   │                         │ Verify password        │
   │                         │ (Argon2 comparison)    │
   │                         │                        │
   │                         │ Check email verified   │
   │                         │                        │
   │                         │ Get user's orgs        │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │<────────────────────────┤                        │
   │ 200 OK                  │                        │
   │ { user, organizations,  │                        │
   │   needsOrgSelection }   │                        │
   │                         │                        │
   │                         │                        │
   │ POST /auth/select-org   │                        │
   ├────────────────────────>│                        │
   │ { organization_id }     │                        │
   │                         │                        │
   │                         │ Verify membership      │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Get user's role        │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Get role permissions   │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Create session         │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Generate JWT           │
   │                         │ (with org context)     │
   │                         │                        │
   │                         │ Create refresh token   │
   │                         ├───────────────────────>│
   │                         │                        │
   │<────────────────────────┤                        │
   │ { access_token,         │                        │
   │   refresh_token,        │                        │
   │   organization }        │                        │
```

**Key Points:**
- Login returns user + organizations list
- Client presents organization selector
- Second request to `/auth/select-organization`
- Session and tokens created AFTER org selection
- JWT contains organization context
- Refresh token scoped to organization

---

### 4. Token Refresh Flow

```
┌──────┐                ┌─────────┐              ┌──────────┐
│Client│                │ Backend │              │ Database │
└──┬───┘                └────┬────┘              └────┬─────┘
   │                         │                        │
   │ POST /auth/refresh      │                        │
   ├────────────────────────>│                        │
   │ { refresh_token }       │                        │
   │                         │                        │
   │                         │ Decode JWT (no verify) │
   │                         │ Extract user_id, org_id│
   │                         │                        │
   │                         │ Find refresh token     │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Verify token hash      │
   │                         │ (Argon2 comparison)    │
   │                         │                        │
   │                         │ Check not expired      │
   │                         │ Check not revoked      │
   │                         │                        │
   │                         │ Get user permissions   │
   │                         ├───────────────────────>│
   │                         │<───────────────────────┤
   │                         │                        │
   │                         │ Generate new access    │
   │                         │ token (JWT)            │
   │                         │                        │
   │                         │ Update session         │
   │                         │ last_activity          │
   │                         ├───────────────────────>│
   │                         │                        │
   │<────────────────────────┤                        │
   │ { access_token,         │                        │
   │   refresh_token }       │                        │
```

**Security:**
- ✅ Refresh token hashed in database
- ✅ One-time use (optionally)
- ✅ Revoked on logout
- ✅ Revoked on password change
- ✅ Scoped to organization
- ✅ Linked to session

---

### 5. Password Reset Flow

```
┌──────┐                ┌─────────┐              ┌──────────┐      ┌──────┐
│Client│                │ Backend │              │ Database │      │ SMTP │
└──┬───┘                └────┬────┘              └────┬─────┘      └──┬───┘
   │                         │                        │               │
   │ POST /auth/forgot-pwd   │                        │               │
   ├────────────────────────>│                        │               │
   │ { email }               │                        │               │
   │                         │                        │               │
   │                         │ Find user by email     │               │
   │                         ├───────────────────────>│               │
   │                         │<───────────────────────┤               │
   │                         │                        │               │
   │                         │ Generate reset token   │               │
   │                         │ (secure random)        │               │
   │                         │                        │               │
   │                         │ Hash and store token   │               │
   │                         ├───────────────────────>│               │
   │                         │<───────────────────────┤               │
   │                         │                        │               │
   │                         │ Send reset email       │               │
   │                         ├────────────────────────────────────────>│
   │                         │                        │               │
   │<────────────────────────┤                        │               │
   │ 200 OK                  │                        │               │
   │ "Email sent"            │                        │               │
   │                         │                        │               │
   │                         │                        │               │
   │ Click email link        │                        │               │
   │ POST /auth/reset-pwd    │                        │               │
   ├────────────────────────>│                        │               │
   │ { token, password }     │                        │               │
   │                         │                        │               │
   │                         │ Find reset token       │               │
   │                         ├───────────────────────>│               │
   │                         │<───────────────────────┤               │
   │                         │                        │               │
   │                         │ Verify token hash      │               │
   │                         │ Check not expired      │               │
   │                         │ Check not used         │               │
   │                         │                        │               │
   │                         │ Hash new password      │               │
   │                         │                        │               │
   │                         │ Update user password   │               │
   │                         ├───────────────────────>│               │
   │                         │<───────────────────────┤               │
   │                         │                        │               │
   │                         │ Mark token as used     │               │
   │                         ├───────────────────────>│               │
   │                         │                        │               │
   │                         │ Revoke all sessions    │               │
   │                         ├───────────────────────>│               │
   │                         │                        │               │
   │<────────────────────────┤                        │               │
   │ 200 OK                  │                        │               │
   │ "Password updated"      │                        │               │
```

**Security:**
- ✅ Token expires in 1 hour
- ✅ One-time use token
- ✅ Rate limited: 3 per hour per email
- ✅ All sessions revoked on reset
- ✅ Email sent even if user not found (timing attack prevention)

---

### 6. OAuth2 Authorization Code Flow (PKCE)

```
┌──────────┐        ┌─────────┐        ┌──────────┐        ┌────────┐
│3rd-Party │        │ Client  │        │  Auth    │        │Database│
│   App    │        │ Browser │        │  Server  │        │        │
└────┬─────┘        └────┬────┘        └────┬─────┘        └───┬────┘
     │                   │                   │                  │
     │ 1. Initiate OAuth │                   │                  │
     ├──────────────────>│                   │                  │
     │                   │                   │                  │
     │                   │ 2. Generate PKCE  │                  │
     │                   │    verifier +     │                  │
     │                   │    challenge      │                  │
     │                   │                   │                  │
     │                   │ 3. GET /oauth/authorize?            │
     │                   │    client_id=xxx                     │
     │                   │    redirect_uri=...                  │
     │                   │    code_challenge=...                │
     │                   │    state=random                      │
     │                   ├──────────────────>│                  │
     │                   │                   │                  │
     │                   │                   │ 4. Show login    │
     │                   │<──────────────────┤    page          │
     │                   │                   │                  │
     │                   │ 5. User login     │                  │
     │                   ├──────────────────>│                  │
     │                   │                   │                  │
     │                   │                   │ 6. Create auth   │
     │                   │                   │    code          │
     │                   │                   ├─────────────────>│
     │                   │                   │                  │
     │                   │ 7. Redirect with  │                  │
     │                   │    code + state   │                  │
     │                   │<──────────────────┤                  │
     │                   │                   │                  │
     │ 8. Callback       │                   │                  │
     │ ?code=xxx&state=..│                   │                  │
     │<──────────────────┤                   │                  │
     │                   │                   │                  │
     │ 9. POST /oauth/token                  │                  │
     │    code=xxx                           │                  │
     │    code_verifier=...                  │                  │
     │    client_id=...                      │                  │
     ├──────────────────────────────────────>│                  │
     │                   │                   │                  │
     │                   │                   │ 10. Validate code│
     │                   │                   ├─────────────────>│
     │                   │                   │<─────────────────┤
     │                   │                   │                  │
     │                   │                   │ 11. Verify PKCE  │
     │                   │                   │  challenge       │
     │                   │                   │                  │
     │                   │                   │ 12. Generate     │
     │                   │                   │  access token    │
     │                   │                   │                  │
     │ 13. Return tokens │                   │                  │
     │<──────────────────────────────────────┤                  │
     │ { access_token,   │                   │                  │
     │   refresh_token } │                   │                  │
```

**PKCE Security:**
1. Client generates random `code_verifier` (43-128 chars)
2. Client creates `code_challenge = SHA256(code_verifier)`
3. Authorization request includes `code_challenge`
4. Auth server stores challenge with auth code
5. Token request includes `code_verifier`
6. Server validates: `SHA256(verifier) == stored_challenge`
7. Only valid verifier can exchange code for tokens

---

### 7. Organization Invitation Flow

```
┌─────────┐         ┌─────────┐         ┌──────────┐       ┌──────┐
│ Admin   │         │ Backend │         │ Database │       │ SMTP │
└────┬────┘         └────┬────┘         └────┬─────┘       └──┬───┘
     │                   │                   │                │
     │ POST /orgs/:id/   │                   │                │
     │ members           │                   │                │
     ├──────────────────>│                   │                │
     │ { email, role_id }│                   │                │
     │                   │                   │                │
     │                   │ Check permission  │                │
     │                   │ "member:invite"   │                │
     │                   │                   │                │
     │                   │ Check if email    │                │
     │                   │ already exists    │                │
     │                   ├──────────────────>│                │
     │                   │<──────────────────┤                │
     │                   │                   │                │
     │                   │ Generate token    │                │
     │                   │ (secure random)   │                │
     │                   │                   │                │
     │                   │ Create membership │                │
     │                   │ (status: invited) │                │
     │                   ├──────────────────>│                │
     │                   │                   │                │
     │                   │ Send invite email │                │
     │                   ├───────────────────────────────────>│
     │                   │                   │                │
     │<──────────────────┤                   │                │
     │ 201 Created       │                   │                │
     │                   │                   │                │
     │                   │                   │                │
┌────┴────┐              │                   │                │
│ Invitee │              │                   │                │
└────┬────┘              │                   │                │
     │                   │                   │                │
     │ Click invite link │                   │                │
     │ POST /invitations/│                   │                │
     │ :token/accept     │                   │                │
     ├──────────────────>│                   │                │
     │                   │                   │                │
     │                   │ Validate token    │                │
     │                   ├──────────────────>│                │
     │                   │<──────────────────┤                │
     │                   │                   │                │
     │                   │ Check not expired │                │
     │                   │                   │                │
     │                   │ Update membership │                │
     │                   │ (status: active)  │                │
     │                   ├──────────────────>│                │
     │                   │                   │                │
     │<──────────────────┤                   │                │
     │ 200 OK            │                   │                │
```

---

## API Endpoint Groups

### Public Endpoints (No Auth Required)
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password
POST   /api/v1/auth/verify-email
POST   /api/v1/auth/resend-verification
GET    /api/v1/invitations/:token
GET    /health
```

### Authenticated Endpoints (JWT Required)
```
POST   /api/v1/auth/select-organization
POST   /api/v1/auth/create-organization
POST   /api/v1/auth/refresh
GET    /api/v1/user/profile
PUT    /api/v1/user/profile
POST   /api/v1/user/change-password
POST   /api/v1/user/logout
GET    /api/v1/user/organizations
```

### Organization Endpoints (Membership Required)
```
GET    /api/v1/organizations/:orgId
PUT    /api/v1/organizations/:orgId
DELETE /api/v1/organizations/:orgId
GET    /api/v1/organizations/:orgId/members
POST   /api/v1/organizations/:orgId/members
PUT    /api/v1/organizations/:orgId/members/:userId
DELETE /api/v1/organizations/:orgId/members/:userId
```

### RBAC Endpoints (Permission-based)
```
GET    /api/v1/organizations/:orgId/roles
POST   /api/v1/organizations/:orgId/roles
PUT    /api/v1/organizations/:orgId/roles/:roleId
DELETE /api/v1/organizations/:orgId/roles/:roleId
GET    /api/v1/organizations/:orgId/permissions
POST   /api/v1/organizations/:orgId/permissions
```

### Admin Endpoints (Superadmin Only)
```
GET    /api/v1/admin/users
PUT    /api/v1/admin/users/:userId/activate
PUT    /api/v1/admin/users/:userId/deactivate
DELETE /api/v1/admin/users/:userId
GET    /api/v1/admin/organizations
GET    /api/v1/admin/rbac/permissions
GET    /api/v1/admin/rbac/roles
POST   /api/v1/admin/rbac/roles
```

### OAuth2 Endpoints
```
GET    /api/v1/oauth/authorize
POST   /api/v1/oauth/token
GET    /api/v1/oauth/userinfo
POST   /api/v1/oauth/logout
GET    /api/v1/oauth/audit/authorizations (admin)
GET    /api/v1/oauth/audit/tokens (admin)
```

---

## Sample API Responses

### Success Response Structure
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response payload varies by endpoint
  }
}
```

### Error Response Structure
```json
{
  "success": false,
  "error_code": "VALIDATION_FAILED",
  "message": "Invalid request data",
  "errors": {
    "email": "Invalid email format",
    "password": "Password must be at least 8 characters"
  },
  "request_id": "7e00ef6e-34fd-428b-9871-52e6689793aa",
  "timestamp": "2025-11-23T06:33:09Z"
}
```

### POST /auth/register
**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "confirm_password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "User registered successfully. Please create or join an organization.",
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "+1234567890",
      "email_verified": false,
      "is_active": true,
      "is_admin": false,
      "created_at": "2025-11-23T10:00:00Z",
      "updated_at": "2025-11-23T10:00:00Z"
    }
  }
}
```

### POST /auth/login
**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200 OK) - Regular User:**
```json
{
  "success": true,
  "message": "Login successful. Please select an organization.",
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "email_verified": true,
      "is_active": true,
      "is_admin": false
    },
    "organizations": [
      {
        "organization_id": "eeacc427-b91d-4e38-a4a7-69375a1b8628",
        "organization_name": "Acme Corporation",
        "organization_slug": "acme-corp",
        "role": "owner",
        "status": "active",
        "joined_at": "2025-01-15T10:30:00Z"
      },
      {
        "organization_id": "f8d12345-1234-5678-9abc-def012345678",
        "organization_name": "Beta Inc",
        "organization_slug": "beta-inc",
        "role": "member",
        "status": "active",
        "joined_at": "2025-02-20T14:00:00Z"
      }
    ]
  }
}
```

**Response (200 OK) - Superadmin:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "admin-uuid",
      "email": "admin@blocksure.io",
      "is_admin": true
    },
    "token": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "rt_abc123def456...",
      "token_type": "Bearer",
      "expires_in": 900
    }
  }
}
```

### POST /auth/select-organization
**Request:**
```json
{
  "organization_id": "eeacc427-b91d-4e38-a4a7-69375a1b8628"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe"
    },
    "organization": {
      "organization_id": "eeacc427-b91d-4e38-a4a7-69375a1b8628",
      "organization_name": "Acme Corporation",
      "organization_slug": "acme-corp",
      "role": "owner",
      "status": "active"
    },
    "token": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjNlNDU2Ny1lODliLTEyZDMtYTQ1Ni00MjY2MTQxNzQwMDAiLCJvcmdfaWQiOiJlZWFjYzQyNy1iOTFkLTRlMzgtYTRhNy02OTM3NWExYjg2MjgiLCJyb2xlIjoib3duZXIiLCJwZXJtaXNzaW9ucyI6WyIqIl0sImV4cCI6MTcwMDAwMDAwMH0...",
      "refresh_token": "rt_def789ghi012...",
      "token_type": "Bearer",
      "expires_in": 900
    }
  }
}
```

### POST /auth/refresh
**Request:**
```json
{
  "refresh_token": "rt_abc123def456..."
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "rt_new789xyz...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

### GET /user/profile
**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "address": "123 Main St, City",
    "email_verified": true,
    "is_active": true,
    "is_admin": false,
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-11-20T14:22:00Z"
  }
}
```

### GET /organizations/{orgId}
**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "eeacc427-b91d-4e38-a4a7-69375a1b8628",
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "description": "Leading provider of enterprise solutions",
    "status": "active",
    "created_by": "123e4567-e89b-12d3-a456-426614174000",
    "owner": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "owner@acme.com",
      "name": "John Doe"
    },
    "member_count": 25,
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-11-20T14:22:00Z"
  }
}
```

### GET /organizations/{orgId}/members
**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "owner@acme.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "owner",
      "status": "active",
      "joined_at": "2025-01-15T10:30:00Z"
    },
    {
      "id": "234f5678-f89c-23e4-b567-537725285111",
      "email": "manager@acme.com",
      "first_name": "Jane",
      "last_name": "Smith",
      "role": "manager",
      "status": "active",
      "joined_at": "2025-02-01T09:15:00Z"
    },
    {
      "id": "345g6789-g90d-34f5-c678-648836396222",
      "email": "member@acme.com",
      "first_name": "Bob",
      "last_name": "Johnson",
      "role": "member",
      "status": "invited",
      "joined_at": null
    }
  ]
}
```

### POST /organizations/{orgId}/members (Invite)
**Request:**
```json
{
  "email": "newmember@example.com",
  "role": "member"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Invitation sent successfully",
  "data": {
    "invitation_id": "inv-123abc-456def",
    "email": "newmember@example.com",
    "organization_name": "Acme Corporation",
    "role_name": "Member",
    "expires_at": "2025-11-30T23:59:59Z",
    "status": "pending"
  }
}
```

### GET /invitations/{token} (Public)
**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "email": "newmember@example.com",
    "organization_name": "Acme Corporation",
    "role_name": "Member",
    "expires_at": "2025-11-30T23:59:59Z",
    "status": "pending"
  }
}
```

### POST /invitations/{token}/accept
**Response (200 OK):**
```json
{
  "success": true,
  "message": "Invitation accepted successfully",
  "data": {
    "organization_id": "eeacc427-b91d-4e38-a4a7-69375a1b8628",
    "role": "member",
    "status": "active"
  }
}
```

### GET /organizations/{orgId}/roles
**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "role-uuid-1",
      "name": "owner",
      "display_name": "Owner",
      "description": "Full access to organization",
      "is_system": true,
      "organization_id": null,
      "permissions": [
        {
          "id": "perm-uuid-1",
          "name": "*",
          "display_name": "All Permissions",
          "category": "system"
        }
      ],
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": "role-uuid-2",
      "name": "manager",
      "display_name": "Manager",
      "description": "Can manage team members and projects",
      "is_system": false,
      "organization_id": "eeacc427-b91d-4e38-a4a7-69375a1b8628",
      "permissions": [
        {
          "id": "perm-uuid-2",
          "name": "member:view",
          "display_name": "View Members",
          "category": "member"
        },
        {
          "id": "perm-uuid-3",
          "name": "member:invite",
          "display_name": "Invite Members",
          "category": "member"
        },
        {
          "id": "perm-uuid-4",
          "name": "project:create",
          "display_name": "Create Projects",
          "category": "project"
        }
      ],
      "created_at": "2025-02-15T10:00:00Z",
      "updated_at": "2025-11-20T14:30:00Z"
    }
  ]
}
```

### GET /organizations/{orgId}/permissions
**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "perm-uuid-1",
      "name": "member:view",
      "display_name": "View Members",
      "description": "View organization members and their details",
      "category": "member",
      "is_system": true,
      "organization_id": null,
      "created_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": "perm-uuid-2",
      "name": "member:invite",
      "display_name": "Invite Members",
      "description": "Send invitations to new members",
      "category": "member",
      "is_system": true,
      "organization_id": null,
      "created_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": "perm-uuid-3",
      "name": "report:export",
      "display_name": "Export Reports",
      "description": "Export reports to various formats",
      "category": "report",
      "is_system": false,
      "organization_id": "eeacc427-b91d-4e38-a4a7-69375a1b8628",
      "created_at": "2025-03-10T12:00:00Z"
    }
  ]
}
```

### GET /health
**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-23T10:30:00Z",
  "version": "1.0.0",
  "components": {
    "database": {
      "status": "healthy",
      "response_time_ms": 5
    },
    "redis": {
      "status": "healthy",
      "response_time_ms": 2
    },
    "email": {
      "status": "healthy"
    }
  }
}
```

### Error Examples

**400 Bad Request - Validation Error:**
```json
{
  "success": false,
  "error_code": "VALIDATION_FAILED",
  "message": "Invalid request data",
  "errors": {
    "email": "Invalid email format",
    "password": "Password must be at least 8 characters"
  },
  "request_id": "7e00ef6e-34fd-428b-9871-52e6689793aa",
  "timestamp": "2025-11-23T06:33:09Z"
}
```

**401 Unauthorized:**
```json
{
  "success": false,
  "error_code": "UNAUTHORIZED",
  "message": "Invalid credentials",
  "request_id": "8f11fg67-45ge-529c-9982-63f6689793bb",
  "timestamp": "2025-11-23T06:35:00Z"
}
```

**403 Forbidden - Missing Permission:**
```json
{
  "success": false,
  "error_code": "FORBIDDEN",
  "message": "You don't have permission to perform this action",
  "errors": {
    "required_permission": "member:invite",
    "user_permissions": ["member:view"]
  },
  "request_id": "9g22gh78-56hf-630d-aa93-74g7790804cc",
  "timestamp": "2025-11-23T06:40:00Z"
}
```

**404 Not Found:**
```json
{
  "success": false,
  "error_code": "NOT_FOUND",
  "message": "Resource not found",
  "request_id": "ah33hi89-67ig-741e-bb04-85h8801915dd",
  "timestamp": "2025-11-23T06:45:00Z"
}
```

**429 Too Many Requests:**
```json
{
  "success": false,
  "error_code": "RATE_LIMIT_EXCEEDED",
  "message": "Too many requests. Please try again later.",
  "errors": {
    "retry_after": 3600,
    "limit": 10,
    "window": "1 hour"
  },
  "request_id": "bi44ij90-78jh-852f-cc15-96i9912026ee",
  "timestamp": "2025-11-23T06:50:00Z"
}
```

**500 Internal Server Error:**
```json
{
  "success": false,
  "error_code": "INTERNAL_ERROR",
  "message": "An internal error occurred",
  "request_id": "cj55jk01-89ki-963g-dd26-a7j0a23137ff",
  "timestamp": "2025-11-23T06:55:00Z"
}
```

---

## Rate Limits

| Scope | Limit | Window |
|-------|-------|--------|
| Registration | 10 requests | 1 hour (per IP) |
| Login | 10 requests | 15 minutes (per IP) |
| Password Reset | 3 requests | 1 hour (per email) |
| Email Verification | 5 requests | 15 minutes (per email) |
| OAuth2 Token | 100 requests | 1 hour (per IP) |
| Token Refresh | 50 requests | 1 hour (per user) |
| API Calls | 1000 requests | 1 hour (per user) |

---

## JWT Token Structure

**Access Token Claims:**
```json
{
  "sub": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "org_id": "eeacc427-b91d-4e38-a4a7-69375a1b8628",
  "role": "owner",
  "permissions": ["*"],
  "iat": 1700000000,
  "exp": 1700000900,
  "iss": "blocksure-auth",
  "aud": "blocksure-api"
}
```

**Refresh Token Claims:**
```json
{
  "sub": "123e4567-e89b-12d3-a456-426614174000",
  "org_id": "eeacc427-b91d-4e38-a4a7-69375a1b8628",
  "token_id": "rt_abc123def456",
  "iat": 1700000000,
  "exp": 1702592000,
  "iss": "blocksure-auth"
}
```

---

**Last Updated**: November 23, 2025  
**API Version**: v1  
**OpenAPI Spec**: [OPENAPI_SPEC.yaml](./OPENAPI_SPEC.yaml)
