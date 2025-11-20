# Deployment Guide

## Table of Contents
1. [Development Environment](#development-environment)
2. [Docker Deployment](#docker-deployment)
3. [Kubernetes Deployment](#kubernetes-deployment)
4. [Environment Variables](#environment-variables)
5. [Database Migrations](#database-migrations)
6. [Health Checks](#health-checks)
7. [Monitoring & Observability](#monitoring--observability)
8. [Production Checklist](#production-checklist)

---

## Development Environment

### Prerequisites
```bash
- Go 1.23+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)
```

### Local Setup

#### 1. Clone Repository
```bash
git clone https://github.com/your-org/auth-service.git
cd auth-service
```

#### 2. Install Dependencies
```bash
# Backend
go mod download

# Frontend
cd frontend
npm install
cd ..
```

#### 3. Configure Environment
```bash
# Copy example env file
cp .env.example .env

# Edit .env with your configuration
vim .env
```

#### 4. Start Database Services
```bash
# Using Docker Compose
docker-compose -f docker-compose.dev.yml up -d postgres redis

# Or manually
# Start PostgreSQL on port 5432
# Start Redis on port 6379
```

#### 5. Run Migrations
```bash
go run cmd/migrate/main.go
```

#### 6. Seed Database (Optional)
```bash
# Run seeder script (if available)
go run internal/seeder/main.go
```

#### 7. Start Backend
```bash
# Development mode with hot reload
./dev.sh

# Or manually
go run cmd/server/main.go
```

#### 8. Start Frontend
```bash
cd frontend
npm run dev
```

#### 9. Access Application
```
Backend API:  http://localhost:8080
Frontend UI:  http://localhost:5173
```

---

## Docker Deployment

### Development Mode

#### Using Docker Compose
```bash
# Build and start all services
docker-compose -f docker-compose.dev.yml up --build

# Run in background
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f auth-service

# Stop services
docker-compose -f docker-compose.dev.yml down
```

#### Services
```yaml
services:
  auth-service:
    - Backend API server
    - Port: 8080
    - Hot reload enabled

  postgres:
    - PostgreSQL 15
    - Port: 5432
    - Volume: postgres_data

  redis:
    - Redis 7
    - Port: 6379
    - Volume: redis_data

  frontend:
    - React development server
    - Port: 5173
```

### Production Build

#### 1. Build Docker Image
```bash
# Multi-stage build
docker build -t auth-service:latest -f Dockerfile .

# With version tag
docker build -t auth-service:v1.0.0 -f Dockerfile .
```

#### 2. Run Container
```bash
docker run -d \
  --name auth-service \
  -p 8080:8080 \
  --env-file .env.production \
  auth-service:latest
```

#### 3. Production Docker Compose
```bash
docker-compose -f docker-compose.yml up -d
```

**Production Compose Stack:**
```yaml
services:
  auth-service:
    - Production build with multi-stage Dockerfile
    - Optimized image size (~50MB)
    - Health checks enabled
    - Restart policy: always

  postgres:
    - PostgreSQL 15 with persistent volume
    - Backup cron job (optional)
    - Connection pooling

  redis:
    - Redis 7 with persistence (AOF + RDB)
    - Memory limits configured

  nginx:
    - Reverse proxy
    - SSL/TLS termination
    - Static file serving (frontend)
    - Rate limiting
```

---

## Kubernetes Deployment

### Architecture
```
                 ┌──────────────┐
                 │   Ingress    │
                 │  (nginx-ic)  │
                 └──────┬───────┘
                        │
          ┌─────────────┴─────────────┐
          │                           │
    ┌─────▼──────┐            ┌──────▼─────┐
    │  Auth API  │            │  Frontend  │
    │ Deployment │            │ Deployment │
    │ (3 pods)   │            │ (2 pods)   │
    └─────┬──────┘            └────────────┘
          │
    ┌─────▼──────────────────┐
    │   ConfigMap & Secret   │
    └─────┬──────────────────┘
          │
    ┌─────┴──────────────────┐
    │                        │
┌───▼────┐              ┌───▼────┐
│ Postgres│             │ Redis  │
│StatefulSet│           │StatefulSet│
└────────┘              └────────┘
```

### Prerequisites
```bash
- Kubernetes cluster (1.25+)
- kubectl configured
- Helm (optional)
- Persistent volume provisioner
```

### 1. Create Namespace
```bash
kubectl apply -f k8s/namespace.yaml
```

**namespace.yaml:**
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: auth-service
  labels:
    name: auth-service
    environment: production
```

### 2. Configure Secrets
```bash
# Create secret from file
kubectl create secret generic auth-service-secrets \
  --from-env-file=.env.production \
  -n auth-service

# Or apply secret manifest
kubectl apply -f k8s/secret.yaml
```

**secret.yaml:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secrets
  namespace: auth-service
type: Opaque
data:
  DB_PASSWORD: <base64-encoded>
  JWT_SECRET: <base64-encoded>
  REDIS_PASSWORD: <base64-encoded>
```

### 3. Apply ConfigMap
```bash
kubectl apply -f k8s/configmap.yaml
```

**configmap.yaml:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-service-config
  namespace: auth-service
data:
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "auth_db"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  LOG_LEVEL: "info"
```

### 4. Deploy PostgreSQL
```bash
kubectl apply -f k8s/postgres-statefulset.yaml
kubectl apply -f k8s/postgres-service.yaml
```

**postgres-statefulset.yaml:**
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: auth-service
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_DB
          valueFrom:
            configMapKeyRef:
              name: auth-service-config
              key: DB_NAME
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: DB_PASSWORD
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
```

### 5. Deploy Redis
```bash
kubectl apply -f k8s/redis-statefulset.yaml
kubectl apply -f k8s/redis-service.yaml
```

### 6. Deploy Auth Service
```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

**deployment.yaml:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: auth-service
  labels:
    app: auth-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      serviceAccountName: auth-service-sa
      containers:
      - name: auth-service
        image: your-registry/auth-service:v1.0.0
        ports:
        - containerPort: 8080
          name: http
        envFrom:
        - configMapRef:
            name: auth-service-config
        - secretRef:
            name: auth-service-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 7. Configure Horizontal Pod Autoscaler
```bash
kubectl apply -f k8s/hpa.yaml
```

**hpa.yaml:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: auth-service-hpa
  namespace: auth-service
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: auth-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### 8. Deploy Ingress
```bash
kubectl apply -f k8s/ingress.yaml
```

**ingress.yaml:**
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: auth-service-ingress
  namespace: auth-service
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - auth.yourdomain.com
    secretName: auth-tls-secret
  rules:
  - host: auth.yourdomain.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: auth-service
            port:
              number: 8080
      - path: /
        pathType: Prefix
        backend:
          service:
            name: auth-service-frontend
            port:
              number: 80
```

### 9. Network Policies
```bash
kubectl apply -f k8s/networkpolicy.yaml
```

**networkpolicy.yaml:**
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: auth-service-netpol
  namespace: auth-service
spec:
  podSelector:
    matchLabels:
      app: auth-service
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: nginx-ingress
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

### 10. Verify Deployment
```bash
# Check all resources
kubectl get all -n auth-service

# Check pod logs
kubectl logs -f deployment/auth-service -n auth-service

# Check service endpoints
kubectl get endpoints -n auth-service

# Test service
kubectl port-forward svc/auth-service 8080:8080 -n auth-service
curl http://localhost:8080/health
```

---

## Environment Variables

### Required Variables

#### Database
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-secure-password
DB_NAME=auth_db
DB_SSLMODE=disable  # Use 'require' in production
```

#### Redis
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password
REDIS_DB=0
```

#### JWT
```bash
JWT_SECRET=your-256-bit-secret-key-here
JWT_ACCESS_TOKEN_DURATION=1h
JWT_REFRESH_TOKEN_DURATION=720h  # 30 days
```

#### Server
```bash
PORT=8080
GIN_MODE=release  # 'debug' for development
LOG_LEVEL=info
CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

#### Email (SMTP)
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourdomain.com
```

#### Frontend
```bash
VITE_API_URL=https://api.yourdomain.com
VITE_APP_NAME=Auth Service
```

### Optional Variables
```bash
# Rate Limiting
RATE_LIMIT_LOGIN=20
RATE_LIMIT_REGISTER=10

# Session
SESSION_DURATION=720h
SESSION_CLEANUP_INTERVAL=1h

# OAuth2
OAUTH2_AUTHORIZATION_CODE_DURATION=10m
OAUTH2_ACCESS_TOKEN_DURATION=1h

# Audit
AUDIT_LOG_RETENTION_DAYS=90
```

---

## Database Migrations

### Running Migrations

#### Development
```bash
# Run all migrations
go run cmd/migrate/main.go

# Or using make
make migrate-up
```

#### Production
```bash
# Run migrations before deploying new version
go run cmd/migrate/main.go --env=production

# Kubernetes Job
kubectl apply -f k8s/migration-job.yaml
```

**migration-job.yaml:**
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: auth-service-migration
  namespace: auth-service
spec:
  template:
    spec:
      containers:
      - name: migrator
        image: your-registry/auth-service:v1.0.0
        command: ["/app/migrate"]
        envFrom:
        - configMapRef:
            name: auth-service-config
        - secretRef:
            name: auth-service-secrets
      restartPolicy: OnFailure
  backoffLimit: 3
```

### Migration Files

**Location:** `/migrations`

**Naming Convention:**
```
001_add_session_fields.go
002_organization_system.up.sql
002_organization_system.down.sql
003_add_permission_is_system.up.sql
003_add_permission_is_system.down.sql
```

### Creating New Migration
```bash
# Create migration files
touch migrations/011_your_migration.up.sql
touch migrations/011_your_migration.down.sql

# Write SQL
vim migrations/011_your_migration.up.sql
```

**Example:**
```sql
-- migrations/011_add_user_metadata.up.sql
ALTER TABLE users ADD COLUMN metadata JSONB DEFAULT '{}'::jsonb;
CREATE INDEX idx_users_metadata ON users USING GIN (metadata);

-- migrations/011_add_user_metadata.down.sql
DROP INDEX IF EXISTS idx_users_metadata;
ALTER TABLE users DROP COLUMN IF EXISTS metadata;
```

---

## Health Checks

### Endpoint
```
GET /health
```

### Response
```json
{
  "status": "ok",
  "timestamp": "2024-11-18T10:30:00Z",
  "services": {
    "database": {
      "status": "up",
      "latency_ms": 5
    },
    "redis": {
      "status": "up",
      "latency_ms": 2
    }
  },
  "version": "1.0.0"
}
```

### Kubernetes Probes

#### Liveness Probe
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

#### Readiness Probe
```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

---

## Monitoring & Observability

### Metrics (Prometheus)

#### Exposed Metrics
```
# HTTP metrics
http_requests_total{method, path, status}
http_request_duration_seconds{method, path}

# Database metrics
db_connections_active
db_connections_idle
db_query_duration_seconds{query}

# Redis metrics
redis_commands_total{command}
redis_connection_errors_total

# Application metrics
auth_login_attempts_total{status}
auth_token_issued_total{type}
rbac_permission_checks_total{result}
```

#### Prometheus Configuration
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'auth-service'
    kubernetes_sd_configs:
    - role: pod
      namespaces:
        names:
        - auth-service
    relabel_configs:
    - source_labels: [__meta_kubernetes_pod_label_app]
      action: keep
      regex: auth-service
    - source_labels: [__meta_kubernetes_pod_name]
      target_label: pod
```

### Logging

#### Log Format
```json
{
  "level": "info",
  "timestamp": "2024-11-18T10:30:00Z",
  "request_id": "uuid",
  "user_id": "uuid",
  "organization_id": "uuid",
  "method": "POST",
  "path": "/api/v1/auth/login",
  "status": 200,
  "latency_ms": 45,
  "ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "message": "User logged in successfully"
}
```

#### Log Aggregation (ELK Stack)
```yaml
# filebeat.yml
filebeat.inputs:
- type: container
  paths:
    - /var/log/containers/auth-service-*.log
  processors:
  - add_kubernetes_metadata:
      in_cluster: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "auth-service-%{+yyyy.MM.dd}"
```

### Distributed Tracing (Jaeger)

#### Instrumentation
```go
import "github.com/opentelemetry/opentelemetry-go"

// Trace HTTP requests
func TracingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, span := tracer.Start(c.Request.Context(), c.Request.URL.Path)
        defer span.End()
        
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```

---

## Production Checklist

### Pre-Deployment

- [ ] **Environment Variables**
  - [ ] All secrets configured
  - [ ] Production database credentials
  - [ ] SMTP configured and tested
  - [ ] JWT secret is strong (256-bit)
  - [ ] CORS origins whitelisted

- [ ] **Database**
  - [ ] Migrations tested
  - [ ] Backup strategy in place
  - [ ] Connection pooling configured
  - [ ] SSL/TLS enabled

- [ ] **Security**
  - [ ] HTTPS enforced
  - [ ] Security headers configured
  - [ ] Rate limiting enabled
  - [ ] CSRF protection enabled
  - [ ] Network policies applied

- [ ] **Performance**
  - [ ] Redis caching enabled
  - [ ] Database indexes created
  - [ ] Connection pooling configured
  - [ ] Resource limits set

- [ ] **Monitoring**
  - [ ] Prometheus metrics configured
  - [ ] Logging aggregation setup
  - [ ] Alerts configured
  - [ ] Health checks enabled

- [ ] **High Availability**
  - [ ] Multiple replicas (min 3)
  - [ ] Pod disruption budget configured
  - [ ] HPA configured
  - [ ] Database replication (if applicable)

### Post-Deployment

- [ ] **Verification**
  - [ ] All pods running
  - [ ] Health checks passing
  - [ ] Database connections working
  - [ ] Redis connections working
  - [ ] Ingress routing correctly

- [ ] **Testing**
  - [ ] Smoke tests passed
  - [ ] Authentication flows working
  - [ ] OAuth2 flows working
  - [ ] RBAC permissions enforced

- [ ] **Monitoring**
  - [ ] Metrics being collected
  - [ ] Logs being aggregated
  - [ ] Alerts firing correctly
  - [ ] Dashboard configured

---

**Last Updated**: November 18, 2025  
**Deployment Team**: devops@example.com
