# API Workflows & Authentication Flows

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

**Last Updated**: November 18, 2025  
**API Version**: v1
