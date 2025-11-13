# CI/CD Pipeline Documentation

This document describes the comprehensive CI/CD pipeline implemented for the authentication microservice.

## Overview

The CI/CD pipeline consists of multiple GitHub Actions workflows that automate the entire software delivery lifecycle, from code quality checks to production deployment.

## Workflows

### 1. CI Pipeline (`ci.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Jobs:**
- **test**: Runs unit and integration tests with PostgreSQL and Redis
- **lint**: Code quality checks with golangci-lint
- **security**: Security scanning with Trivy and Gosec
- **build**: Docker image building and basic validation

**Key Features:**
- Race condition detection in tests
- Code coverage reporting
- Multi-stage Docker builds
- Security vulnerability scanning

### 2. CD Pipeline (`cd.yml`)

**Triggers:**
- Successful completion of CI pipeline on `main` branch
- Manual deployment trigger

**Jobs:**
- **build-and-push**: Multi-platform Docker image building and registry push
- **deploy-staging**: Automated deployment to staging environment
- **deploy-production**: Manual approval required for production deployment
- **health-check**: Post-deployment health verification

**Key Features:**
- Blue-green deployment strategy
- Automated rollback on failure
- Health checks and monitoring
- Slack notifications

### 3. Security Pipeline (`security.yml`)

**Triggers:**
- Daily schedule (9 AM UTC)
- Manual trigger
- Security-related events

**Jobs:**
- **vulnerability-scan**: Comprehensive vulnerability scanning
- **dependency-check**: Go module vulnerability assessment
- **secrets-detection**: Sensitive data detection
- **compliance-check**: Security compliance validation

**Key Features:**
- Automated dependency updates via PR
- Integration with security dashboards
- Compliance reporting

### 4. Code Quality Pipeline (`quality.yml`)

**Triggers:**
- Push to feature branches
- Pull requests

**Jobs:**
- **formatting**: Code formatting validation
- **complexity**: Cyclomatic complexity analysis
- **coverage**: Test coverage validation (80% minimum)
- **static-analysis**: Advanced static code analysis

**Key Features:**
- Automated code formatting
- Complexity thresholds enforcement
- Coverage reporting and trends

### 5. Release Pipeline (`release.yml`)

**Triggers:**
- Git tag creation (semantic versioning: `v*.*.*`)

**Jobs:**
- **release**: Automated release creation with binaries
- **docker-release**: Multi-platform Docker image release

**Key Features:**
- Cross-platform binary builds (Linux, macOS, Windows)
- Automated checksum generation
- GitHub releases with release notes
- Container registry publishing

### 6. Dependency Updates (`dependency-updates.yml`)

**Triggers:**
- Weekly schedule (Mondays 9 AM UTC)
- Manual trigger

**Jobs:**
- **update-dependencies**: Go module updates
- **update-docker-base**: Docker base image updates

**Key Features:**
- Automated PR creation for updates
- Test validation before merge
- Security patch automation

### 7. Monitoring (`monitoring.yml`)

**Triggers:**
- Every 6 hours
- Manual trigger

**Jobs:**
- **health-check**: Service health validation
- **performance-monitoring**: Benchmark execution
- **security-monitoring**: Ongoing security scans
- **dependency-vulnerability-check**: Vulnerability monitoring

**Key Features:**
- Automated health monitoring
- Performance regression detection
- Security posture monitoring

### 8. Database Migration (`database-migration.yml`)

**Triggers:**
- Manual trigger with environment selection

**Jobs:**
- **migrate-database**: Controlled database migrations
- **rollback-on-failure**: Automatic rollback on migration failure

**Key Features:**
- Environment-specific migrations
- Migration status tracking
- Automated rollback procedures

### 9. Backup & Recovery (`backup-recovery.yml`)

**Triggers:**
- Daily schedule (2 AM UTC)
- Manual trigger

**Jobs:**
- **database-backup**: Automated database backups
- **configuration-backup**: Configuration and manifest backups
- **test-recovery**: Backup integrity validation
- **cleanup-old-backups**: Backup retention management

**Key Features:**
- Encrypted backup storage (S3)
- Backup integrity verification
- Automated cleanup policies

### 10. Incident Response (`incident-response.yml`)

**Triggers:**
- Manual trigger for incident declaration

**Jobs:**
- **assess-incident**: Incident assessment and logging
- **isolate-affected-systems**: System isolation for critical incidents
- **collect-forensics**: Forensic data collection for security incidents
- **notify-stakeholders**: Automated stakeholder notifications
- **create-incident-response-plan**: Structured response planning

**Key Features:**
- Severity-based response automation
- Forensic evidence collection
- Stakeholder communication
- Response plan generation

## Environment Configuration

### Required Secrets

Set these secrets in your GitHub repository:

```bash
# AWS Configuration
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
AWS_REGION
EKS_CLUSTER_NAME

# Docker Registry
DOCKER_USERNAME
DOCKER_PASSWORD
DOCKER_REGISTRY

# Database
DB_HOST
DB_PORT
DB_NAME
DB_USERNAME
DB_PASSWORD

# Redis
REDIS_URL
REDIS_PASSWORD

# Monitoring
SLACK_WEBHOOK_URL
PAGERDUTY_INTEGRATION_KEY

# Backup
BACKUP_BUCKET
BACKUP_ENCRYPTION_KEY
```

### Required Environment Variables

```bash
# Staging Environment
STAGING_K8S_NAMESPACE
STAGING_DB_HOST
STAGING_REDIS_URL

# Production Environment
PRODUCTION_K8S_NAMESPACE
PRODUCTION_DB_HOST
PRODUCTION_REDIS_URL
```

## Usage Guide

### Deploying to Staging

Staging deployments happen automatically when code is merged to `main`:

1. CI pipeline runs and validates code
2. If successful, CD pipeline deploys to staging
3. Health checks verify deployment success
4. Notifications sent to team

### Deploying to Production

Production deployments require manual approval:

1. Ensure staging deployment is successful
2. Go to Actions → CD → Run workflow
3. Select "production" environment
4. Approve the deployment
5. Monitor deployment progress

### Handling Incidents

For production incidents:

1. Go to Actions → Incident Response → Run workflow
2. Select incident type and severity
3. Provide incident description
4. Follow the generated response plan

### Database Migrations

To run database migrations:

1. Go to Actions → Database Migration → Run workflow
2. Select target environment
3. Choose migration command (up/down/status)
4. Monitor execution

## Monitoring and Alerts

### Health Monitoring

- Service health checked every 6 hours
- Automated alerts on failures
- Performance benchmarks tracked

### Security Monitoring

- Daily vulnerability scans
- Automated dependency updates
- Security incident alerting

### Backup Monitoring

- Daily backup verification
- Backup integrity checks
- Retention policy enforcement

## Troubleshooting

### Common Issues

1. **CI Pipeline Failures**
   - Check test logs for failing tests
   - Verify code formatting with `gofmt`
   - Ensure dependencies are properly vendored

2. **Deployment Failures**
   - Check Kubernetes pod logs
   - Verify environment secrets
   - Confirm network connectivity

3. **Security Scan Failures**
   - Review vulnerability reports
   - Update dependencies
   - Implement security fixes

### Debugging Workflows

- Use `workflow_dispatch` for manual runs
- Check workflow logs in Actions tab
- Use artifacts for detailed reports
- Enable debug logging with `ACTIONS_RUNNER_DEBUG=true`

## Best Practices

1. **Branch Protection**: Require CI checks on main branch
2. **Code Reviews**: Mandatory reviews for production changes
3. **Testing**: Maintain >80% code coverage
4. **Security**: Regular dependency updates and scans
5. **Monitoring**: Active monitoring and alerting
6. **Documentation**: Keep runbooks and procedures updated

## Contributing

When adding new features:

1. Ensure CI passes for your changes
2. Add appropriate tests
3. Update documentation
4. Follow security best practices
5. Test deployments in staging first

## Support

For issues with the CI/CD pipeline:

1. Check workflow logs in GitHub Actions
2. Review this documentation
3. Contact the DevOps team
4. Create an issue with detailed information