# toggle-master-iac

Toggle Master (toggle-master-iac) is a microservices-based feature flags platform with infrastructure as code for local and AWS deployment.

## 🚀 Overview

- **Objective:** provide a feature toggle service with auth, evaluation, targeting, analytics, and flag management.
- **Architecture:** 5 services (auth, analytics, evaluation, flag, targeting) plus shared infrastructure using Terraform.
- **Infra:** EKS cluster, RDS PostgreSQL, Redis, SQS, DynamoDB, ECR, and IAM.

## 📦 Repository layout

- `auth-service/` - Go API for user auth, token issuance, and client validation.
- `evaluation-service/` - Go service that processes flag evaluation events and sends analytics.
- `analytics-service/` - Python service for collecting/aggregating usage metrics.
- `flag-service/` - Python service for flag CRUD and evaluation rules.
- `targeting-service/` - Python service with segmentation and targeting logic.
- `GitOps/` - Kubernetes manifests for deploy (deployments, services, namespaces, ingress, HPA).
- `terraform/` - IaC for AWS setup and environment bootstrap.

## 🛠️ Local development

1. `cd <service>`
2. For Python services: `pip install -r requirements.txt`
3. For Go services: `go mod tidy`
4. Run with `python app.py` (or `go run .` for Go services)
5. Use local `docker-compose` or Kubernetes minikube as needed.

## ☁️ AWS deployment (high level)

1. In `terraform/bootstrap/envs/dev`, configure `backend.tf` and `providers.tf` with AWS profile/region.
2. `terraform init && terraform apply` in root/`terraform` and module directories.
3. Build and push container images to ECR.
4. Apply Kubernetes YAML from `GitOps/`.

## 🔐 Authentication

- Auth service uses JWT and `auth-service/key.go` for signing.
- Ensure AWS IAM role is configured (lab role via environment variables) for deployment automation.

## 🧪 Testing

- Add tests in each service folder (`*_test.go` for Go, `pytest`/`unittest` for Python).
- For infra, use `terraform validate` and `terraform plan`.

## 📌 Notes

- This repository is on branch `feature/documentation` and is based on `main`.
- Keep secrets out of source; use Kubernetes Secrets or AWS Secrets Manager.

## 📘 Extras

- `auth-service/db/init.sql`, `flag-service/db/init.sql`, `targeting-service/db/init.sql` initialize local DB schemas.

---

### Quick commands

- `git clone <repo>`
- `cd terraform && terraform init && terraform apply`
- `kubectl apply -f GitOps/<service>/` 
