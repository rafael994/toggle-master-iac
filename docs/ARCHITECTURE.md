# Architecture

This document describes the architecture of the `toggle-master-iac` platform, a feature flag solution with microservices, infrastructure-as-code, and Kubernetes deployment.

## Goals

- Enable runtime feature flags with flexible targeting.
- Provide observable feature evaluation analytics.
- Secure access with JWT-based authentication.
- Deploy portable stack using AWS services (EKS, RDS, Redis, SQS, DynamoDB).
- Support local development and staging with clear infra definitions.

## High-level components

1. Auth Service (`auth-service/`)
   - Go web service.
   - Handles login/auth validation and JWT issuance.
   - Common entrypoint for API clients with `Authorization: Bearer <token>`.

2. Flag Service (`flag-service/`)
   - Python web API for feature flags CRUD.
   - Manages flags, rules, default values.
   - Persists data to PostgreSQL and Redis cache (local + cloud expected).

3. Targeting Service (`targeting-service/`)
   - Python service with user segmentation logic.
   - Evaluates client context against targeting conditions.
   - Supports segments, environments, rules.

4. Evaluation Service (`evaluation-service/`)
   - Go service, core evaluate pipeline.
   - Receives event requests, resolves flags, decides on rollout.
   - Emits analytics events to SQS and/or direct HTTP to analytics service.

5. Analytics Service (`analytics-service/`)
   - Python worker collects and aggregates evaluation events.
   - Stores metrics in DynamoDB/Redis and exposes reports.

6. Infrastructure / GitOps (`GitOps/` and `terraform/`)
   - Terraform modules for networking, EKS, RDS, Redis, SQS, DynamoDB, ECR.
   - Kubernetes manifests for services, secrets, ingress, auto-scaling.

## Service interaction flow

1. Client authenticates to Auth Service to receive JWT.
2. Client calls Flag/Targeting/Evaluation APIs with JWT bearer token.
3. Evaluation Service verifies token (via Auth if required) and obtains flag rules from Flag service or cache.
4. Targeting Service evaluates user/session attributes against segment definitions.
5. Evaluation result is returned to client and/or SDK.
6. Evaluation Service emits an analytics event to SQS.
7. Analytics Service ingests events, persists metrics, and provides aggregation endpoints.

## CI/CD Pipeline

- **GitHub** hosts the main repository with app code and `GitOps/` manifests.
- **GitHub Actions** workflow:
  - Triggered on every push to main/feature branches.
  - Runs tests, linters, and security scans.
  - Builds Docker images and pushes to ECR with commit SHA tag.
  - (Optional) Creates PR or pushes updated manifest image tags for Argo sync.
- **Argo CD** continuously watches `GitOps/` manifests:
  - Detects new image tags.
  - Automatically syncs manifests to EKS cluster.
  - Provides web UI dashboard for reviewing sync history and manual interventions.

## Data stores

- PostgreSQL (RDS) for user, flags, segment definitions.
- Redis for fast cache, session, rate-limiting, and on-the-fly rule store.
- DynamoDB for analytics/usage metrics in high scale.
- SQS for event buffering between evaluation and analytics.
- Kubernetes ConfigMaps/Secrets for config and sensitive values.

## Deployment

### Terraform

- Use `terraform/bootstrap/envs/dev` for environment-specific configuration.
- `terraform init` + `terraform plan` + `terraform apply` on root + module folders.

### GitHub Actions (CI/CD)

- Triggered on push and PR events.
- Workflow steps:
  1. Check out code.
  2. Run unit/integration tests.
  3. Build Docker images for each service.
  4. Push images to ECR with git commit SHA tag.
  5. Update `GitOps/` manifest image tags and push back to repo (or via pull request).

### Argo CD (GitOps)

- Watches the `GitOps/` directory for changes.
- Automatically syncs manifests to EKS cluster.
- Provides web UI for viewing sync status and manual rollback.
- Configuration: define Argo CD Application CRDs pointing to repo + path.

### Kubernetes

- Deploy services with `kubectl apply -f GitOps/<service>/` (manual) or via Argo CD (automated).
- Secrets and ConfigMaps are managed in `GitOps/<service>/secret.yaml`.
- Autoscaling configured in `hpa.yaml` for CPU-based scaling.

## Local development

- Run each service independently via local venv, `go run`, or containerization.
- Use local databases in `service/db/init.sql` for startup.
- `docker-compose` can be added if needed (not currently in repo).

## Security considerations

- JWT signing key in `auth-service/key.go` should be secret-injected.
- Use AWS IAM roles and least privilege for Terraform-managed services.
- Do not commit credentials; use environment variables or secrets manager.

## Observability

- Add logs and metrics in each service (structured JSON logs recommended).
- Expose health/readiness endpoints (e.g., `/healthz`).
- Centralize with Prometheus/Grafana from EKS observability stack.
