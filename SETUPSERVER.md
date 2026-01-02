# Server Setup & Deployment Guide

This guide describes how to set up the server and deploy the **Petualangan Cuan** application using Docker and GitHub Actions.

## Prerequisites

1.  **Server (VPS/Cloud)**: Ubuntu 20.04/22.04 LTS (recommended).
2.  **Domain**: A valid domain name pointing to your server IP.
3.  **Docker & Docker Compose**: Installed on the server.

## 1. Manual Deployment (Docker Compose)

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

# Verify installation
sudo docker --version
sudo docker compose version
```

### Step 2: Set Project & Environment Variables

1.  Clone the repository:
    ```bash
    git clone https://github.com/alimurrofid/petualangan-cuan.git
    cd petualangan-cuan
    ```

2.  Create `.env` file from example (or use the configured variables in CI/CD):
    The `docker-compose.yml` reads environment variables. You can create a `.env` file in the root directory:
    ```env
    DB_USER=postgres
    DB_PASSWORD=secure_password
    DB_NAME=petualangan_cuan
    JWT_SECRET=very_secure_secret
    GOOGLE_CLIENT_ID=...
    GOOGLE_CLIENT_SECRET=...
    GOOGLE_REDIRECT_URL=https://your-domain.com/auth/google/callback
    FRONTEND_URL=https://your-domain.com
    ```

### Step 3: Run the Application
```bash
sudo docker compose up -d --build
```
- Access Frontend: `http://<your-ip>:5173` (or port 80 if configured)
- Access Backend: `http://<your-ip>:8080`

---


## 2. CI/CD Pipeline (GitHub Actions)

This project is configured for automated deployment using GitHub Actions. The workflow file is located at [`.github/workflows/deploy.yml`](.github/workflows/deploy.yml).

### Setup GitHub Secrets
Go to **Settings > Secrets and variables > Actions** in your GitHub repository and add:

- `SERVER_IP`: IP address of your VPS.
- `SERVER_USER`: Username (e.g., `root` or `ubuntu`).
- `SSH_PRIVATE_KEY`: Private SSH key to access the server.
- `ENV_FILE`: The full content of your production `.env` file.

### Workflow Explanation
The workflow performs the following steps:
1.  **Checkout**: Pulls the latest code.
2.  **SCP**: Copies the project files to the server.
3.  **SSH**: Connects to the server, updates the `.env` file with secrets, and restarts Docker containers.
