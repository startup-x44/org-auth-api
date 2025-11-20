# Technology Stack - Complete Breakdown

## Backend Stack

### Core Framework
- **Go 1.23** - Primary programming language
- **Gin v1.9.1** - HTTP web framework
  - Fast routing with radix tree
  - Middleware support
  - JSON binding/validation
  - Error handling

### Database & ORM
- **PostgreSQL 15** - Primary database
  - UUID primary keys (`gen_random_uuid()`)
  - JSONB for flexible schemas
  - Foreign key constraints
  - Full-text search capabilities
  
- **GORM v1.25.2** - ORM library
  - Auto-migrations
  - Associations management
  - Transaction support
  - Hooks (BeforeCreate, BeforeSave)

### Caching & Session Store
- **Redis 7** - In-memory data store
  - Session management
  - Rate limiting counters
  - Token blacklist
  - Cache layer

### Authentication & Security
- **golang-jwt/jwt v4.5.2** - JWT implementation
  - HS256 signing
  - Custom claims
  - Token validation
  
- **golang.org/x/crypto** - Cryptographic operations
  - Argon2 password hashing
  - bcrypt fallback
  - Secure random generation
  
- **PKCE** - OAuth2 security extension
  - Code challenge/verifier
  - SHA256 hashing
  - Public client protection

### Email
- **SMTP Integration** - Email delivery
  - Verification emails
  - Password reset
  - Invitation emails
  - Configurable SMTP settings

### Validation & Utilities
- **validator/v10** - Struct validation
- **google/uuid** - UUID generation
- **godotenv** - Environment variables

---

## Frontend Stack

### Core Framework
- **React 18.2** - UI library
  - Functional components
  - Hooks (useState, useEffect, useContext)
  - Concurrent features
  
- **TypeScript 5.1.6** - Type safety
  - Interface definitions
  - Type checking
  - Better IDE support

### Build Tools
- **Vite 7.2.2** - Build tool & dev server
  - Fast HMR (Hot Module Replacement)
  - Optimized bundling
  - Tree shaking
  - Code splitting

### Routing & State
- **React Router DOM v6.11.1** - Client-side routing
  - Protected routes
  - Nested routes
  - URL parameters
  
- **Zustand v5.0.8** - State management
  - Lightweight (< 1KB)
  - Persist middleware
  - No boilerplate
  - TypeScript support

### HTTP Client
- **Axios v1.4.0** - HTTP requests
  - Request/response interceptors
  - Token refresh handling
  - Error handling
  - Request cancellation

### UI Library
- **Radix UI** - Headless component library
  - `@radix-ui/react-dialog` - Modals
  - `@radix-ui/react-dropdown-menu` - Dropdowns
  - `@radix-ui/react-tabs` - Tabs
  - `@radix-ui/react-toast` - Notifications
  - `@radix-ui/react-alert-dialog` - Confirmations
  - Accessibility built-in (ARIA)
  - Keyboard navigation

### Styling
- **Tailwind CSS v3.3.2** - Utility-first CSS
  - Custom design system
  - Responsive utilities
  - Dark mode support
  
- **Framer Motion v12.23.24** - Animation library
  - Page transitions
  - Component animations
  - Gesture support

### Icons & Assets
- **Lucide React v0.553.0** - Icon library
  - 1000+ icons
  - Tree-shakeable
  - Customizable

---

## DevOps & Infrastructure

### Containerization
- **Docker** - Container runtime
  - Multi-stage builds
  - Alpine Linux base
  - Optimized layers
  
- **Docker Compose** - Multi-container orchestration
  - Development environment
  - Testing environment
  - Production-like setup

### Orchestration
- **Kubernetes** - Container orchestration
  - Deployment manifests
  - Service definitions
  - ConfigMaps & Secrets
  - Horizontal Pod Autoscaling (HPA)
  - Network Policies
  - Pod Disruption Budgets

### Deployment Files
```
k8s/
├── namespace.yaml           # Isolated namespace
├── configmap.yaml          # Configuration
├── secret.yaml             # Sensitive data
├── deployment.yaml         # App deployment
├── service.yaml            # Service exposure
├── ingress.yaml            # External access
├── hpa.yaml                # Auto-scaling
├── networkpolicy.yaml      # Network rules
├── poddisruptionbudget.yaml # Availability
└── serviceaccount.yaml     # RBAC
```

### Development Tools
- **Air** - Live reload for Go
- **npm/yarn** - Package management
- **Git** - Version control

---

## Database Schema

### Core Tables

#### Users Table
```sql
users
├── id (UUID, PK)
├── email (UNIQUE, NOT NULL)
├── password_hash (NOT NULL)
├── firstname, lastname, phone, address
├── is_superadmin (BOOLEAN)
├── global_role (user, admin)
├── status (active, suspended, deactivated)
├── email_verified_at
├── created_at, updated_at
```

#### Organizations Table
```sql
organizations
├── id (UUID, PK)
├── name (NOT NULL)
├── slug (UNIQUE, NOT NULL)
├── description
├── settings (JSONB)
├── status (active, suspended, archived)
├── created_by (FK → users.id)
├── created_at, updated_at
```

#### Organization Memberships
```sql
organization_memberships
├── id (UUID, PK)
├── organization_id (FK → organizations.id)
├── user_id (FK → users.id)
├── role_id (FK → roles.id)
├── status (active, invited, pending, suspended)
├── invited_by, invited_at, joined_at
├── UNIQUE(organization_id, user_id)
```

#### Roles Table
```sql
roles
├── id (UUID, PK)
├── organization_id (NULL for system roles)
├── name (NOT NULL)
├── display_name
├── description
├── is_system (BOOLEAN - system vs custom)
├── created_by (FK → users.id)
├── created_at, updated_at
```

#### Permissions Table
```sql
permissions
├── id (UUID, PK)
├── name (UNIQUE with organization_id)
├── display_name
├── description
├── category
├── is_system (BOOLEAN)
├── organization_id (NULL for system permissions)
├── created_at, updated_at
```

#### Role Permissions (Junction Table)
```sql
role_permissions
├── role_id (FK → roles.id, PK)
├── permission_id (FK → permissions.id, PK)
├── created_at
```

#### Sessions Table
```sql
user_sessions
├── id (UUID, PK)
├── user_id (FK → users.id)
├── organization_id (FK → organizations.id)
├── token_hash (UNIQUE, NOT NULL)
├── ip_address (INET)
├── user_agent
├── device_fingerprint
├── location (geographic)
├── is_active (BOOLEAN)
├── last_activity
├── expires_at
├── revoked_at, revoked_reason
├── created_at, updated_at
```

#### Refresh Tokens
```sql
refresh_tokens
├── id (UUID, PK)
├── user_id (FK → users.id)
├── organization_id (FK → organizations.id)
├── session_id (FK → user_sessions.id)
├── token_hash (Argon2, NOT NULL)
├── expires_at
├── revoked_at, revoked_reason
├── created_at, updated_at
```

#### OAuth2 Tables
```sql
client_apps
├── id (UUID, PK)
├── client_id (UNIQUE)
├── client_secret_hash
├── name
├── redirect_uris (ARRAY)
├── is_confidential (BOOLEAN - requires secret vs PKCE)
├── created_by, created_at, updated_at

authorization_codes
├── id (UUID, PK)
├── code_hash
├── client_id (FK)
├── user_id (FK)
├── organization_id (FK)
├── redirect_uri
├── scope
├── code_challenge, code_challenge_method (PKCE)
├── expires_at, used_at
├── created_at

oauth_refresh_tokens
├── id (UUID, PK)
├── token_hash
├── client_id (FK)
├── user_id (FK)
├── organization_id (FK)
├── scope
├── expires_at, revoked_at
├── created_at
```

#### API Keys
```sql
api_keys
├── id (UUID, PK)
├── user_id (FK → users.id)
├── organization_id (FK → organizations.id)
├── name
├── key_hash
├── key_prefix (for display: "ak_abc...")
├── scopes (ARRAY)
├── last_used_at
├── expires_at, revoked_at
├── created_at
```

---

## External Integrations

### Current
- **SMTP Server** - Email delivery (Mailtrap for dev)
- **PostgreSQL** - Primary database
- **Redis** - Caching & sessions

### Planned/Optional
- **AWS S3** - File storage
- **Stripe** - Payment processing
- **Twilio** - SMS verification
- **Sentry** - Error tracking
- **DataDog** - Observability
- **Auth0** - Social login fallback

---

## Testing Stack

### Backend Testing
- **testify** - Assertion library
- **HTTP testing** - Built-in Go testing
- **Test fixtures** - Sample data setup

### Frontend Testing
- **@testing-library/react** - Component testing
- **@testing-library/jest-dom** - DOM matchers
- **@testing-library/user-event** - User interaction

### Integration Testing
- **Docker Compose** - Test environment isolation
- **Feature tests** - E2E workflows

---

## Build & Deployment Pipeline

### Build Process
```
1. Frontend Build (Vite)
   ├── TypeScript compilation
   ├── Tree shaking
   ├── Code splitting (vendor, ui, utils)
   └── Asset optimization

2. Backend Build (Go)
   ├── Dependency download
   ├── CGO disabled compilation
   ├── Static binary generation
   └── Docker image creation

3. Docker Image
   ├── Multi-stage build
   ├── Alpine Linux (< 20MB)
   ├── Non-root user
   └── Health check configuration
```

### Deployment Strategy
- **Development** - Docker Compose with hot reload
- **Staging** - Kubernetes cluster (single replica)
- **Production** - Kubernetes cluster (HA, 3+ replicas)

---

**Tech Stack Version**: 1.0.0  
**Last Updated**: November 18, 2025  
**Go Version**: 1.23  
**Node Version**: 22.x
