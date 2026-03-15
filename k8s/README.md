# Kubernetes deployment

## Prerequisites

- `kubectl` configured for your cluster
- Docker image `grbac-api` built and available to the cluster (e.g. push to registry or load with `kind load docker-image grbac-api:latest`)

## Apply order

```bash
# 1. Namespace
kubectl apply -f namespace.yaml

# 2. ConfigMap (app config)
kubectl apply -f configmap.yaml

# 3. Secret (edit secret.yaml with real DATABASE_URL and JWT_SECRET first)
kubectl apply -f secret.yaml

# 4. Migrations ConfigMap (from repo root)
kubectl create configmap grbac-migrations --from-file=internal/db/migrations/ -n grbac --dry-run=client -o yaml | kubectl apply -f -

# 5. Run migrations (once)
kubectl apply -f migration-job.yaml
kubectl wait --for=condition=complete job/grbac-migrate -n grbac --timeout=120s

# 6. Deployment & Service
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# 7. (Optional) Ingress
kubectl apply -f ingress.yaml
```

## Using a private image registry

Update `deployment.yaml`: set `image` to e.g. `your-registry.io/grbac-api:0.1.0` and add `imagePullSecrets` if required.

## Postgres in cluster (optional)

For a minimal in-cluster DB, you can run Postgres via Helm or a simple Deployment + Service and point `DATABASE_URL` in the secret to it (e.g. `postgres://user:pass@postgres-service:5432/grbac`).
