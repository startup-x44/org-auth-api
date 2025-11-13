# Auth Service Kubernetes Deployment

This directory contains Kubernetes manifests for deploying the auth-service in a production environment.

## Prerequisites

- Kubernetes cluster (v1.19+)
- kubectl configured to access the cluster
- NGINX Ingress Controller installed
- cert-manager installed (for automatic TLS)
- PostgreSQL and Redis databases available

## Components

### Core Resources
- **namespace.yaml**: Creates the `auth-service` namespace
- **serviceaccount.yaml**: Service account with minimal permissions
- **configmap.yaml**: Non-sensitive configuration
- **secret.yaml**: Sensitive configuration (base64 encoded)
- **deployment.yaml**: Application deployment with health checks and security
- **service.yaml**: ClusterIP service for internal communication
- **ingress.yaml**: External access with TLS termination

### Scaling & Reliability
- **hpa.yaml**: Horizontal Pod Autoscaler for automatic scaling
- **poddisruptionbudget.yaml**: Ensures minimum availability during maintenance

### Security
- **networkpolicy.yaml**: Restricts network traffic to necessary services only

### Configuration
- **kustomization.yaml**: Kustomize configuration for environment management
- **health-checks.md**: Documentation for required health check endpoints

## Deployment

### Using kubectl

```bash
# Deploy to Kubernetes
kubectl apply -f k8s/

# Or using kustomize
kubectl apply -k k8s/
```

### Using Helm (alternative)

If you prefer Helm, you can create a Helm chart based on these manifests.

## Configuration

### Environment Variables

The application is configured through environment variables from ConfigMaps and Secrets:

**Required Secrets** (update before deployment):
- `JWT_SECRET`: Strong random string for JWT signing
- `DATABASE_*`: Database connection details
- `REDIS_*`: Redis connection details
- `EMAIL_API_KEY`: Email service API key

**Configurable via ConfigMap**:
- Rate limiting settings
- Token expiry times
- Database connection pool settings

### Database Setup

Ensure PostgreSQL and Redis are running and accessible:

```bash
# Example: Deploy PostgreSQL (for development only)
kubectl run postgres --image=postgres:13 --env="POSTGRES_PASSWORD=password" --port=5432

# Example: Deploy Redis (for development only)
kubectl run redis --image=redis:7 --port=6379
```

For production, use managed database services or proper StatefulSets.

## Health Checks

The deployment includes readiness and liveness probes that expect these endpoints:

- `GET /health/live` - Liveness probe (basic health)
- `GET /health/ready` - Readiness probe (dependencies check)
- `GET /metrics` - Prometheus metrics

See `health-checks.md` for implementation details.

## Monitoring

The service is configured for Prometheus monitoring:
- Metrics endpoint: `/metrics`
- Service annotations for automatic discovery

## Security Features

- **Network Policies**: Restrict traffic to necessary services
- **Security Context**: Non-root user, read-only filesystem
- **Resource Limits**: Prevent resource exhaustion
- **TLS**: Automatic certificate management via cert-manager
- **Security Headers**: Configured in Ingress

## Scaling

The HPA automatically scales based on CPU (70%) and memory (80%) utilization:
- Minimum: 3 replicas
- Maximum: 10 replicas

## Troubleshooting

### Check pod status
```bash
kubectl get pods -n auth-service
kubectl describe pod <pod-name> -n auth-service
```

### Check logs
```bash
kubectl logs -f deployment/auth-service -n auth-service
```

### Check ingress
```bash
kubectl get ingress -n auth-service
kubectl describe ingress auth-service-ingress -n auth-service
```

### Port forward for testing
```bash
kubectl port-forward svc/auth-service 8080:80 -n auth-service
```

## Production Considerations

1. **Secrets Management**: Use external secret management (Vault, AWS Secrets Manager, etc.)
2. **Database**: Use managed PostgreSQL (RDS, Cloud SQL, etc.)
3. **Redis**: Use managed Redis (ElastiCache, Memorystore, etc.)
4. **Monitoring**: Set up proper monitoring and alerting
5. **Backup**: Configure database backups
6. **Updates**: Use rolling updates with proper testing

## Development

For local development with minikube:

```bash
# Start minikube
minikube start

# Enable ingress
minikube addons enable ingress

# Deploy
kubectl apply -k k8s/

# Get service URL
minikube service auth-service -n auth-service --url
```