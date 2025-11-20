# Appendix: Detailed Flow Explanations

This appendix provides step-by-step breakdowns of complex workflows in the authentication service, with detailed explanations of each step, security considerations, and edge cases.

---

## Table of Contents
1. [Complete Authentication Flow (End-to-End)](#1-complete-authentication-flow-end-to-end)
2. [OAuth2 Authorization Code Flow with PKCE](#2-oauth2-authorization-code-flow-with-pkce)
3. [RBAC Permission Check Flow](#3-rbac-permission-check-flow)
4. [Multi-Tenant Session Management](#4-multi-tenant-session-management)
5. [Password Reset Security Flow](#5-password-reset-security-flow)
6. [Organization Invitation & Onboarding](#6-organization-invitation--onboarding)
7. [Token Refresh & Session Extension](#7-token-refresh--session-extension)

---

## 1. Complete Authentication Flow (End-to-End)

**Scenario:** New user signs up, verifies email, logs in, selects organization, and makes first API call.

### Step-by-Step Breakdown

#### Step 1: User Registration
```
User Action: Fill registration form
POST /api/v1/auth/register
{
  "email": "john@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Backend Processing:**

1.1. **Input Validation**
```go
// Validate email format (RFC 5322)
if !isValidEmail(req.Email) {
    return error("invalid email format")
}

// Validate password length
if len(req.Password) < 8 {
    return error("password too short")
}
```

1.2. **Check Duplicate Email**
```sql
SELECT id FROM users WHERE email = 'john@example.com'
-- Result: empty (no duplicate)
```

1.3. **Hash Password (Argon2id)**
```go
// Generate random salt (16 bytes)
salt := crypto.RandBytes(16)

// Hash password
hash := argon2.IDKey(
    []byte(password),
    salt,
    3,      // iterations
    64*1024, // memory (64MB)
    4,      // threads
    32,     // key length
)

// Store format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
hashedPassword := encodeHash(salt, hash)
```

**Time:** ~50ms (intentionally slow to resist brute force)

1.4. **Create User Record**
```sql
INSERT INTO users (
    id,
    email,
    password_hash,
    first_name,
    last_name,
    status,
    created_at
) VALUES (
    gen_random_uuid(),
    'john@example.com',
    '$argon2id$v=19$m=65536,t=3,p=4$...',
    'John',
    'Doe',
    'pending',  -- Not verified yet
    NOW()
)
RETURNING id
```

1.5. **Generate Verification Token**
```go
// Secure random token (32 bytes = 256 bits)
token := hex.EncodeToString(crypto.RandBytes(32))
// Example: "a3f7e8d9c1b2...64 chars"

// Store with expiration
INSERT INTO verification_tokens (
    user_id,
    token,
    type,
    expires_at
) VALUES (
    user_id,
    token,
    'email_verification',
    NOW() + INTERVAL '24 hours'
)
```

1.6. **Send Verification Email**
```go
emailService.Send(Email{
    To:       "john@example.com",
    Subject:  "Verify your email",
    Template: "verify-email",
    Data: {
        Name: "John",
        VerificationLink: "https://app.example.com/verify?token=" + token,
    },
})
```

**Response to Client:**
```json
{
  "success": true,
  "message": "Registration successful. Please check your email to verify your account.",
  "user": {
    "id": "uuid",
    "email": "john@example.com",
    "status": "pending"
  }
}
```

**Security Notes:**
- ✅ Password never stored in plaintext
- ✅ Email sent even if user exists (timing attack prevention)
- ✅ Token is cryptographically random (not guessable)
- ✅ Token expires in 24 hours

---

#### Step 2: Email Verification

**User Action:** Click link in email

```
GET /api/v1/auth/verify-email?token=a3f7e8d9c1b2...
```

**Backend Processing:**

2.1. **Validate Token**
```sql
SELECT 
    user_id,
    expires_at,
    used_at
FROM verification_tokens
WHERE token = 'a3f7e8d9c1b2...'
  AND type = 'email_verification'
```

2.2. **Check Expiration**
```go
if token.ExpiresAt.Before(time.Now()) {
    return error("Token expired. Please request a new verification email.")
}
```

2.3. **Check Already Used**
```go
if token.UsedAt != nil {
    return error("Token already used.")
}
```

2.4. **Mark Email as Verified**
```sql
BEGIN TRANSACTION;

-- Update user
UPDATE users
SET 
    status = 'active',
    email_verified_at = NOW()
WHERE id = token.user_id;

-- Mark token as used
UPDATE verification_tokens
SET used_at = NOW()
WHERE token = 'a3f7e8d9c1b2...';

COMMIT;
```

2.5. **Create Audit Log**
```sql
INSERT INTO audit_logs (
    user_id,
    action,
    resource_type,
    ip_address,
    user_agent,
    created_at
) VALUES (
    user_id,
    'email_verified',
    'user',
    request.IP,
    request.UserAgent,
    NOW()
)
```

**Response:**
```json
{
  "success": true,
  "message": "Email verified successfully. You can now log in."
}
```

**Frontend:** Redirect to login page

---

#### Step 3: User Login

**User Action:** Enter credentials

```
POST /api/v1/auth/login
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Backend Processing:**

3.1. **Rate Limit Check**
```go
// Redis key: "ratelimit:login:192.168.1.100"
attempts := redis.Incr("ratelimit:login:" + request.IP)
if attempts > 20 {
    return error("Too many login attempts. Try again in 15 minutes.")
}
redis.Expire("ratelimit:login:" + request.IP, 15 * time.Minute)
```

3.2. **Find User**
```sql
SELECT 
    id,
    email,
    password_hash,
    status,
    failed_login_attempts,
    locked_until
FROM users
WHERE email = 'john@example.com'
```

3.3. **Check Account Status**
```go
if user.Status == "pending" {
    return error("Please verify your email before logging in.")
}

if user.Status == "suspended" {
    return error("Account suspended. Contact support.")
}

if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
    return error("Account locked due to too many failed attempts.")
}
```

3.4. **Verify Password**
```go
// Parse stored hash
salt, hash := parseArgon2Hash(user.PasswordHash)

// Hash provided password with same salt
providedHash := argon2.IDKey(
    []byte(password),
    salt,
    3,      // iterations
    64*1024, // memory
    4,      // threads
    32,     // key length
)

// Constant-time comparison (timing attack prevention)
if !subtle.ConstantTimeCompare(hash, providedHash) {
    // Increment failed attempts
    UPDATE users
    SET failed_login_attempts = failed_login_attempts + 1
    WHERE id = user.ID
    
    // Lock account after 5 failed attempts
    if user.FailedLoginAttempts + 1 >= 5 {
        UPDATE users
        SET locked_until = NOW() + INTERVAL '30 minutes'
        WHERE id = user.ID
    }
    
    return error("Invalid credentials")
}
```

**Time:** ~50ms (Argon2 verification)

3.5. **Reset Failed Attempts**
```sql
UPDATE users
SET 
    failed_login_attempts = 0,
    last_login_at = NOW()
WHERE id = user.ID
```

3.6. **Get User's Organizations**
```sql
SELECT 
    o.id,
    o.name,
    o.logo_url,
    uo.role_id,
    r.name AS role_name
FROM organizations o
JOIN user_organizations uo ON o.id = uo.organization_id
JOIN roles r ON uo.role_id = r.id
WHERE uo.user_id = user.ID
  AND uo.status = 'active'
ORDER BY o.name
```

**Response:**
```json
{
  "success": true,
  "user": {
    "id": "uuid",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe"
  },
  "organizations": [
    {
      "id": "org-uuid-1",
      "name": "Acme Corp",
      "role": "Owner"
    },
    {
      "id": "org-uuid-2",
      "name": "Startup Inc",
      "role": "Member"
    }
  ],
  "needs_organization_selection": true
}
```

**Security Notes:**
- ✅ Constant-time password comparison (timing attack prevention)
- ✅ Account lockout after 5 failed attempts
- ✅ Rate limiting (20 attempts per 15 min)
- ✅ Failed attempts logged for audit

---

#### Step 4: Organization Selection

**User Action:** Select organization from list

```
POST /api/v1/auth/select-organization
{
  "organization_id": "org-uuid-1"
}
```

**Backend Processing:**

4.1. **Verify Membership**
```sql
SELECT 
    role_id,
    status
FROM user_organizations
WHERE user_id = authenticated_user.ID
  AND organization_id = 'org-uuid-1'
```

4.2. **Check Membership Status**
```go
if membership == nil {
    return error("You are not a member of this organization")
}

if membership.Status != "active" {
    return error("Your membership is not active")
}
```

4.3. **Get Role & Permissions**
```sql
-- Get role
SELECT 
    r.id,
    r.name,
    r.is_system
FROM roles r
WHERE r.id = membership.role_id

-- Get permissions (including system permissions if role is system)
SELECT 
    p.id,
    p.name,
    p.resource,
    p.action
FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.id
WHERE rp.role_id = membership.role_id
  AND (
    p.organization_id = 'org-uuid-1'
    OR p.organization_id IS NULL  -- System permissions
  )
```

**Example Permissions:**
```json
[
  { "name": "user:read", "resource": "user", "action": "read" },
  { "name": "user:write", "resource": "user", "action": "write" },
  { "name": "role:manage", "resource": "role", "action": "manage" },
  // ... 20 more permissions
]
```

4.4. **Create Session**
```sql
INSERT INTO sessions (
    id,
    user_id,
    organization_id,
    ip_address,
    user_agent,
    device_name,
    last_activity_at,
    expires_at,
    created_at
) VALUES (
    gen_random_uuid(),
    user.ID,
    'org-uuid-1',
    '192.168.1.100',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...',
    'Chrome on macOS',
    NOW(),
    NOW() + INTERVAL '30 days',
    NOW()
)
RETURNING id
```

4.5. **Generate JWT (Access Token)**
```go
claims := jwt.MapClaims{
    "sub":             user.ID,
    "email":           user.Email,
    "organization_id": "org-uuid-1",
    "role_id":         role.ID,
    "role_name":       role.Name,
    "permissions":     permissionNames, // ["user:read", "user:write", ...]
    "iat":             time.Now().Unix(),
    "exp":             time.Now().Add(1 * time.Hour).Unix(),
    "iss":             "auth-service",
    "aud":             "api",
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
accessToken, _ := token.SignedString([]byte(jwtSecret))
```

**Access Token (decoded):**
```json
{
  "sub": "user-uuid",
  "email": "john@example.com",
  "organization_id": "org-uuid-1",
  "role_id": "role-uuid",
  "role_name": "Owner",
  "permissions": ["user:read", "user:write", "role:manage", ...],
  "iat": 1700000000,
  "exp": 1700003600,
  "iss": "auth-service",
  "aud": "api"
}
```

4.6. **Generate Refresh Token**
```go
// Opaque token (not JWT)
refreshToken := base64.URLEncoding.EncodeToString(crypto.RandBytes(32))

// Hash before storing (Argon2id)
hashedRefreshToken := argon2.Hash(refreshToken)

// Store in database
UPDATE sessions
SET refresh_token = hashedRefreshToken
WHERE id = session.ID
```

4.7. **Create Audit Log**
```sql
INSERT INTO audit_logs (
    user_id,
    organization_id,
    action,
    metadata,
    created_at
) VALUES (
    user.ID,
    'org-uuid-1',
    'login',
    jsonb_build_object(
        'device', 'Chrome on macOS',
        'ip', '192.168.1.100',
        'session_id', session.ID
    ),
    NOW()
)
```

**Response to Client:**
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "random-base64-string...",
  "expires_in": 3600,
  "organization": {
    "id": "org-uuid-1",
    "name": "Acme Corp"
  },
  "user": {
    "id": "user-uuid",
    "email": "john@example.com",
    "role": "Owner",
    "permissions": ["user:read", "user:write", ...]
  }
}
```

**Frontend:**
```typescript
// Store tokens
localStorage.setItem('access_token', response.access_token)
localStorage.setItem('refresh_token', response.refresh_token)
localStorage.setItem('organization', JSON.stringify(response.organization))

// Redirect to dashboard
router.push('/dashboard')
```

---

#### Step 5: First API Call

**User Action:** View dashboard (fetches user list)

```
GET /api/v1/organizations/org-uuid-1/members
Headers:
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Backend Processing:**

5.1. **Extract JWT from Header**
```go
authHeader := request.Header.Get("Authorization")
// "Bearer eyJhbGciOiJI..."

tokenString := strings.TrimPrefix(authHeader, "Bearer ")
```

5.2. **Validate JWT**
```go
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    // Verify signing method
    if token.Method != jwt.SigningMethodHS256 {
        return nil, errors.New("invalid signing method")
    }
    return []byte(jwtSecret), nil
})

if err != nil {
    return 401, "Invalid token"
}

if !token.Valid {
    return 401, "Token expired"
}
```

5.3. **Extract Claims**
```go
claims := token.Claims.(jwt.MapClaims)

userID := claims["sub"].(string)
organizationID := claims["organization_id"].(string)
permissions := claims["permissions"].([]interface{})
```

5.4. **Verify Organization Access**
```go
// Check token's org matches requested resource's org
if organizationID != "org-uuid-1" {
    return 403, "Forbidden: Wrong organization"
}
```

5.5. **Check Permission**
```go
requiredPermission := "user:read"

hasPermission := false
for _, perm := range permissions {
    if perm.(string) == requiredPermission {
        hasPermission = true
        break
    }
}

if !hasPermission {
    // Log permission denial
    INSERT INTO audit_logs (
        user_id,
        organization_id,
        action,
        resource_type,
        metadata
    ) VALUES (
        userID,
        organizationID,
        'permission_denied',
        'user',
        jsonb_build_object(
            'required_permission', 'user:read',
            'user_permissions', permissions
        )
    )
    
    return 403, "Forbidden: Missing permission 'user:read'"
}
```

5.6. **Query Database (Multi-Tenant Filtered)**
```sql
SELECT 
    u.id,
    u.email,
    u.first_name,
    u.last_name,
    uo.role_id,
    r.name AS role_name,
    uo.status
FROM users u
JOIN user_organizations uo ON u.id = uo.user_id
JOIN roles r ON uo.role_id = r.id
WHERE uo.organization_id = 'org-uuid-1'  -- ✅ Organization filter
  AND uo.status = 'active'
ORDER BY u.created_at DESC
```

**Response:**
```json
{
  "members": [
    {
      "id": "user-uuid-1",
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "Owner",
      "status": "active"
    },
    {
      "id": "user-uuid-2",
      "email": "jane@example.com",
      "first_name": "Jane",
      "last_name": "Smith",
      "role": "Admin",
      "status": "active"
    }
  ],
  "total": 2
}
```

**Security Validations:**
- ✅ JWT signature verified
- ✅ JWT expiration checked
- ✅ Organization context validated
- ✅ Permission checked
- ✅ Multi-tenant filter applied
- ✅ Permission denial logged

---

## 2. OAuth2 Authorization Code Flow with PKCE

**Scenario:** Third-party app "Dashboard Pro" wants to access user's data.

### PKCE (Proof Key for Code Exchange)

**Purpose:** Prevent authorization code interception attack (especially on mobile/SPA)

**How it works:**
1. Client generates random `code_verifier`
2. Client hashes verifier → `code_challenge`
3. Authorization request includes `code_challenge`
4. Token request includes `code_verifier`
5. Server validates: `hash(verifier) == stored_challenge`

---

### Step-by-Step Flow

#### Step 1: Client App Generates PKCE Values

**Client (Dashboard Pro) Code:**
```typescript
// 1. Generate code verifier (random 43-128 chars)
function generateCodeVerifier(): string {
    const randomBytes = crypto.getRandomValues(new Uint8Array(32))
    return base64URLEncode(randomBytes)
    // Example: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
}

// 2. Generate code challenge (SHA256 hash of verifier)
async function generateCodeChallenge(verifier: string): Promise<string> {
    const encoder = new TextEncoder()
    const data = encoder.encode(verifier)
    const hash = await crypto.subtle.digest('SHA-256', data)
    return base64URLEncode(new Uint8Array(hash))
    // Example: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
}

// Usage
const codeVerifier = generateCodeVerifier()
const codeChallenge = await generateCodeChallenge(codeVerifier)

// Store verifier in session storage (needed later)
sessionStorage.setItem('code_verifier', codeVerifier)
```

---

#### Step 2: Authorization Request

**Client redirects user to:**
```
GET https://auth.example.com/api/v1/oauth/authorize?
  client_id=dashboard-pro-client-id&
  redirect_uri=https://dashboardpro.com/callback&
  response_type=code&
  scope=user:read user:write&
  state=random-csrf-token&
  code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&
  code_challenge_method=S256
```

**Backend Processing:**

2.1. **Validate Client App**
```sql
SELECT 
    id,
    name,
    client_secret,
    redirect_uris,
    allowed_scopes,
    is_confidential
FROM client_apps
WHERE client_id = 'dashboard-pro-client-id'
  AND is_active = true
```

2.2. **Validate Redirect URI**
```go
if !contains(clientApp.RedirectUris, request.RedirectUri) {
    return error("Invalid redirect_uri")
}
```

2.3. **Validate Scopes**
```go
requestedScopes := strings.Split(request.Scope, " ")
for _, scope := range requestedScopes {
    if !contains(clientApp.AllowedScopes, scope) {
        return error("Scope '" + scope + "' not allowed")
    }
}
```

2.4. **Check User Authentication**
```go
if !isAuthenticated(request) {
    // Redirect to login with return URL
    redirectTo("/login?return_to=" + url.Encode(request.URL))
}
```

2.5. **Show Consent Screen**
```html
<!-- Consent Page -->
<h2>Dashboard Pro wants to access your account</h2>
<p>This application will be able to:</p>
<ul>
  <li>Read your profile information</li>
  <li>Modify your user settings</li>
</ul>
<button>Allow</button>
<button>Deny</button>
```

---

#### Step 3: User Approves

**User clicks "Allow"**

```
POST /api/v1/oauth/authorize/consent
{
  "client_id": "dashboard-pro-client-id",
  "scopes": ["user:read", "user:write"],
  "approved": true
}
```

**Backend Processing:**

3.1. **Generate Authorization Code**
```go
authCode := base64.URLEncoding.EncodeToString(crypto.RandBytes(32))
// Example: "SplxlOBeZQQYbYS6WxSbIA"
```

3.2. **Store Authorization Code**
```sql
INSERT INTO oauth_authorization_codes (
    code,
    client_id,
    user_id,
    organization_id,
    redirect_uri,
    scopes,
    code_challenge,
    code_challenge_method,
    expires_at,
    created_at
) VALUES (
    'SplxlOBeZQQYbYS6WxSbIA',
    'dashboard-pro-client-id',
    current_user.ID,
    current_user.OrganizationID,
    'https://dashboardpro.com/callback',
    ARRAY['user:read', 'user:write'],
    'E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM',
    'S256',
    NOW() + INTERVAL '10 minutes',  -- Short expiration
    NOW()
)
```

**Expiration:** 10 minutes (authorization codes are short-lived)

3.3. **Redirect to Client's Callback**
```
HTTP 302 Redirect
Location: https://dashboardpro.com/callback?
  code=SplxlOBeZQQYbYS6WxSbIA&
  state=random-csrf-token
```

---

#### Step 4: Client Exchanges Code for Tokens

**Client (Dashboard Pro) Backend:**
```
POST https://auth.example.com/api/v1/oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code=SplxlOBeZQQYbYS6WxSbIA&
redirect_uri=https://dashboardpro.com/callback&
client_id=dashboard-pro-client-id&
client_secret=super-secret-key&  (if confidential client)
code_verifier=dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk
```

**Backend Processing:**

4.1. **Validate Authorization Code**
```sql
SELECT 
    user_id,
    organization_id,
    client_id,
    redirect_uri,
    scopes,
    code_challenge,
    code_challenge_method,
    expires_at,
    used_at
FROM oauth_authorization_codes
WHERE code = 'SplxlOBeZQQYbYS6WxSbIA'
```

4.2. **Check Expiration**
```go
if authCode.ExpiresAt.Before(time.Now()) {
    return error("Authorization code expired")
}
```

4.3. **Check Already Used (Prevent Replay)**
```go
if authCode.UsedAt != nil {
    // ⚠️ SECURITY: Code already used - potential attack
    // Revoke ALL tokens issued with this code
    DELETE FROM oauth_access_tokens
    WHERE authorization_code_id = authCode.ID
    
    return error("Authorization code already used")
}
```

4.4. **Verify PKCE Challenge**
```go
// Hash provided verifier
providedChallenge := sha256.Sum256([]byte(request.CodeVerifier))
providedChallengeB64 := base64.URLEncoding.EncodeToString(providedChallenge[:])

// Compare with stored challenge (constant-time)
if !subtle.ConstantTimeCompare(
    []byte(providedChallengeB64),
    []byte(authCode.CodeChallenge),
) {
    return error("Invalid code_verifier")
}
```

**Security:** This prevents attackers who intercept the code from using it (they don't have the verifier)

4.5. **Verify Client Credentials (Confidential Clients)**
```go
if clientApp.IsConfidential {
    if request.ClientSecret != clientApp.ClientSecret {
        return error("Invalid client_secret")
    }
}
```

4.6. **Verify Redirect URI Matches**
```go
if request.RedirectUri != authCode.RedirectUri {
    return error("redirect_uri mismatch")
}
```

4.7. **Generate Access Token (JWT)**
```go
accessTokenClaims := jwt.MapClaims{
    "sub":             authCode.UserID,
    "client_id":       authCode.ClientID,
    "organization_id": authCode.OrganizationID,
    "scope":           strings.Join(authCode.Scopes, " "),
    "iat":             time.Now().Unix(),
    "exp":             time.Now().Add(1 * time.Hour).Unix(),
    "iss":             "auth-service",
    "aud":             "api",
}

accessToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims).
    SignedString([]byte(jwtSecret))
```

4.8. **Generate Refresh Token**
```go
refreshToken := base64.URLEncoding.EncodeToString(crypto.RandBytes(32))
hashedRefreshToken := argon2.Hash(refreshToken)
```

4.9. **Store Tokens in Database**
```sql
-- Store access token
INSERT INTO oauth_access_tokens (
    token,
    client_id,
    user_id,
    organization_id,
    scopes,
    expires_at
) VALUES (
    accessToken,
    clientApp.ID,
    authCode.UserID,
    authCode.OrganizationID,
    authCode.Scopes,
    NOW() + INTERVAL '1 hour'
)

-- Store refresh token
INSERT INTO oauth_refresh_tokens (
    token_hash,
    client_id,
    user_id,
    organization_id,
    scopes,
    expires_at
) VALUES (
    hashedRefreshToken,
    clientApp.ID,
    authCode.UserID,
    authCode.OrganizationID,
    authCode.Scopes,
    NOW() + INTERVAL '30 days'
)
```

4.10. **Mark Authorization Code as Used**
```sql
UPDATE oauth_authorization_codes
SET used_at = NOW()
WHERE code = 'SplxlOBeZQQYbYS6WxSbIA'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "random-base64-string...",
  "scope": "user:read user:write"
}
```

---

#### Step 5: Client Uses Access Token

**Client (Dashboard Pro) API Request:**
```
GET https://auth.example.com/api/v1/oauth/userinfo
Headers:
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Backend Processing:**

5.1. **Validate Access Token (same as regular JWT)**
5.2. **Check Scopes**
```go
tokenScopes := strings.Split(claims["scope"].(string), " ")
if !contains(tokenScopes, "user:read") {
    return 403, "Insufficient scope"
}
```

5.3. **Return User Info**
```json
{
  "sub": "user-uuid",
  "email": "john@example.com",
  "name": "John Doe",
  "organization": "Acme Corp"
}
```

---

**Security Summary:**
- ✅ PKCE prevents code interception attacks
- ✅ Authorization code expires in 10 minutes
- ✅ Code can only be used once (replay prevention)
- ✅ Redirect URI validated (prevents token theft)
- ✅ Client credentials verified (confidential clients)
- ✅ Scopes enforced (principle of least privilege)

---

## 3. RBAC Permission Check Flow

**Scenario:** User tries to delete another user.

```
DELETE /api/v1/organizations/org-uuid-1/members/target-user-uuid
```

**Required Permission:** `user:delete`

### Step-by-Step Processing

#### Step 1: Extract JWT Claims (Middleware)
```go
func AuthMiddleware(c *gin.Context) {
    // Extract token
    token := extractToken(c.Request)
    
    // Validate and parse
    claims, err := validateJWT(token)
    if err != nil {
        c.JSON(401, gin.H{"error": "Invalid token"})
        c.Abort()
        return
    }
    
    // Store in context
    c.Set("user_id", claims["sub"])
    c.Set("organization_id", claims["organization_id"])
    c.Set("permissions", claims["permissions"])
    
    c.Next()
}
```

#### Step 2: Permission Check (Handler)
```go
func DeleteMember(c *gin.Context) {
    // Get from context
    permissions := c.MustGet("permissions").([]interface{})
    
    // Check permission
    if !hasPermission(permissions, "user:delete") {
        // Log denial
        auditLog.Record(AuditLog{
            UserID:         c.GetString("user_id"),
            OrganizationID: c.GetString("organization_id"),
            Action:         "permission_denied",
            ResourceType:   "user",
            ResourceID:     c.Param("userId"),
            Metadata: map[string]interface{}{
                "required_permission": "user:delete",
                "user_permissions":    permissions,
                "endpoint":            c.Request.URL.Path,
            },
        })
        
        c.JSON(403, gin.H{
            "error": "Forbidden: You don't have permission to delete users",
        })
        return
    }
    
    // Permission granted - proceed
    targetUserID := c.Param("userId")
    err := userService.DeleteMember(targetUserID)
    // ...
}
```

#### Step 3: Additional Checks (Business Logic)

**Prevent Self-Deletion:**
```go
if targetUserID == currentUserID {
    return error("Cannot delete yourself")
}
```

**Prevent Deleting Owner:**
```sql
SELECT role_id
FROM user_organizations
WHERE user_id = targetUserID
  AND organization_id = currentOrgID

-- Check if role is "Owner" (system role)
SELECT is_system, name
FROM roles
WHERE id = membership.role_id

if role.IsSystem && role.Name == "Owner" {
    return error("Cannot delete organization owner")
}
```

**Verify Organization Context:**
```sql
SELECT 1
FROM user_organizations
WHERE user_id = targetUserID
  AND organization_id = currentOrgID  -- ✅ Ensure same org
```

---

## 4. Multi-Tenant Session Management

### Session Creation (After Org Selection)

**Session Record:**
```sql
CREATE TABLE sessions (
    id                UUID PRIMARY KEY,
    user_id           UUID NOT NULL,
    organization_id   UUID NOT NULL,  -- ✅ Scoped to org
    refresh_token     TEXT,           -- Hashed
    ip_address        VARCHAR(45),
    user_agent        TEXT,
    device_name       VARCHAR(255),
    last_activity_at  TIMESTAMP,
    expires_at        TIMESTAMP,
    created_at        TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (organization_id) REFERENCES organizations(id)
)
```

**Key Points:**
1. Sessions are **scoped to organization**
2. User has **separate session per organization**
3. Switching orgs creates **new session**

### Session Listing (User's Active Sessions)

```sql
SELECT 
    s.id,
    s.organization_id,
    o.name AS organization_name,
    s.device_name,
    s.ip_address,
    s.last_activity_at,
    s.created_at
FROM sessions s
JOIN organizations o ON s.organization_id = o.id
WHERE s.user_id = current_user.ID
  AND s.expires_at > NOW()
ORDER BY s.last_activity_at DESC
```

**Example Response:**
```json
{
  "sessions": [
    {
      "id": "session-uuid-1",
      "organization": "Acme Corp",
      "device": "Chrome on macOS",
      "ip": "192.168.1.100",
      "last_activity": "2024-11-18T10:30:00Z",
      "created": "2024-11-18T09:00:00Z"
    },
    {
      "id": "session-uuid-2",
      "organization": "Startup Inc",
      "device": "Safari on iPhone",
      "ip": "192.168.1.105",
      "last_activity": "2024-11-18T08:15:00Z",
      "created": "2024-11-17T14:00:00Z"
    }
  ]
}
```

### Remote Logout (Revoke Specific Session)

```
DELETE /api/v1/user/sessions/session-uuid-1
```

**Backend:**
```sql
BEGIN TRANSACTION;

-- Delete session
DELETE FROM sessions
WHERE id = 'session-uuid-1'
  AND user_id = current_user.ID;  -- ✅ Verify ownership

-- Revoke associated OAuth tokens (if any)
DELETE FROM oauth_refresh_tokens
WHERE session_id = 'session-uuid-1';

COMMIT;
```

---

## 5. Password Reset Security Flow

### Step 1: Request Reset

```
POST /api/v1/auth/forgot-password
{ "email": "john@example.com" }
```

**Backend:**
```go
// Always return success (timing attack prevention)
defer func() {
    c.JSON(200, gin.H{
        "message": "If the email exists, a password reset link has been sent.",
    })
}()

// Find user
user := findUserByEmail(email)
if user == nil {
    return  // Don't reveal user existence
}

// Rate limit per email
attempts := redis.Incr("reset:" + email)
if attempts > 3 {
    return  // Silently fail (don't tell attacker)
}
redis.Expire("reset:" + email, 1 * time.Hour)

// Generate token
token := crypto.RandBytes(32)

// Store
INSERT INTO password_reset_tokens (
    user_id,
    token,
    expires_at
) VALUES (
    user.ID,
    token,
    NOW() + INTERVAL '1 hour'
)

// Send email
emailService.Send(PasswordResetEmail{
    To:   user.Email,
    Link: "https://app.example.com/reset-password?token=" + token,
})
```

**Security:**
- ✅ Same response regardless of email existence
- ✅ Rate limited (3 per hour per email)
- ✅ Token expires in 1 hour
- ✅ One-time use

### Step 2: Reset Password

```
POST /api/v1/auth/reset-password
{
  "token": "random-token",
  "password": "NewSecurePassword123!"
}
```

**Backend:**
```sql
BEGIN TRANSACTION;

-- Validate token
SELECT user_id, expires_at, used_at
FROM password_reset_tokens
WHERE token = 'random-token'
FOR UPDATE;  -- Lock row

-- Checks
if token.ExpiresAt < NOW() {
    return error("Token expired")
}

if token.UsedAt != NULL {
    return error("Token already used")
}

-- Update password
UPDATE users
SET password_hash = argon2.Hash(newPassword)
WHERE id = token.user_id;

-- Mark token as used
UPDATE password_reset_tokens
SET used_at = NOW()
WHERE token = 'random-token';

-- Revoke all sessions (force re-login)
DELETE FROM sessions
WHERE user_id = token.user_id;

-- Revoke all refresh tokens
UPDATE sessions SET refresh_token = NULL
WHERE user_id = token.user_id;

COMMIT;
```

**Security:**
- ✅ Token can only be used once
- ✅ All sessions revoked (attacker loses access)
- ✅ Transaction ensures atomic update

---

**Last Updated**: November 18, 2025  
**For Questions**: Ask the development team
