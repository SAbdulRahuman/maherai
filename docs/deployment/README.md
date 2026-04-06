# Maher AI — QuantOps — Deployment Guide

> **Status:** Active  
> **Last Updated:** 2026-04-06

## Overview

This guide covers deploying Maher AI — QuantOps across all environments,
from local development to production Kubernetes clusters.

---

## Table of Contents

- [Environments](#environments)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [CI/CD Pipeline](#cicd-pipeline)
- [Infrastructure as Code](#infrastructure-as-code)
- [Monitoring & Observability](#monitoring--observability)
- [Scaling](#scaling)
- [Disaster Recovery](#disaster-recovery)
- [Runbook](#runbook)

---

## Environments

| Environment | Purpose | Infrastructure | Deployment Method |
|-------------|---------|---------------|-------------------|
| **Local** | Development & debugging | Docker Compose | `docker-compose up` |
| **Dev** | Integration testing | K8s namespace | `kubectl apply` / Helm |
| **Staging** | Pre-production validation | K8s cluster (prod-like) | ArgoCD |
| **Production** | Live traffic | K8s cluster (HA, multi-AZ) | ArgoCD (GitOps) |

### Environment Promotion Flow

```
Local → Dev → Staging → Production
  │       │       │          │
  │       │       │          └── ArgoCD auto-sync from main
  │       │       └───────────── ArgoCD manual sync from release/*
  │       └───────────────────── kubectl / Helm from develop
  └───────────────────────────── docker-compose
```

---

## Docker Deployment

### Build Images

```bash
# Build all services
docker-compose build

# Build individual service
docker build -t maherai/ai-engine:latest -f src/infra/docker/Dockerfile.ai-engine .
docker build -t maherai/market-service:latest -f src/infra/docker/Dockerfile.market-service .
docker build -t maherai/gateway:latest -f src/infra/docker/Dockerfile.gateway .
docker build -t maherai/dashboard:latest -f src/infra/docker/Dockerfile.dashboard .
```

### Image Naming Convention

```
ghcr.io/seenimoa/maherai/<service>:<version>

# Examples:
ghcr.io/seenimoa/maherai/ai-engine:v0.1.0
ghcr.io/seenimoa/maherai/market-service:v0.1.0
ghcr.io/seenimoa/maherai/gateway:v0.1.0
ghcr.io/seenimoa/maherai/dashboard:v0.1.0
```

### Docker Compose (Local)

```bash
# Start all services + infrastructure
docker-compose up -d

# Check health
docker-compose ps
docker-compose exec gateway curl http://localhost:8000/health

# View logs
docker-compose logs -f ai-engine

# Scale a service
docker-compose up -d --scale market-service=2

# Stop and clean up
docker-compose down -v
```

---

## Kubernetes Deployment

### Cluster Requirements

| Resource | Dev | Staging | Production |
|----------|-----|---------|------------|
| Nodes | 1–2 | 3 | 5+ (multi-AZ) |
| CPU / node | 4 cores | 4 cores | 8 cores |
| Memory / node | 8 GB | 16 GB | 32 GB |
| Storage | 50 GB | 200 GB | 500 GB+ (SSD) |

### Namespace Strategy

```
maherai-dev        # Development environment
maherai-staging    # Staging environment
maherai-prod       # Production environment
monitoring         # Prometheus, Grafana, Loki (shared)
ingress            # Ingress controller
cert-manager       # TLS certificate management
```

### Deploy with Kustomize

```bash
# Dev environment
kubectl apply -k src/infra/k8s/overlays/dev

# Staging
kubectl apply -k src/infra/k8s/overlays/staging

# Production
kubectl apply -k src/infra/k8s/overlays/prod
```

### Deploy with Helm

```bash
# Add required Helm repos
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Install monitoring stack
helm install prometheus prometheus-community/kube-prometheus-stack \
  -n monitoring --create-namespace \
  -f src/infra/helm/values-prometheus.yaml

helm install loki grafana/loki-stack \
  -n monitoring \
  -f src/infra/helm/values-loki.yaml

# Install application
helm install maherai src/infra/helm/maherai \
  -n maherai-prod --create-namespace \
  -f src/infra/helm/values-prod.yaml
```

### Kubernetes Resource Specifications

```yaml
# Example: maher-ai-engine deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: maher-ai-engine
spec:
  replicas: 2
  template:
    spec:
      containers:
        - name: ai-engine
          image: ghcr.io/seenimoa/maherai/ai-engine:v0.1.0
          resources:
            requests:
              cpu: 500m
              memory: 512Mi
            limits:
              cpu: 2000m
              memory: 2Gi
          ports:
            - containerPort: 8001
          livenessProbe:
            httpGet:
              path: /health
              port: 8001
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8001
            initialDelaySeconds: 10
            periodSeconds: 5
```

---

## CI/CD Pipeline

### Pipeline Stages

```
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│  Lint    │──►│  Test    │──►│  Build   │──►│  Deploy  │──►│  Verify  │
│  (ruff,  │   │  (pytest,│   │  (Docker │   │  (Helm / │   │  (smoke  │
│   eslint)│   │   jest)  │   │   build) │   │   ArgoCD)│   │   tests) │
└──────────┘   └──────────┘   └──────────┘   └──────────┘   └──────────┘
```

### GitHub Actions Workflows

| Workflow | Trigger | Actions |
|----------|---------|---------|
| `ci.yml` | PR to `main`/`develop` | Lint, test, build, security scan |
| `release.yml` | Tag `v*` | Build images, push to GHCR, create release |
| `deploy-staging.yml` | Push to `develop` | Deploy to staging via ArgoCD |
| `deploy-prod.yml` | Push to `main` | Deploy to production via ArgoCD |
| `validate.yml` | PR (docs changes) | Markdown link check |

### GitOps with ArgoCD

```
GitHub Repository (Source of Truth)
         │
         ├── src/infra/k8s/overlays/staging/
         │         │
         │         └──► ArgoCD App (staging)
         │                    │
         │                    └──► K8s Staging Cluster
         │
         └── src/infra/k8s/overlays/prod/
                   │
                   └──► ArgoCD App (production)
                              │
                              └──► K8s Production Cluster
```

---

## Infrastructure as Code

### Terraform

```bash
# Initialize
cd src/infra/terraform
terraform init

# Plan
terraform plan -var-file=environments/prod.tfvars

# Apply
terraform apply -var-file=environments/prod.tfvars
```

### Terraform Modules

| Module | Purpose |
|--------|---------|
| `eks` / `gke` / `aks` | Managed Kubernetes cluster |
| `networking` | VPC, subnets, security groups |
| `database` | Managed PostgreSQL (RDS / Cloud SQL) |
| `redis` | Managed Redis (ElastiCache / Memorystore) |
| `storage` | S3 / GCS buckets for model artifacts |
| `dns` | Route53 / Cloud DNS records |
| `monitoring` | CloudWatch / Stackdriver integration |

---

## Monitoring & Observability

### Dashboards

| Dashboard | Content | Users |
|-----------|---------|-------|
| **System Overview** | CPU, memory, pod status, network | Ops team |
| **API Metrics** | Request rate, latency, error rate | Backend team |
| **AI Engine** | Recommendation count, confidence distribution, LLM latency | AI team |
| **Market Data** | Data freshness, ingestion rate, API health | Data team |
| **Business** | Active users, API consumers, recommendations delivered | Product |

### Alerting Rules

| Alert | Condition | Severity | Channel |
|-------|-----------|----------|---------|
| API Error Rate High | 5xx > 5% for 5 min | Critical | Slack, PagerDuty |
| AI Engine Latency | P95 > 2s for 5 min | Warning | Slack |
| Market Data Stale | No update > 30s during market hours | Critical | Slack, PagerDuty |
| Pod CrashLoop | Restart count > 3 in 10 min | Critical | Slack, PagerDuty |
| Disk Usage High | > 80% on any PV | Warning | Slack |
| Certificate Expiry | < 30 days | Warning | Email |

---

## Scaling

### Horizontal Pod Autoscaler (HPA)

| Service | Min | Max | CPU Target | Memory Target |
|---------|-----|-----|-----------|---------------|
| `gateway` | 2 | 10 | 70% | 80% |
| `market-service` | 1 | 5 | 60% | 70% |
| `ai-engine` | 2 | 8 | 70% | 80% |
| `sentiment` | 1 | 4 | 60% | 70% |
| `alert-service` | 1 | 3 | 50% | 60% |
| `dashboard` | 2 | 6 | 60% | 70% |

### Scaling Strategy

```
Normal Load        Peak (Market Open/Close)    Extreme (Market Crash/Rally)
gateway: 2         gateway: 5                  gateway: 10
ai-engine: 2       ai-engine: 4               ai-engine: 8
market-svc: 1      market-svc: 3              market-svc: 5
```

---

## Disaster Recovery

### Backup Strategy

| Data | Backup Method | Frequency | Retention |
|------|-------------|-----------|-----------|
| PostgreSQL | Automated snapshots | Every 6 hours | 30 days |
| Redis | RDB snapshots | Every 1 hour | 7 days |
| Prometheus TSDB | Volume snapshots | Daily | 90 days |
| AI Model artifacts | S3 versioning | On change | Permanent |
| K8s manifests | Git (source of truth) | On commit | Permanent |

### Recovery Targets

| Metric | Target |
|--------|--------|
| RTO (Recovery Time Objective) | < 15 minutes |
| RPO (Recovery Point Objective) | < 5 minutes |

---

## Runbook

### Common Operations

#### Scale AI Engine

```bash
kubectl scale deployment maher-ai-engine --replicas=5 -n maherai-prod
```

#### Rolling Restart

```bash
kubectl rollout restart deployment maher-ai-engine -n maherai-prod
kubectl rollout status deployment maher-ai-engine -n maherai-prod
```

#### Check Pod Health

```bash
kubectl get pods -n maherai-prod
kubectl describe pod <pod-name> -n maherai-prod
kubectl logs <pod-name> -n maherai-prod --tail=100
```

#### Database Migration

```bash
kubectl exec -it deploy/maher-gateway -n maherai-prod -- python -m alembic upgrade head
```

#### Emergency: Rollback Deployment

```bash
kubectl rollout undo deployment maher-ai-engine -n maherai-prod
kubectl rollout status deployment maher-ai-engine -n maherai-prod
```

---

[Back to README](../../README.md) • [Development Guide](../development/README.md) • [Architecture](../architecture/README.md)
