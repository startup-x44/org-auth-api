# ✅ Phase 3 Testing Checklist

Complete this checklist to verify all OAuth2.1 features are working correctly.

## 📋 Pre-Testing Setup

### Start Development Environment

```bash
# Terminal 1: Start backend + database
./dev.sh dev

# Terminal 2: Start callback server (for testing)
node callback-server.js
```

**Verify**:
- [ ] Backend running at http://localhost:8080
- [ ] Frontend running at http://localhost:5173
- [ ] Callback server running at http://localhost:3000
- [ ] Database is accessible
- [ ] You have superadmin credentials

---

## 🔐 Test 1: OAuth Client Apps Management

### 1.1 Access Page
- [ ] Login as superadmin at http://localhost:5173/login
- [ ] Navigate to `/admin`
- [ ] Click "OAuth Client Apps" quick action card
- [ ] Page loads at `/oauth/client-apps` without errors
- [ ] "Create Client App" button visible

### 1.2 Create Client App
- [ ] Click "Create Client App" button
- [ ] Dialog opens
- [ ] Fill in form:
  - **Name**: `Test Application`
  - **Description**: `OAuth2.1 integration testing`
  - **Redirect URIs** (one per line):
    ```
    http://localhost:3000/callback
    http://localhost:5173/oauth/callback
    ```
  - **Allowed Scopes**: `profile email org:read`
- [ ] Click "Create Application"
- [ ] Success toast appears
- [ ] Secret modal displays with one-time warning
- [ ] Client ID visible and copyable
- [ ] Client secret visible and copyable
- [ ] **SAVE CREDENTIALS NOW**:
  ```
  Client ID: _______________________
  Client Secret: ____________________
  ```
- [ ] Click "I've Saved It" - modal closes
- [ ] New app appears in list

### 1.3 View Details
- [ ] App card shows correct name and description
- [ ] Client ID displayed (partially hidden)
- [ ] Click eye icon - full Client ID revealed
- [ ] Click copy icon - ID copied to clipboard
- [ ] Redirect URIs listed correctly
- [ ] Scopes shown as badges
- [ ] Created/Updated timestamps visible

### 1.4 Edit Client App
- [ ] Click Edit (pencil) icon
- [ ] Edit dialog opens with current values
- [ ] Change name to: `Updated Test App`
- [ ] Add redirect URI: `https://example.com/callback`
- [ ] Update scopes to: `profile email org:read org:write`
- [ ] Click "Update Application"
- [ ] Success toast appears
- [ ] Changes reflected in app card
- [ ] Updated timestamp changed

### 1.5 Rotate Secret
- [ ] Click Rotate (refresh/arrows) icon
- [ ] Confirmation dialog appears
- [ ] Warning about old secret invalidation
- [ ] Click "Rotate Secret"
- [ ] New secret modal appears
- [ ] New secret displayed
- [ ] Copy button works
- [ ] **SAVE NEW SECRET**: ____________________
- [ ] Click "I've Saved It"

### 1.6 Delete Client App (Optional - Skip if you need app for testing)
- [ ] Click Delete (trash) icon
- [ ] Confirmation dialog appears
- [ ] App name shown in warning
- [ ] Click "Delete" to confirm
- [ ] Success toast appears
- [ ] App removed from list

---

## 🎫 Test 2: OAuth Consent Screen

### 2.1 Generate PKCE Parameters

Open browser console and run:

```javascript
function generateRandomString(length) {
  const charset = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
  const randomValues = new Uint8Array(length);
  crypto.getRandomValues(randomValues);
  return Array.from(randomValues).map(v => charset[v % charset.length]).join('');
}

async function sha256(plain) {
  const encoder = new TextEncoder();
  const data = encoder.encode(plain);
  const hash = await crypto.subtle.digest('SHA-256', data);
  const bytes = new Uint8Array(hash);
  const binary = String.fromCharCode(...bytes);
  const base64 = btoa(binary);
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

const codeVerifier = generateRandomString(128);
const codeChallenge = await sha256(codeVerifier);

console.log('Code Verifier:', codeVerifier);
console.log('Code Challenge:', codeChallenge);
sessionStorage.setItem('code_verifier', codeVerifier);
```

**Save Output**:
```
Code Verifier: _________________________
Code Challenge: ________________________
```

### 2.2 Navigate to Consent Screen

**Build Authorization URL** (replace placeholders):
```
http://localhost:5173/oauth/authorize?client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost:3000/callback&response_type=code&scope=profile%20email&code_challenge=YOUR_CODE_CHALLENGE&code_challenge_method=S256&state=test_state_123
```

**Test**:
- [ ] Login as a **regular user** (NOT superadmin)
- [ ] Navigate to the authorization URL above
- [ ] Consent screen loads without errors

### 2.3 Verify Consent Screen UI
- [ ] Shows "Signed in as [user email]"
- [ ] Application name displayed: "Test Application"
- [ ] Application description shown
- [ ] Requested scopes listed:
  - [ ] Profile scope with icon and description
  - [ ] Email scope with icon and description
- [ ] "Allow Access" button visible
- [ ] "Deny" button visible
- [ ] "Stay signed in" checkbox visible

### 2.4 Test Allow Flow
- [ ] Click "Allow Access"
- [ ] Backend processes request
- [ ] Redirects to `http://localhost:3000/callback?code=...&state=test_state_123`
- [ ] Callback server displays success page
- [ ] Authorization code visible
- [ ] State parameter matches: `test_state_123`
- [ ] **SAVE AUTH CODE**: ____________________

### 2.5 Test Deny Flow
- [ ] Go back and navigate to authorization URL again
- [ ] Click "Deny"
- [ ] Redirects to `http://localhost:3000/callback?error=access_denied&error_description=...&state=test_state_123`
- [ ] Callback server displays error page
- [ ] Error message shown

### 2.6 Test Stay Signed In
- [ ] Navigate to authorization URL again
- [ ] Check "Stay signed in for 30 days"
- [ ] Click "Allow Access"
- [ ] Checkbox state maintained
- [ ] Redirect successful

---

## 🔄 Test 3: OAuth Callback Handler (Frontend)

### 3.1 Success Callback
Navigate to:
```
http://localhost:5173/oauth/callback?code=test_code_123&state=test_state
```

- [ ] Callback page loads
- [ ] Green checkmark icon displayed
- [ ] "Authorization Successful" title
- [ ] Success message shown
- [ ] Authorization code visible (dev mode)
- [ ] State parameter visible (dev mode)
- [ ] "Return to Dashboard" button works

### 3.2 Error Callback
Navigate to:
```
http://localhost:5173/oauth/callback?error=access_denied&error_description=User%20denied%20access
```

- [ ] Callback page loads
- [ ] Red X icon displayed
- [ ] "Authorization Failed" title
- [ ] Error description shown
- [ ] "Return to Dashboard" button works

### 3.3 Invalid Callback
Navigate to:
```
http://localhost:5173/oauth/callback
```
(no parameters)

- [ ] Error state displayed
- [ ] "Invalid callback" message
- [ ] Red X icon shown

---

## 📚 Test 4: Developer Documentation

### 4.1 Access Documentation
- [ ] Navigate to `/developer/docs`
- [ ] Page loads without errors
- [ ] 4 tabs visible: Quick Start, JavaScript, Node.js, React

### 4.2 Test Tabs
- [ ] Click "Quick Start" tab - content loads
- [ ] Click "JavaScript" tab - browser code examples shown
- [ ] Click "Node.js" tab - server code examples shown
- [ ] Click "React" tab - React component examples shown
- [ ] No console errors during tab switching

### 4.3 Test Code Copy
- [ ] Hover over any code block
- [ ] Copy button appears
- [ ] Click copy button
- [ ] Toast notification: "Copied!"
- [ ] Check mark icon appears briefly
- [ ] Paste clipboard - code is there

### 4.4 Verify Code Examples
- [ ] Auth server URL is correct: `http://localhost:8080`
- [ ] PKCE implementation shown
- [ ] All OAuth parameters documented
- [ ] Token exchange example provided
- [ ] API usage with Bearer token shown

---

## 🔐 Test 5: Token Exchange (Backend API)

### 5.1 Exchange Authorization Code

Use the auth code from Test 2.4:

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=YOUR_AUTH_CODE" \
  -d "redirect_uri=http://localhost:3000/callback" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "code_verifier=YOUR_CODE_VERIFIER"
```

**Verify Response**:
- [ ] Status: 200 OK
- [ ] Response contains:
  - [ ] `access_token` (JWT)
  - [ ] `token_type`: "Bearer"
  - [ ] `expires_in`: 3600
  - [ ] `refresh_token`
  - [ ] `id_token` (optional)
- [ ] **SAVE TOKENS**:
  ```
  Access Token: ___________________
  Refresh Token: __________________
  ```

### 5.2 Test UserInfo Endpoint

```bash
curl http://localhost:8080/oauth/userinfo \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Verify Response**:
- [ ] Status: 200 OK
- [ ] Contains `sub` (user ID)
- [ ] Contains `email`
- [ ] Contains user profile data
- [ ] No errors

### 5.3 Verify JWT Claims

Decode the access token at https://jwt.io or use:

```bash
echo "YOUR_ACCESS_TOKEN" | cut -d'.' -f2 | base64 -d | jq '.'
```

**Verify Claims**:
- [ ] `user_id` present
- [ ] `organization_id` present
- [ ] `role_id` present
- [ ] `permissions` array present
- [ ] `exp` (expiration) present
- [ ] `iat` (issued at) present

### 5.4 Test Refresh Token

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=YOUR_REFRESH_TOKEN" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

**Verify**:
- [ ] Status: 200 OK
- [ ] New access_token received
- [ ] New refresh_token received (rotated)
- [ ] Old refresh token now invalid

---

## 📊 Test 6: OAuth Audit Logs

### 6.1 Access Audit Logs
- [ ] Login as superadmin
- [ ] Navigate to `/admin`
- [ ] Click "OAuth Audit Logs" quick action card
- [ ] Page loads at `/oauth/audit`

### 6.2 Authorizations Tab
- [ ] Tab selected by default
- [ ] Table shows recent authorizations
- [ ] Columns visible:
  - [ ] Client App name
  - [ ] User email
  - [ ] Scopes
  - [ ] Status (used/expired/active)
  - [ ] Created timestamp
  - [ ] Used At timestamp
- [ ] Filter by Client App works
- [ ] Filter by User ID works
- [ ] Status badges color-coded correctly

### 6.3 Active Tokens Tab
- [ ] Click "Active Tokens" tab
- [ ] Table shows refresh tokens
- [ ] Columns visible:
  - [ ] Client App name
  - [ ] User email
  - [ ] Scopes
  - [ ] Status (active/inactive)
  - [ ] Created timestamp
  - [ ] Last Used timestamp
  - [ ] Expires timestamp
- [ ] Filter by Client App works
- [ ] Filter by User ID works
- [ ] Active status shown correctly

### 6.4 Statistics Tab
- [ ] Click "Statistics" tab
- [ ] 3 stat cards displayed:
  - [ ] Total Authorizations (with "today" count)
  - [ ] Active Tokens (with "of total" count)
  - [ ] Unique Users (with "across X clients")
- [ ] Numbers are accurate
- [ ] Icons displayed correctly

---

## 🔒 Test 7: Security & Access Control

### 7.1 Superadmin-Only Routes
- [ ] Logout
- [ ] Login as **regular user** (not superadmin)
- [ ] Try to access `/oauth/client-apps`
- [ ] **Expected**: Redirected to `/dashboard`
- [ ] Try to access `/oauth/audit`
- [ ] **Expected**: Redirected to `/dashboard`

### 7.2 Protected Routes
- [ ] Logout completely
- [ ] Try to access `/oauth/authorize` while logged out
- [ ] **Expected**: Redirected to `/login`
- [ ] Try to access `/developer/docs` while logged out
- [ ] **Expected**: Redirected to `/login`

### 7.3 Public Callback Route
- [ ] While logged out, navigate to:
  ```
  http://localhost:5173/oauth/callback?code=test
  ```
- [ ] **Expected**: Page loads (not redirected)
- [ ] Callback handler works without authentication

---

## 📱 Test 8: Responsive Design

### 8.1 Mobile View
- [ ] Open Chrome DevTools
- [ ] Toggle device toolbar (Cmd+Shift+M)
- [ ] Select iPhone 12 or similar
- [ ] Navigate to `/oauth/client-apps`
- [ ] Cards stack vertically
- [ ] Buttons accessible
- [ ] Table scrolls horizontally
- [ ] Navigate to `/oauth/consent`
- [ ] Form is mobile-friendly
- [ ] Navigate to `/developer/docs`
- [ ] Code blocks scrollable

### 8.2 Tablet View
- [ ] Select iPad or tablet size
- [ ] Test all OAuth pages
- [ ] Layout adapts correctly
- [ ] Grid columns adjust (2 columns)

---

## 🎯 Test 9: Complete End-to-End Flow

Run the automated test script:

```bash
./test-oauth-flow.sh
```

**Follow prompts**:
- [ ] Script checks backend is running
- [ ] Enter Client ID
- [ ] Enter Client Secret
- [ ] PKCE parameters generated automatically
- [ ] Authorize in browser when prompted
- [ ] Paste authorization code when prompted
- [ ] Script exchanges code for tokens
- [ ] Script tests UserInfo endpoint
- [ ] Script decodes and verifies JWT claims
- [ ] **All steps pass** ✅

---

## ⚡ Test 10: Error Handling

### 10.1 Network Errors
- [ ] Stop backend server
- [ ] Try to create client app in UI
- [ ] **Expected**: Error toast appears
- [ ] **Expected**: Graceful error message (not crash)
- [ ] **Expected**: Loading spinner stops

### 10.2 Validation Errors
- [ ] Try to create client app with:
  - [ ] Empty name → Form validation prevents submit
  - [ ] No redirect URIs → Button disabled
  - [ ] Invalid redirect URI format → Error shown

### 10.3 Invalid Authorization Requests
Try authorization URL with missing parameters:
```
http://localhost:5173/oauth/authorize?client_id=invalid
```

- [ ] Error displayed on consent screen
- [ ] OR backend returns error redirect
- [ ] No uncaught exceptions

---

## ✅ Final Verification

### All Features Working
- [ ] ✅ OAuth Client Apps CRUD
- [ ] ✅ Client Secret rotation
- [ ] ✅ OAuth Consent Screen
- [ ] ✅ Authorization flow (PKCE)
- [ ] ✅ Token exchange
- [ ] ✅ Refresh token rotation
- [ ] ✅ UserInfo endpoint
- [ ] ✅ JWT with RBAC claims
- [ ] ✅ OAuth Audit Logs
- [ ] ✅ Developer Documentation
- [ ] ✅ Security (superadmin access)
- [ ] ✅ Responsive design

### No Errors
- [ ] No console errors in browser
- [ ] No backend errors in logs
- [ ] No TypeScript compilation errors
- [ ] No broken links or 404s

### Documentation
- [ ] Code examples work
- [ ] All URIs correct
- [ ] PKCE implementation accurate
- [ ] Testing guide followed successfully

---

## 🎉 Success!

If all checkboxes are marked, **Phase 3 is COMPLETE** ✅

### Next Steps Options:
1. **Task 4**: Build System RBAC Management UI (optional)
2. **Phase 4**: Create JavaScript SDK NPM package
3. **Phase 5**: Billing integration
4. **Production Deployment**: Deploy to staging/production

---

## 📝 Test Results

**Date**: _______________  
**Tester**: _______________  
**Result**: [ ] PASS [ ] FAIL  

**Issues Found**:
```
1. _____________________________
2. _____________________________
3. _____________________________
```

**Notes**:
```
_________________________________
_________________________________
_________________________________
```
