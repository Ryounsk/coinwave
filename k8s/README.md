# Kubernetes Deployment Instructions

This directory contains the Kubernetes manifests to deploy CoinWave.

## Prerequisites

- A running Kubernetes cluster (e.g., Minikube, Docker Desktop K8s, GKE, EKS).
- `kubectl` configured to talk to your cluster.
- Docker images `coinwave-backend:latest` and `coinwave-frontend:latest` available to your cluster nodes. 
  - If using Minikube: `eval $(minikube docker-env)` then build images.
  - If using Docker Desktop: Images built locally are usually available.
  - Otherwise, push images to a registry and update the image names in `04-backend.yaml` and `05-frontend.yaml`.

## Deployment Steps

1.  **Apply Configuration and Secrets:**
    ```bash
    kubectl apply -f k8s/01-config-secrets.yaml
    ```

2.  **Deploy Database (MySQL):**
    ```bash
    kubectl apply -f k8s/02-mysql.yaml
    ```

3.  **Deploy Redis:**
    ```bash
    kubectl apply -f k8s/03-redis.yaml
    ```

4.  **Deploy Backend:**
    ```bash
    kubectl apply -f k8s/04-backend.yaml
    ```

5.  **Deploy Frontend:**
    ```bash
    kubectl apply -f k8s/05-frontend.yaml
    ```

6.  **Verify Deployment:**
    ```bash
    kubectl get pods
    kubectl get services
    ```

7.  **Access the Application:**
    - If using **Docker Desktop** or **LoadBalancer** support: Access via `http://localhost` (Frontend Service Port 80).
    - If using **Minikube**:
      ```bash
      minikube service frontend
      ```

## Architecture

- **MySQL**: StatefulSet with PersistentVolumeClaim for data persistence.
- **Redis**: StatefulSet with PersistentVolumeClaim for caching and rankings.
- **Backend**: Stateless Deployment (2 replicas) connecting to MySQL and Redis.
- **Frontend**: Stateless Deployment (2 replicas) serving the Vue app via Nginx.
