# 🕐 Temporal on Amazon EKS

> Self-hosted Temporal cluster on Amazon EKS with full RDS (PostgreSQL) persistence, HTTPS ingress, IRSA-based secrets management, and centralized CloudWatch logging.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)
![Temporal](https://img.shields.io/badge/Temporal-self--hosted-000000?style=flat&logo=temporal&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-EKS-326CE5?style=flat&logo=kubernetes&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-RDS-4169E1?style=flat&logo=postgresql&logoColor=white)
![AWS](https://img.shields.io/badge/AWS-ECR%20%7C%20ACM%20%7C%20ALB%20%7C%20CloudWatch-FF9900?style=flat&logo=amazonaws&logoColor=white)

---

## 📋 Table of Contents

- [Overview](#-overview)
- [Architecture](#-architecture)
- [Demo Workflow](#-demo-workflow)
- [Prerequisites](#-prerequisites)
- [Repository Structure](#-repository-structure)
- [Setup Guide](#-setup-guide)
  - [1. Install Go](#1-install-go)
  - [2. Local Temporal Setup](#2-local-temporal-setup)
  - [3. ECR — Push Images](#3-ecr--push-images)
  - [4. VPC Setup](#4-vpc-setup)
  - [5. RDS Setup](#5-rds-setup)
  - [6. EKS Cluster](#6-eks-cluster)
  - [7. kubectl & Helm](#7-kubectl--helm)
  - [8. Deploy Temporal Stack](#8-deploy-temporal-stack)
  - [9. SSL/TLS Certificate](#9-ssltls-certificate)
  - [10. Ingress & Load Balancer](#10-ingress--load-balancer)
  - [11. Domain Setup](#11-domain-setup)
  - [12. Deploy Worker & Starter](#12-deploy-worker--starter)
  - [13. Logging — Fluent Bit + CloudWatch](#13-logging--fluent-bit--cloudwatch)

---

## 🔍 Overview

This project deploys a production-grade, **fully self-hosted Temporal cluster** on AWS using:

| Component | Technology |
|---|---|
| Orchestration | Amazon EKS (Kubernetes) |
| Persistence | Amazon RDS — PostgreSQL |
| Secrets | AWS Secrets Manager + ESO + IRSA |
| Container Registry | Amazon ECR |
| Ingress | AWS ALB + ACM (HTTPS) |
| Logging | Fluent Bit → Amazon CloudWatch |
| Workflow SDK | Go (`go.temporal.io/sdk`) |

---

## 🏗 Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                              AWS VPC                                  │
│                                                                       │
│   Public Subnets                  Private Subnets                    │
│  ┌──────────────────┐           ┌────────────────────────────────┐   │
│  │  ALB / Ingress   │──────────►│         EKS Node Group          │   │
│  │   (HTTPS :443)   │   gRPC    │                                 │   │
│  └──────────────────┘   :7233   │  [temporal-frontend]    :7233  │   │
│                                 │  [temporal-history]             │   │
│                                 │  [temporal-matching]            │   │
│                                 │  [temporal-worker]              │   │
│                                 │  [temporal-ui]          :8080  │   │
│                                 │                                 │   │
│                                 │  [demo-order-worker]   ────────►│   │
│                                 │  [demo-workflow-starter]        │   │
│                                 └────────────────┬────────────────┘   │
│                                                  │ :5432              │
│                                 ┌────────────────▼────────────────┐   │
│                                 │    RDS PostgreSQL (Single-AZ)   │   │
│                                 │    DB: temporal                 │   │
│                                 │    DB: temporal_visibility      │   │
│                                 └─────────────────────────────────┘   │
│                                                                       │
│   [ECR]  [Secrets Manager]  [CloudWatch]  [IAM / IRSA]  [ACM]       │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🛒 Demo Workflow

An **Order Processing** workflow that demonstrates durable execution and automatic retries:

```
START
  │
  ▼  ValidateOrder          (validateOrder.go)
     • Check item availability (mock)
     • Retry: max 3 attempts, 2s backoff
  │
  ▼  ReserveInventory       (ReserveInventory.go)
     • Deduct stock (mock DB write)
     • Heartbeat every 2s
     • Retry: max 5 attempts, 5s backoff
  │
  ▼  ChargePayment          (ChargePayment.go)  ← intentionally fails 50% of the time
     • Simulate payment gateway
     • Retry: max 3 attempts, 10s backoff
  │
  ▼  SendConfirmationEmail  (ConfirmationEmail.go)
     • Logs "email sent" to stdout
  │
  ▼  COMPLETE → returns { OrderID, Status }
```

> **`ChargePayment` fails 50% of the time by design.** Temporal retries it automatically — no manual intervention needed.

---

## ✅ Prerequisites

- AWS CLI configured with sufficient IAM permissions
- Docker installed locally
- Go 1.21+
- A registered domain name (for HTTPS ingress via ACM + Route 53)

---

## 📁 Repository Structure

```
temporal/
├── .gitignore
├── README.md
│
├── starter/                         # HTTP API server — triggers workflow executions
│   ├── api/
│   │   └── handler.go               # Route handlers (POST /order, etc.)
│   ├── cmd/
│   │   └── main.go                  # Starter entry point
│   ├── models/
│   │   └── models.go                # Request/response structs
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── worker/                          # Workflow executor — runs activities & workflows
│   ├── activities/
│   │   ├── ChargePayment.go         # Simulates payment gateway (50% failure rate)
│   │   ├── ConfirmationEmail.go     # Logs confirmation email to stdout
│   │   ├── ReserveInventory.go      # Mock stock deduction with heartbeating
│   │   └── validateOrder.go         # Mock item availability check
│   ├── cmd/
│   │   └── main.go                  # Worker entry point
│   ├── models/
│   │   └── models.go                # Shared data structs
│   ├── workflows/
│   │   └── OrderWorkflow.go         # Orchestrates all 4 activities in sequence
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
└── k8s deployments/                 # All Kubernetes & Helm manifests
    ├── values.yaml                  # Temporal Helm overrides (RDS, IRSA, schema setup)
    ├── temporal-sa.yaml             # Service account annotated with IRSA role
    ├── temporal-secret.yaml         # SecretProviderClass (CSI driver)
    ├── secret-store.yaml            # ESO SecretStore → AWS Secrets Manager
    ├── external-secret.yaml         # ESO ExternalSecret → creates K8s Secret
    ├── ingress.yaml                 # ALB Ingress (HTTPS, ACM cert)
    ├── worker-deployment.yaml
    ├── starter-deployment.yaml
    ├── starter-service.yaml
    ├── fluent-bit-sa.yaml
    ├── fluent-bit-config.yaml
    └── fluent-bit-daemonset.yaml
```

---

## 🚀 Setup Guide

### 1. Install Go

```bash
# Extract to /usr/local
tar -C /usr/local -xzf go<version>.linux-amd64.tar.gz

# Add to PATH (append to ~/.profile)
export PATH=$PATH:/usr/local/go/bin

# Install Temporal SDK in each module
cd starter && go get go.temporal.io/sdk
cd ../worker && go get go.temporal.io/sdk
```

---

### 2. Local Temporal Setup

```bash
# Download Temporal CLI
curl -O "https://temporal.download/cli/archive/latest?platform=linux&arch=amd64"
tar -xzvf temporal_cli_<version>_linux_amd64.tar.gz
cp temporal /usr/local/bin/

# Start with in-memory store (data lost on restart)
temporal server start-dev

# Start with file-backed persistence
temporal server start-dev --db-filename demo.db
```

UI available at: `http://localhost:8233`

> Temporal needs two stores: a **default store** (execution history, tasks) and a **visibility store** (search/filter data).

---

### 3. ECR — Push Images

`starter` and `worker` each have their own `Dockerfile` and are pushed as independent images so they can be scaled separately.

```bash
# Authenticate Docker with ECR
aws ecr get-login-password --region us-east-2 \
  | docker login --username AWS --password-stdin \
    <account-id>.dkr.ecr.us-east-2.amazonaws.com

# --- Worker ---
docker build -t temporal-worker ./worker
docker tag temporal-worker \
  <account-id>.dkr.ecr.us-east-2.amazonaws.com/temporal-worker:latest
docker push \
  <account-id>.dkr.ecr.us-east-2.amazonaws.com/temporal-worker:latest

# --- Starter ---
docker build -t temporal-starter ./starter
docker tag temporal-starter \
  <account-id>.dkr.ecr.us-east-2.amazonaws.com/temporal-starter:latest
docker push \
  <account-id>.dkr.ecr.us-east-2.amazonaws.com/temporal-starter:latest
```

---

### 4. VPC Setup

Create a VPC with:
- **2 public subnets** (2 AZs) — for the ALB
- **2 private subnets** (2 AZs) — for EKS nodes and RDS
- **NAT Gateways** — so private nodes can reach the internet

> ALB requires at least 2 subnets in different AZs.

---

### 5. RDS Setup

1. Create a **PostgreSQL** RDS instance in the private subnets.
2. Create a **DB Subnet Group** targeting both private subnets.
3. Store credentials using **AWS Secrets Manager** (KMS-encrypted).
4. Two databases are needed: `temporal` and `temporal_visibility`.

> Settings used here: Dev/Test template, Single-AZ, `db.t3.medium`, 30 GB storage.

---

### 6. EKS Cluster

<details>
<summary><strong>IAM Role for the Cluster</strong></summary>

Create an IAM role with:
- **Trust policy**: `eks.amazonaws.com`
- **Attached policy**: `AmazonEKSClusterPolicy`

</details>

<details>
<summary><strong>Cluster Add-ons</strong></summary>

| Add-on | Purpose |
|---|---|
| Amazon VPC CNI | Pod networking |
| CoreDNS | In-cluster DNS / service discovery |
| kube-proxy | Network routing for Services |
| Node Monitoring Agent | Node health monitoring |
| EKS Pod Identity Agent | IAM for pods via Service Accounts |
| AWS Secrets Store CSI Driver | Mount Secrets Manager secrets into pods |

</details>

<details>
<summary><strong>Node Group Configuration</strong></summary>

| Setting | Value |
|---|---|
| Capacity type | On-Demand |
| Instance type | `t2.xlarge` |
| Disk size | 40 GB |
| Subnets | Private subnets only |

</details>

<details>
<summary><strong>Access Entry (RBAC)</strong></summary>

Create an IAM role `TemporalClusterAccessRole` with:
- Trust principal: `ec2.amazonaws.com`
- Policies: `AmazonEKSClusterAdminPolicy`, `EKSDescribeClusterPolicy`

Add it as an EKS **Access Entry** with `AmazonEKSClusterAdminPolicy`, then attach the role to your management EC2 instance.

</details>

<details>
<summary><strong>OIDC Setup</strong></summary>

Register the EKS OIDC provider in IAM → Identity Providers. This is required for IRSA to work.

</details>

---

### 7. kubectl & Helm

```bash
# kubectl
sudo snap install kubectl --classic
aws eks --region us-east-2 update-kubeconfig --name Temporal-Cluster

# Helm + Temporal repo
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
helm repo add temporalio https://temporalio.github.io/helm-charts
helm repo update
```

---

### 8. Deploy Temporal Stack

#### Create namespace

```bash
kubectl create namespace temporal
```

#### IRSA — Secrets Manager access

Create IAM role `TemporalSecretManagerSARole`:
- **Policy**: `AWSSecretsManagerClientReadOnlyAccess`
- **Trust policy**: scoped to `temporal-sa` service account in the `temporal` namespace

```bash
kubectl apply -f "k8s deployments/temporal-sa.yaml"
```

#### External Secrets Operator (ESO)

ESO pre-populates a Kubernetes Secret from Secrets Manager **before** the Temporal Helm chart runs its schema job — avoiding a race condition where the schema setup fails because RDS credentials aren't mounted yet.

```bash
# Install CRDs
kubectl apply -f https://raw.githubusercontent.com/external-secrets/external-secrets/main/deploy/crds/bundle.yaml --server-side

# Install ESO
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets --create-namespace --set installCRDs=false

# Deploy SecretStore and ExternalSecret
kubectl apply -f "k8s deployments/secret-store.yaml"
kubectl apply -f "k8s deployments/external-secret.yaml"
```

#### Deploy Temporal via Helm

```bash
helm install temporal temporal/temporal \
  -n temporal \
  -f "k8s deployments/values.yaml" \
  --timeout 900s

# Verify all pods are running
kubectl get pods -n temporal
```

`values.yaml` configures: service account binding, CSI volume mounts, RDS connection, env var mappings, and schema setup jobs for both `temporal` and `temporal_visibility` databases.

---

### 9. SSL/TLS Certificate

1. Request a public certificate in **AWS Certificate Manager (ACM)** for your domain (e.g. `temporal.yourdomain.com`).
2. Add the CNAME record provided by ACM to **Route 53** to complete domain validation.

---

### 10. Ingress & Load Balancer

```bash
# Download and create the IAM policy
curl -O https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/main/docs/install/iam_policy.json
aws iam create-policy \
  --policy-name AWSLoadBalancerControllerIAMPolicy \
  --policy-document file://iam_policy.json

# Install the controller
helm repo add eks https://aws.github.io/eks-charts
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=Temporal-Cluster \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller \
  --set region=us-east-2 \
  --set vpcId=<your-vpc-id>

# Deploy Ingress
kubectl apply -f "k8s deployments/ingress.yaml"
```

The Ingress exposes:
- **Temporal Web UI** — `https://temporal.yourdomain.com`
- **Starter API** — `https://temporal.yourdomain.com/order`

---

### 11. Domain Setup

In Route 53, create an **A record (Alias)** for `temporal.yourdomain.com` pointing to the ALB DNS name generated by the ingress controller.

---

### 12. Deploy Worker & Starter

```bash
# Worker — connects to Temporal and executes workflows/activities
kubectl apply -f "k8s deployments/worker-deployment.yaml"

# Starter — HTTP server listens on :8080, triggers workflows on POST /order
kubectl apply -f "k8s deployments/starter-deployment.yaml"
kubectl apply -f "k8s deployments/starter-service.yaml"
```

Test the workflow:

```bash
curl -X POST https://temporal.yourdomain.com/order
```

---

### 13. Logging — Fluent Bit + CloudWatch

```bash
kubectl create namespace amazon-cloudwatch

kubectl apply -f "k8s deployments/fluent-bit-sa.yaml"
kubectl apply -f "k8s deployments/fluent-bit-config.yaml"
kubectl apply -f "k8s deployments/fluent-bit-daemonset.yaml"
```

Fluent Bit runs as a **DaemonSet** (one pod per node), reads container logs from `/var/log/containers/`, filters for `starter` and `worker` pods only, and ships them to **CloudWatch Log Groups**.
