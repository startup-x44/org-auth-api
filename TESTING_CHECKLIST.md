# Auth-Service API Testing Checklist

## Prerequisites
- [ ] Start development environment: `./dev.sh dev`
- [ ] Verify database is running and migrations applied
- [ ] Verify Redis is running
- [ ] Have Postman/cURL ready with base URL: `http://localhost:8080/api/v1`
- [ ] Create test email accounts for verification

## 1. Public Authentication Routes

### 1.1 User Registration
**Endpoint:** `POST /auth/register`

**Success Cases:**
- [ ] Register new user with valid data
  ```json
  {
    "email": "test@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe"
  }
  ```
  Expected: 201 Created, user created, welcome email sent

- [ ] Register user with minimal required fields
  ```json
  {
    "email": "minimal@example.com",
    "password": "SecurePass123!"
  }
  ```
  Expected: 201 Created

**Failure Cases:**
- [ ] Register with existing email
  Expected: 400 Bad Request, "email already exists"

- [ ] Register with invalid email format
  Expected: 400 Bad Request, validation error

- [ ] Register with weak password (no uppercase/special chars)
  Expected: 400 Bad Request, password validation error

- [ ] Register with missing required fields
  Expected: 400 Bad Request, validation errors

### 1.2 User Login
**Endpoint:** `POST /auth/login`

**Success Cases:**
- [ ] Login with correct credentials
  ```json
  {
    "email": "test@example.com",
    "password": "SecurePass123!"
  }
  ```
  Expected: 200 OK, JWT tokens returned, user data

- [ ] Login after account activation
  Expected: 200 OK

**Failure Cases:**
- [ ] Login with wrong password
  Expected: 401 Unauthorized, failed attempt recorded

- [ ] Login with non-existent email
  Expected: 401 Unauthorized

- [ ] Login with inactive account
  Expected: 401 Unauthorized, "account not activated"

- [ ] Login with account lockout (5+ failed attempts)
  Expected: 429 Too Many Requests, account locked for 15 minutes

### 1.3 Token Refresh
**Endpoint:** `POST /auth/refresh`

**Success Cases:**
- [ ] Refresh with valid refresh token
  ```json
  {
    "refresh_token": "valid_refresh_token_here"
  }
  ```
  Expected: 200 OK, new access token

**Failure Cases:**
- [ ] Refresh with expired refresh token
  Expected: 401 Unauthorized

- [ ] Refresh with invalid refresh token
  Expected: 401 Unauthorized

- [ ] Refresh with malformed token
  Expected: 400 Bad Request

### 1.4 Forgot Password
**Endpoint:** `POST /auth/forgot-password`

**Success Cases:**
- [ ] Request password reset for existing user
  ```json
  {
    "email": "test@example.com"
  }
  ```
  Expected: 200 OK, reset email sent

**Failure Cases:**
- [ ] Request reset for non-existent email
  Expected: 200 OK (security: don't reveal if email exists)

- [ ] Request reset too frequently (rate limited)
  Expected: 429 Too Many Requests

### 1.5 Reset Password
**Endpoint:** `POST /auth/reset-password`

**Success Cases:**
- [ ] Reset password with valid token
  ```json
  {
    "token": "valid_reset_token",
    "password": "NewSecurePass123!"
  }
  ```
  Expected: 200 OK, password updated, old sessions invalidated

**Failure Cases:**
- [ ] Reset with expired token
  Expected: 400 Bad Request, "token expired"

- [ ] Reset with invalid token
  Expected: 400 Bad Request, "invalid token"

- [ ] Reset with weak password
  Expected: 400 Bad Request, password validation error

## 2. Protected User Routes

### 2.1 Get Profile
**Endpoint:** `GET /user/profile`
**Auth:** Bearer token required

**Success Cases:**
- [ ] Get own profile with valid token
  Expected: 200 OK, user profile data

**Failure Cases:**
- [ ] Access without token
  Expected: 401 Unauthorized

- [ ] Access with expired token
  Expected: 401 Unauthorized

- [ ] Access with invalid token
  Expected: 401 Unauthorized

### 2.2 Update Profile
**Endpoint:** `PUT /user/profile`
**Auth:** Bearer token required

**Success Cases:**
- [ ] Update profile fields
  ```json
  {
    "first_name": "Jane",
    "last_name": "Smith",
    "phone": "+1234567890"
  }
  ```
  Expected: 200 OK, profile updated

**Failure Cases:**
- [ ] Update with invalid data
  Expected: 400 Bad Request, validation errors

- [ ] Update without auth
  Expected: 401 Unauthorized

### 2.3 Change Password
**Endpoint:** `POST /user/change-password`
**Auth:** Bearer token required

**Success Cases:**
- [ ] Change password with correct current password
  ```json
  {
    "current_password": "SecurePass123!",
    "new_password": "NewSecurePass456!"
  }
  ```
  Expected: 200 OK, password changed, other sessions invalidated

**Failure Cases:**
- [ ] Change with wrong current password
  Expected: 400 Bad Request, "current password incorrect"

- [ ] Change to weak password
  Expected: 400 Bad Request, password validation error

- [ ] Change without auth
  Expected: 401 Unauthorized

### 2.4 Logout
**Endpoint:** `POST /user/logout`
**Auth:** Bearer token required

**Success Cases:**
- [ ] Logout with valid token
  Expected: 200 OK, token invalidated

**Failure Cases:**
- [ ] Logout without auth
  Expected: 401 Unauthorized

## 3. Organization Routes

### 3.1 Create Organization
**Endpoint:** `POST /organizations`
**Auth:** Bearer token required

**Success Cases:**
- [ ] Create organization as authenticated user
  ```json
  {
    "name": "Test Company",
    "slug": "test-company",
    "description": "A test organization"
  }
  ```
  Expected: 201 Created, organization created, user becomes admin member

**Failure Cases:**
- [ ] Create with duplicate slug
  Expected: 400 Bad Request, "slug already exists"

- [ ] Create with invalid slug format
  Expected: 400 Bad Request, validation error

- [ ] Create without auth
  Expected: 401 Unauthorized

### 3.2 List User Organizations
**Endpoint:** `GET /organizations`
**Auth:** Bearer token required

**Success Cases:**
- [ ] List organizations where user is member
  Expected: 200 OK, array of organizations

**Failure Cases:**
- [ ] List without auth
  Expected: 401 Unauthorized

### 3.3 Get Organization
**Endpoint:** `GET /organizations/{orgId}`
**Auth:** Bearer token required, membership required

**Success Cases:**
- [ ] Get organization as member
  Expected: 200 OK, organization details

**Failure Cases:**
- [ ] Get organization as non-member
  Expected: 403 Forbidden

- [ ] Get non-existent organization
  Expected: 404 Not Found

- [ ] Get without auth
  Expected: 401 Unauthorized

### 3.4 Update Organization
**Endpoint:** `PUT /organizations/{orgId}`
**Auth:** Bearer token required, org admin required

**Success Cases:**
- [ ] Update organization as admin
  ```json
  {
    "name": "Updated Company",
    "description": "Updated description"
  }
  ```
  Expected: 200 OK, organization updated

**Failure Cases:**
- [ ] Update as regular member
  Expected: 403 Forbidden

- [ ] Update non-existent organization
  Expected: 404 Not Found

- [ ] Update without auth
  Expected: 401 Unauthorized

### 3.5 Delete Organization
**Endpoint:** `DELETE /organizations/{orgId}`
**Auth:** Bearer token required, org admin required

**Success Cases:**
- [ ] Delete organization as admin
  Expected: 200 OK, organization deleted

**Failure Cases:**
- [ ] Delete as regular member
  Expected: 403 Forbidden

- [ ] Delete non-existent organization
  Expected: 404 Not Found

### 3.6 List Organization Members
**Endpoint:** `GET /organizations/{orgId}/members`
**Auth:** Bearer token required, membership required

**Success Cases:**
- [ ] List members as organization member
  Expected: 200 OK, array of members with roles

**Failure Cases:**
- [ ] List members as non-member
  Expected: 403 Forbidden

### 3.7 Invite User to Organization
**Endpoint:** `POST /organizations/{orgId}/members`
**Auth:** Bearer token required, org admin required

**Success Cases:**
- [ ] Invite user as admin
  ```json
  {
    "email": "newuser@example.com",
    "role": "rto"
  }
  ```
  Expected: 201 Created, invitation created, email sent

**Failure Cases:**
- [ ] Invite as regular member
  Expected: 403 Forbidden

- [ ] Invite already invited user
  Expected: 400 Bad Request, "user already invited"

- [ ] Invite with invalid role
  Expected: 400 Bad Request, validation error

### 3.8 Update Membership
**Endpoint:** `PUT /organizations/{orgId}/members/{userId}`
**Auth:** Bearer token required, org admin required

**Success Cases:**
- [ ] Update member role as admin
  ```json
  {
    "role": "admin"
  }
  ```
  Expected: 200 OK, membership updated

**Failure Cases:**
- [ ] Update as regular member
  Expected: 403 Forbidden

- [ ] Update to invalid role
  Expected: 400 Bad Request

### 3.9 Remove Member
**Endpoint:** `DELETE /organizations/{orgId}/members/{userId}`
**Auth:** Bearer token required, org admin required

**Success Cases:**
- [ ] Remove member as admin
  Expected: 200 OK, membership removed

**Failure Cases:**
- [ ] Remove as regular member
  Expected: 403 Forbidden

- [ ] Remove self (last admin)
  Expected: 400 Bad Request, "cannot remove last admin"

### 3.10 Get Organization Invitations
**Endpoint:** `GET /organizations/{orgId}/invitations`
**Auth:** Bearer token required, org admin required

**Success Cases:**
- [ ] Get invitations as admin
  Expected: 200 OK, array of pending invitations

**Failure Cases:**
- [ ] Get invitations as regular member
  Expected: 403 Forbidden

### 3.11 Cancel Invitation
**Endpoint:** `DELETE /organizations/{orgId}/invitations/{invitationId}`
**Auth:** Bearer token required, org admin required

**Success Cases:**
- [ ] Cancel invitation as admin
  Expected: 200 OK, invitation cancelled

**Failure Cases:**
- [ ] Cancel as regular member
  Expected: 403 Forbidden

### 3.12 Accept Invitation
**Endpoint:** `POST /invitations/{token}/accept`
**Auth:** Bearer token required (user must be logged in)

**Success Cases:**
- [ ] Accept invitation with valid token
  Expected: 200 OK, user added to organization

**Failure Cases:**
- [ ] Accept with expired token
  Expected: 400 Bad Request, "invitation expired"

- [ ] Accept with invalid token
  Expected: 404 Not Found

- [ ] Accept without being logged in
  Expected: 401 Unauthorized

### 3.13 Get Invitation Details
**Endpoint:** `GET /invitations/{token}`
**Auth:** None (public)

**Success Cases:**
- [ ] Get invitation details with valid token
  Expected: 200 OK, invitation info (org name, inviter, role)

**Failure Cases:**
- [ ] Get with expired token
  Expected: 404 Not Found

- [ ] Get with invalid token
  Expected: 404 Not Found

## 4. Admin Routes (Superadmin Only)

### 4.1 List Users
**Endpoint:** `GET /admin/users`
**Auth:** Bearer token required, superadmin required

**Success Cases:**
- [ ] List all users as superadmin
  Expected: 200 OK, paginated user list

**Failure Cases:**
- [ ] List users as regular user
  Expected: 403 Forbidden

- [ ] List users without auth
  Expected: 401 Unauthorized

### 4.2 Activate User
**Endpoint:** `PUT /admin/users/{userId}/activate`
**Auth:** Bearer token required, superadmin required

**Success Cases:**
- [ ] Activate user as superadmin
  Expected: 200 OK, user activated

**Failure Cases:**
- [ ] Activate as regular user
  Expected: 403 Forbidden

- [ ] Activate non-existent user
  Expected: 404 Not Found

### 4.3 Deactivate User
**Endpoint:** `PUT /admin/users/{userId}/deactivate`
**Auth:** Bearer token required, superadmin required

**Success Cases:**
- [ ] Deactivate user as superadmin
  Expected: 200 OK, user deactivated, sessions invalidated

**Failure Cases:**
- [ ] Deactivate as regular user
  Expected: 403 Forbidden

### 4.4 Delete User
**Endpoint:** `DELETE /admin/users/{userId}`
**Auth:** Bearer token required, superadmin required

**Success Cases:**
- [ ] Delete user as superadmin
  Expected: 200 OK, user and all data deleted

**Failure Cases:**
- [ ] Delete as regular user
  Expected: 403 Forbidden

## 5. Middleware Testing

### 5.1 Authentication Middleware
- [ ] All protected routes reject requests without Authorization header
- [ ] All protected routes reject requests with invalid JWT
- [ ] All protected routes reject requests with expired JWT
- [ ] Routes accept valid Bearer tokens

### 5.2 Admin Middleware
- [ ] Admin routes reject requests from non-superadmin users
- [ ] Admin routes accept requests from superadmin users
- [ ] Admin middleware properly checks `is_superadmin` claim

### 5.3 Organization Middleware
- [ ] Organization routes require valid organization ID in URL
- [ ] MembershipRequired allows access to organization members
- [ ] OrgAdminRequired restricts access to organization admins only
- [ ] Middleware properly validates organization existence and status

### 5.4 CORS Middleware
- [ ] Preflight OPTIONS requests handled correctly
- [ ] Allowed origins accepted
- [ ] Disallowed origins rejected
- [ ] Proper CORS headers set

### 5.5 CSRF Protection
- [ ] State-changing requests require valid CSRF token
- [ ] CSRF tokens validated on POST/PUT/DELETE requests
- [ ] CSRF tokens properly rotated

### 5.6 Rate Limiting
- [ ] Auth endpoints rate limited (login, register, etc.)
- [ ] Password reset requests rate limited
- [ ] Failed login attempts tracked and limited

### 5.7 Security Headers
- [ ] Security headers present: X-Frame-Options, X-Content-Type-Options, etc.
- [ ] HTTPS redirection enforced in production
- [ ] Secure cookie flags set

## 6. Database & Redis Integration

### 6.1 Database Operations
- [ ] User registration creates user record
- [ ] Organization creation creates organization and membership records
- [ ] Invitation acceptance creates membership record
- [ ] User deletion cascades to related records
- [ ] Failed login attempts recorded in database

### 6.2 Redis Caching
- [ ] Session data stored in Redis
- [ ] Account lockout status cached in Redis
- [ ] Rate limiting counters maintained in Redis
- [ ] Cache invalidation on logout/password change

### 6.3 Email Integration
- [ ] Welcome email sent on registration
- [ ] Password reset email sent with valid token
- [ ] Organization invitation email sent with accept link
- [ ] Email templates properly formatted
- [ ] Email sending doesn't block API responses

## 7. Error Handling & Edge Cases

### 7.1 Input Validation
- [ ] All endpoints validate required fields
- [ ] Email format validation
- [ ] Password strength requirements
- [ ] UUID format validation for IDs
- [ ] Slug format validation for organizations

### 7.2 Concurrency
- [ ] Multiple simultaneous logins handled correctly
- [ ] Concurrent organization operations don't cause race conditions
- [ ] Database transactions used for complex operations

### 7.3 Data Integrity
- [ ] Foreign key constraints maintained
- [ ] Unique constraints enforced (emails, slugs)
- [ ] Data consistency across related tables

### 7.4 Performance
- [ ] Database queries optimized with proper indexes
- [ ] Pagination implemented for large result sets
- [ ] Redis caching reduces database load

## 8. Integration Testing

### 8.1 Complete User Flows
- [ ] User registration → email verification → login → profile management
- [ ] Organization creation → member invitation → acceptance → role management
- [ ] Password reset flow: request → email → reset → login
- [ ] Admin user management: activate/deactivate/delete users

### 8.2 Cross-Service Integration
- [ ] Database migrations run correctly
- [ ] Redis connectivity maintained
- [ ] Email service integration working
- [ ] External API calls handled properly

### 8.3 Load Testing
- [ ] Concurrent user registrations handled
- [ ] High-frequency login attempts managed
- [ ] Large organization member lists paginated efficiently

## 9. Security Audit

### 9.1 Authentication Security
- [ ] Passwords properly hashed with Argon2id
- [ ] JWT tokens have reasonable expiration times
- [ ] Refresh tokens properly invalidated on logout
- [ ] No sensitive data leaked in responses

### 9.2 Authorization Security
- [ ] Role-based access properly enforced
- [ ] Superadmin permissions correctly scoped
- [ ] Organization boundaries maintained
- [ ] No privilege escalation possible

### 9.3 Data Protection
- [ ] PII properly handled and stored
- [ ] Database queries prevent SQL injection
- [ ] Input sanitization prevents XSS
- [ ] Secure headers prevent common attacks

---

## Testing Tools & Commands

**Start Services:**
```bash
./dev.sh dev    # Start all services
./dev.sh test   # Run test suite
```

**Database Operations:**
```bash
# Check migrations
go run cmd/migrate/main.go status

# Reset database for testing
go run cmd/migrate/main.go reset
```

**API Testing with cURL:**
```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123!"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123!"}'

# Access protected route
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Postman Collection:**
- Import `docs/postman-collection.json`
- Set environment variables for base URL and tokens
- Run collection tests in sequence</content>
<parameter name="filePath">/Users/niloflora/fligno/blocksure/abc/auth-service/TESTING_CHECKLIST.md