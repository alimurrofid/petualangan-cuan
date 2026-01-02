# Server Setup & Deployment Guide

This guide describes how to set up the server and deploy the **Petualangan Cuan** application using the optimized Docker and GitHub Actions pipeline.

## Prerequisites

1.  **Server (VPS/Cloud)**: Ubuntu 20.04/22.04 LTS (recommended).
2.  **Domain**: A valid domain name pointing to your server IP.
3.  **Docker & Docker Compose**: Installed on the server.
4.  **GitHub Package Access**: A GitHub Personal Access Token (PAT) with `read:packages` scope (if repository is private).

## 1. Initial Server Setup (One-Time)

### Step 1: Install Docker
Run the following commands on your server:

```bash
# Update packages
sudo apt update && sudo apt upgrade -y

# Install prerequisites
sudo apt install apt-transport-https ca-certificates curl software-properties-common -y

# Add Docker GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Add Docker repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io docker-compose-plugin -y

# Verify
docker compose version
```

### Step 2: Authenticate with GitHub Container Registry
To allow the server to pull images from GitHub Container Registry (GHCR):

1.  Generate a PAT on GitHub (Settings > Developer settings > Personal access tokens) with `read:packages`.
2.  Log in on your server:
    ```bash
    echo "YOUR_GITHUB_PAT" | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin
    ```

### Step 3: Setup Project Directory & Environment
1.  Create the project directory:
    ```bash
    mkdir -p /home/ubuntu/petualangan-cuan
    cd /home/ubuntu/petualangan-cuan
    ```

2.  **Crucial**: Create the `.env` file manually.
    ```bash
    nano .env
    ```
    Paste your production environment variables:
    ```env
    # Database
    DB_USER=postgres
    DB_PASSWORD=your_secure_db_password
    DB_NAME=petualangan_cuan
    
    # Backend Config
    PORT=8080
    JWT_SECRET=your_very_secure_jwt_secret
    JWT_EXPIRY_HOURS=24
    
    # OAuth
    GOOGLE_CLIENT_ID=...
    GOOGLE_CLIENT_SECRET=...
    GOOGLE_REDIRECT_URL=https://your-domain.com/auth/google/callback
    FRONTEND_URL=https://your-domain.com
    ```
    > **Security Note**: This file is excluded from git and should never be committed.

## 2. CI/CD Deployment (GitHub Actions)

The project uses a **Container Registry** workflow. Images are built in GitHub Actions, pushed to GHCR, and then pulled by the server.

### Setup GitHub Secrets
Go to **Settings > Secrets and variables > Actions** in your GitHub repository and add:

- `SERVER_IP`: IP address of your VPS.
- `SERVER_USER`: Username (e.g., `ubuntu`).
- `SSH_PRIVATE_KEY`: Private SSH key to access the server.

### How Deployment Works
1.  **Build & Push**: Commits to `main` trigger a build. Images are tagged and pushed to `ghcr.io/<user>/petualangan-cuan/backend:latest` and `frontend:latest`.
2.  **Deploy**: The workflow uses SSH to execute:
    ```bash
    docker compose pull
    docker compose up -d --remove-orphans
    docker image prune -f
    ```

## 3. Troubleshooting

- **Check Logs**:
  ```bash
  docker compose logs -f backend
  docker compose logs -f frontend
  ```
- **Permission Errors**: Ensure the user running docker commands is in the `docker` group or use `sudo`.
