# Deployment Guide

This guide covers deploying the SaaS Authentication Microservice to various environments.

## Prerequisites

- Kubernetes cluster (v1.19+)
- PostgreSQL database (v15+)
- Redis instance (v7+)
- Email service (SendGrid, SES, etc.)
- kubectl configured for your cluster

## Environment Variables

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `JWT_SECRET` | JWT signing secret (32+ chars) | `your-super-secret-jwt-key-change-in-production` |
| `DATABASE_HOST` | PostgreSQL host | `postgres-service.default.svc.cluster.local` |
| `DATABASE_PORT` | PostgreSQL port | `5432` |
| `DATABASE_USER` | PostgreSQL username | `auth_user` |
| `DATABASE_PASSWORD` | PostgreSQL password | `secure-password` |
| `DATABASE_NAME` | PostgreSQL database | `auth_db` |
| `REDIS_HOST` | Redis host | `redis-service.default.svc.cluster.local` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | `redis-password` |
| `EMAIL_API_KEY` | Email service API key | `SG.xxxxxxxx` |

### Optional Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Server port |
| `ENVIRONMENT` | `production` | Environment name |
| `JWT_ACCESS_TOKEN_EXPIRY` | `3600` | Access token expiry (seconds) |
| `JWT_REFRESH_TOKEN_EXPIRY` | `604800` | Refresh token expiry (seconds) |
| `RATE_LIMIT_LOGIN_ATTEMPTS` | `5` | Login attempts per 15min per IP |
| `RATE_LIMIT_API_CALLS` | `1000` | API calls per minute per user |
| `MAX_CONCURRENT_SESSIONS` | `5` | Max sessions per user |

## Kubernetes Deployment

### 1. Create Namespace

```bash
kubectl create namespace auth-service
```

### 2. Create Secrets

```bash
# Database and Redis secrets
kubectl create secret generic auth-service-secret \
  --namespace auth-service \
  --from-literal=JWT_SECRET='your-super-secret-jwt-key-change-in-production' \
  --from-literal=DATABASE_HOST='postgres-service.default.svc.cluster.local' \
  --from-literal=DATABASE_PORT='5432' \
  --from-literal=DATABASE_USER='auth_user' \
  --from-literal=DATABASE_PASSWORD='secure-db-password' \
  --from-literal=DATABASE_NAME='auth_db' \
  --from-literal=REDIS_HOST='redis-service.default.svc.cluster.local' \
  --from-literal=REDIS_PORT='6379' \
  --from-literal=REDIS_PASSWORD='redis-password' \
  --from-literal=EMAIL_API_KEY='SG.your-sendgrid-api-key' \
  --from-literal=EMAIL_FROM_EMAIL='noreply@yourdomain.com' \
  --from-literal=APP_SECRET_KEY='another-secret-key'
```

### 3. Create ConfigMap

```bash
kubectl create configmap auth-service-config \
  --namespace auth-service \
  --from-literal=SERVER_PORT='8080' \
  --from-literal=ENVIRONMENT='production' \
  --from-literal=JWT_ACCESS_TOKEN_EXPIRY='3600' \
  --from-literal=JWT_REFRESH_TOKEN_EXPIRY='604800' \
  --from-literal=RATE_LIMIT_LOGIN_ATTEMPTS='5' \
  --from-literal=RATE_LIMIT_API_CALLS='1000' \
  --from-literal=MAX_CONCURRENT_SESSIONS='5' \
  --from-literal=EMAIL_PROVIDER='sendgrid' \
  --from-literal=EMAIL_FROM_NAME='Auth Service' \
  --from-literal=PASSWORD_RESET_URL='https://app.yourdomain.com/reset-password' \
  --from-literal=LOG_LEVEL='info' \
  --from-literal=DATABASE_SSLMODE='require' \
  --from-literal=DATABASE_MAX_CONNECTIONS='10' \
  --from-literal=DATABASE_MAX_IDLE_CONNECTIONS='5' \
  --from-literal=DATABASE_CONNECTION_MAX_LIFETIME='300' \
  --from-literal=REDIS_DB='0' \
  --from-literal=REDIS_POOL_SIZE='10' \
  --from-literal=REDIS_MIN_IDLE_CONNS='5' \
  --from-literal=REDIS_CONN_MAX_LIFETIME='300'
```

### 4. Apply Kubernetes Manifests

```bash
# Apply in order
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/ingress.yaml
kubectl apply -f k8s/networkpolicy.yaml
kubectl apply -f k8s/poddisruptionbudget.yaml
kubectl apply -f k8s/serviceaccount.yaml
```

### 5. Verify Deployment

```bash
# Check pod status
kubectl get pods -n auth-service

# Check service
kubectl get svc -n auth-service

# Check ingress
kubectl get ingress -n auth-service

# View logs
kubectl logs -f deployment/auth-service -n auth-service
```

## Database Setup

### PostgreSQL Setup

1. **Create Database and User:**

```sql
-- Connect as superuser
CREATE DATABASE auth_db;
CREATE USER auth_user WITH ENCRYPTED PASSWORD 'secure-password';
GRANT ALL PRIVILEGES ON DATABASE auth_db TO auth_user;

-- Connect to auth_db
\c auth_db

-- Grant schema permissions
GRANT ALL ON SCHEMA public TO auth_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO auth_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO auth_user;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO auth_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO auth_user;
```

2. **Run Migrations:**

The application will automatically run migrations on startup. You can also run them manually:

```bash
# Via kubectl
kubectl exec -it deployment/auth-service -n auth-service -- ./migrate up

# Or check the migration status
kubectl exec -it deployment/auth-service -n auth-service -- ./migrate status
```

## Redis Setup

### Redis Configuration

Ensure Redis is configured with:

- **Persistence**: AOF or RDB enabled
- **Security**: Password authentication
- **Memory**: Sufficient memory for session storage
- **Backup**: Regular backups configured

### Redis Cluster (Optional)

For high availability:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-cluster
  namespace: auth-service
spec:
  serviceName: redis-cluster
  replicas: 6
  template:
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command: ["redis-server", "/etc/redis/redis.conf"]
        volumeMounts:
        - name: redis-config
          mountPath: /etc/redis
        - name: redis-data
          mountPath: /data
      volumes:
      - name: redis-config
        configMap:
          name: redis-config
      - name: redis-data
        persistentVolumeClaim:
          claimName: redis-data
```

## Monitoring Setup

### Health Checks

The service provides health endpoints:

- **Liveness**: `GET /health/live`
- **Readiness**: `GET /health/ready`
- **General Health**: `GET /health`

### Metrics

For Prometheus metrics, ensure the `/metrics` endpoint is exposed.

### Logging

Logs are structured in JSON format. Configure your log aggregator to parse:

```json
{
  "timestamp": "2023-01-01T12:00:00Z",
  "level": "info",
  "service": "auth-service",
  "request_id": "req-123",
  "message": "User login successful",
  "user_id": "user-456",
  "ip": "192.168.1.1"
}
```

## Scaling

### Horizontal Pod Autoscaling

The HPA is configured based on CPU and memory:

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

### Database Scaling

For high traffic:

1. **Connection Pooling**: Adjust `DATABASE_MAX_CONNECTIONS`
2. **Read Replicas**: Configure read replicas for read-heavy workloads
3. **Sharding**: Implement tenant-based sharding for multi-tenant scaling

## Backup and Recovery

### Database Backup

```bash
# Create backup
pg_dump -h postgres-host -U auth_user -d auth_db > auth_backup.sql

# Restore backup
psql -h postgres-host -U auth_user -d auth_db < auth_backup.sql
```

### Redis Backup

```bash
# Save RDB snapshot
redis-cli -h redis-host -a password SAVE

# Copy dump.rdb file
kubectl cp redis-pod:/data/dump.rdb ./redis-backup.rdb
```

## Troubleshooting

### Common Issues

1. **Pod CrashLoopBackOff**
   ```bash
   kubectl logs -f deployment/auth-service -n auth-service
   # Check for configuration errors or database connectivity
   ```

2. **Database Connection Issues**
   ```bash
   # Test database connectivity
   kubectl exec -it deployment/auth-service -n auth-service -- nc -zv postgres-host 5432
   ```

3. **Redis Connection Issues**
   ```bash
   # Test Redis connectivity
   kubectl exec -it deployment/auth-service -n auth-service -- redis-cli -h redis-host -a password PING
   ```

4. **High Memory Usage**
   ```bash
   # Check memory usage
   kubectl top pods -n auth-service
   # Consider increasing memory limits or optimizing queries
   ```

### Debug Commands

```bash
# Port forward for local testing
kubectl port-forward svc/auth-service 8080:8080 -n auth-service

# Execute into pod
kubectl exec -it deployment/auth-service -n auth-service -- /bin/sh

# View events
kubectl get events -n auth-service --sort-by=.metadata.creationTimestamp

# Describe pod
kubectl describe pod auth-service-pod-name -n auth-service
```

## Security Hardening

### Network Policies

The included network policy restricts traffic:

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
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
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

### Security Context

Pods run with restricted privileges:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1001
  runAsGroup: 1001
  fsGroup: 1001
  readOnlyRootFilesystem: true
```

### Secrets Management

Use Kubernetes secrets or external secret management:

- **AWS Secrets Manager**
- **Azure Key Vault**
- **HashiCorp Vault**
- **GCP Secret Manager**

## Performance Tuning

### Database Optimization

```sql
-- Create indexes for performance
CREATE INDEX CONCURRENTLY idx_users_email_tenant ON users(email, tenant_id);
CREATE INDEX CONCURRENTLY idx_sessions_user_expires ON user_sessions(user_id, expires_at);

-- Analyze tables
ANALYZE users;
ANALYZE user_sessions;
```

### Redis Optimization

```redis
# Redis configuration
maxmemory 512mb
maxmemory-policy allkeys-lru
tcp-keepalive 300
timeout 300
```

### Application Tuning

```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

## Maintenance

### Regular Tasks

1. **Log Rotation**: Configure log aggregation system
2. **Backup Verification**: Test backup restoration monthly
3. **Security Updates**: Keep base images updated
4. **Performance Monitoring**: Monitor key metrics
5. **Database Maintenance**: Run VACUUM and ANALYZE regularly

### Updates

For application updates:

```bash
# Update image
kubectl set image deployment/auth-service auth-service=auth-service:v1.1.0 -n auth-service

# Monitor rollout
kubectl rollout status deployment/auth-service -n auth-service

# Rollback if needed
kubectl rollout undo deployment/auth-service -n auth-service
```

This deployment guide provides a comprehensive setup for production deployment of the authentication service.