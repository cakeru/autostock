# Deployment Guide

## Overview

AutoStock can be deployed in multiple ways depending on your needs:
1. **Docker Compose (Recommended)** - All-in-one deployment on a single VPS
2. **Split Deployment** - Frontend on Vercel, backend + DB on VPS
3. **Manual Deployment** - Direct binary deployment without Docker

## Deploying on a Mac Mini (Apple Silicon)

This is the setup when the shop's server is a Mac Mini M4 instead of a Linux
VPS. Everything below still applies, with these differences:

- **No Coolify.** Coolify is a Linux-only control panel; on macOS run the plain
  `docker-compose.yml` with Docker Desktop (or OrbStack). The Settings →
  **"Update now"** button (the `updater` container) is your update path.
- **Keep the repo under your home folder.** Docker Desktop only shares paths
  under `$HOME` (and `/tmp`) into the Docker VM. Clone to e.g.
  `~/autostock` — a repo in `/opt` or `/srv` will fail with "Mounts denied".
- **The Docker socket lives at a per-user path.** Docker Desktop for Mac keeps
  it at `~/.docker/run/docker.sock` (not `/var/run/docker.sock`). In `.env` set
  the literal path:
  ```bash
  REPO_DIR=/Users/yourname/autostock
  DOCKER_SOCKET=/Users/yourname/.docker/run/docker.sock
  BACKUP_DIR=/Users/yourname/autostock/backups
  ```
  (Alternatively enable "Allow the default Docker socket to be used" in Docker
  Desktop settings and leave `DOCKER_SOCKET` unset.)
- **Arm64 images build natively** on the M4 — no emulation. The frontend build
  already avoids x64-only packages.
- **Docker must be running** for the stack to start: enable "Start Docker
  Desktop when you sign in", and log into the Mac once after a reboot. The
  containers use `restart: unless-stopped`, so they come back automatically
  once Docker Desktop is up.
- **Shop access:** the frontend listens on `0.0.0.0:3000`, so staff can reach
  the app from other devices at `http://<mac-ip>:3000` on the shop LAN. The
  backend (8080) and Postgres (5433) stay bound to localhost.

Otherwise follow Option 1 below: copy `.env.example` → `.env` with a strong
`DB_PASSWORD` and a 32+ char `JWT_SECRET`, run `docker compose up -d --build`,
log in with `admin` / `admin123`, and change the password immediately.

## Prerequisites

- **Domain name** (optional but recommended)
- **VPS** with at least 1GB RAM, 20GB storage (DigitalOcean, Vultr, Linode, etc.)
- **Docker** and **Docker Compose** installed on the VPS
- **SSL certificate** (Let's Encrypt, free)

## Option 1: Docker Compose (Recommended)

This is the simplest deployment option. Everything runs in containers on a single VPS.

### Architecture

```
┌─────────────────────────────────────────┐
│              VPS (Ubuntu)               │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │        Docker Compose             │ │
│  │                                   │ │
│  │  ┌─────────┐  ┌───────────────┐  │ │
│  │  │  Nginx  │  │   Frontend    │  │ │
│  │  │  :80    │  │   (React)     │  │ │
│  │  │  :443   │  │   :3000       │  │ │
│  │  └────┬────┘  └───────────────┘  │ │
│  │       │                           │ │
│  │  ┌────┴────┐  ┌───────────────┐  │ │
│  │  │ Reverse │  │   Backend     │  │ │
│  │  │ Proxy   │  │   (Go API)    │  │ │
│  │  │         │  │   :8080       │  │ │
│  │  └─────────┘  └───────┬───────┘  │ │
│  │                       │           │ │
│  │               ┌───────┴───────┐   │ │
│  │               │  PostgreSQL   │   │ │
│  │               │   :5432       │   │ │
│  │               └───────────────┘   │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### Step 1: Prepare VPS

```bash
# SSH into your VPS
ssh root@your-vps-ip

# Update system
apt update && apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh

# Install Docker Compose
apt install docker-compose-plugin -y

# Create application directory
mkdir -p /opt/autostock
cd /opt/autostock
```

### Step 2: Clone Repository

```bash
# Clone the repository
git clone https://github.com/your-org/autostock.git
cd autostock
```

### Step 3: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit environment file
nano .env
```

**Required Environment Variables:**

```bash
# Database
DB_PASSWORD=your-secure-password-here
DB_NAME=autostock
DB_USER=autostock

# Backend
JWT_SECRET=your-jwt-secret-here-min-32-chars
BACKEND_PORT=8080
API_URL=http://localhost:8080/api/v1

# Frontend
FRONTEND_PORT=3000
VITE_API_URL=https://your-domain.com/api/v1

# Telegram (optional)
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=

# Application
APP_NAME=AutoStock
APP_ENV=production
LOG_LEVEL=info
```

**Generate Secure Secrets:**

```bash
# Generate random password
openssl rand -base64 32

# Generate JWT secret
openssl rand -base64 48
```

### Step 4: Build and Start

```bash
# Build images
docker-compose build

# Start services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### Step 5: Initialize Database

```bash
# Run migrations
docker-compose exec backend migrate -path ./migrations -database "postgres://autostock:${DB_PASSWORD}@postgres:5432/autostock?sslmode=disable" up

# Or if using the built-in migration command
docker-compose exec backend ./autostock migrate
```

### Step 6: Access Application

- **Frontend**: http://your-vps-ip:3000
- **Backend API**: http://your-vps-ip:8080
- **API Docs**: http://your-vps-ip:8080/swagger

### Step 7: Setup Domain (Optional)

**Install Nginx:**

```bash
apt install nginx -y
```

**Create Nginx Config:**

```bash
nano /etc/nginx/sites-available/autostock
```

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # Backend API
    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $http_host;
    }

    # Swagger docs
    location /swagger {
        proxy_pass http://localhost:8080/swagger;
    }
}
```

**Enable Site:**

```bash
ln -s /etc/nginx/sites-available/autostock /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

**Setup SSL with Let's Encrypt:**

```bash
apt install certbot python3-certbot-nginx -y
certbot --nginx -d your-domain.com
```

### Step 8: Firewall Configuration

```bash
# Enable firewall
ufw enable

# Allow SSH
ufw allow 22

# Allow HTTP/HTTPS
ufw allow 80
ufw allow 443

# Block direct access to app ports (optional, if using Nginx)
# ufw deny 3000
# ufw deny 8080
# ufw deny 5432

# Check status
ufw status
```

## Option 2: Split Deployment

Deploy frontend on Vercel (free tier) and backend + database on VPS.

### Frontend on Vercel

1. **Push code to GitHub**

2. **Import to Vercel**
   - Go to https://vercel.com
   - Import your repository
   - Select `frontend` directory as root
   - Add environment variables:
     - `VITE_API_URL`: `https://api.your-domain.com/api/v1`

3. **Deploy**
   - Vercel will automatically deploy on every push
   - Get your Vercel URL (e.g., `https://autostock.vercel.app`)

### Backend on VPS

Follow the same steps as Option 1, but:
- Remove frontend service from `docker-compose.yml`
- Update CORS settings in backend to allow Vercel domain:

```go
// backend/internal/middleware/cors.go
corsConfig := cors.Config{
    AllowOrigins: []string{
        "https://autostock.vercel.app",
        "http://localhost:3000",
    },
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
}
```

### Nginx Config for Split Deployment

```nginx
server {
    listen 80;
    server_name api.your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $http_host;
    }
}
```

## Option 3: Manual Deployment

For advanced users who want direct binary deployment without Docker.

### Prerequisites

```bash
# Install Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install PostgreSQL
apt install postgresql postgresql-contrib -y

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
apt install nodejs -y
```

### Build Backend

```bash
cd backend

# Build binary
go build -o autostock ./cmd/server

# Move to /opt
mv autostock /opt/autostock/
```

### Build Frontend

```bash
cd frontend

# Install dependencies
npm ci

# Build
npm run build

# Move to /opt
mv dist /opt/autostock/frontend
```

### Setup Systemd Services

**Backend Service:**

```bash
nano /etc/systemd/system/autostock-backend.service
```

```ini
[Unit]
Description=AutoStock Backend
After=network.target postgresql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/autostock
ExecStart=/opt/autostock/autostock
Restart=always
Environment=DATABASE_URL=postgres://autostock:password@localhost:5432/autostock
Environment=JWT_SECRET=your-secret
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
```

**Frontend Service (using serve):**

```bash
npm install -g serve
```

```bash
nano /etc/systemd/system/autostock-frontend.service
```

```ini
[Unit]
Description=AutoStock Frontend
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/autostock/frontend
ExecStart=/usr/bin/serve -s . -l 3000
Restart=always

[Install]
WantedBy=multi-user.target
```

**Enable Services:**

```bash
systemctl daemon-reload
systemctl enable autostock-backend
systemctl enable autostock-frontend
systemctl start autostock-backend
systemctl start autostock-frontend
```

## Docker Compose Configuration

### Production docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: autostock-postgres
    restart: always
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    ports:
      - "127.0.0.1:5432:5432"  # Only accessible from localhost
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: autostock-backend
    restart: always
    environment:
      DATABASE_URL: postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable
      JWT_SECRET: ${JWT_SECRET}
      PORT: ${BACKEND_PORT}
      TELEGRAM_BOT_TOKEN: ${TELEGRAM_BOT_TOKEN}
      TELEGRAM_CHAT_ID: ${TELEGRAM_CHAT_ID}
      LOG_LEVEL: ${LOG_LEVEL}
    ports:
      - "127.0.0.1:${BACKEND_PORT}:8080"
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
      args:
        VITE_API_URL: ${VITE_API_URL}
    container_name: autostock-frontend
    restart: always
    ports:
      - "${FRONTEND_PORT}:80"
    depends_on:
      - backend

volumes:
  postgres_data:
    driver: local
```

### Backend Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o autostock ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary
COPY --from=builder /app/autostock .
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run binary
CMD ["./autostock"]
```

### Frontend Dockerfile

```dockerfile
# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source code
COPY . .

# Build arguments
ARG VITE_API_URL
ENV VITE_API_URL=${VITE_API_URL}

# Build
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built files
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

### Frontend nginx.conf

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Health check endpoint
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
```

## Backup Strategy

### Automated Database Backups

**Create Backup Script:**

```bash
nano /opt/autostock/backup.sh
```

```bash
#!/bin/bash

# Configuration
BACKUP_DIR="/opt/autostock/backups"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
docker-compose exec -T postgres pg_dump -U autostock autostock | gzip > "$BACKUP_DIR/autostock_$DATE.sql.gz"

# Check if backup was successful
if [ $? -eq 0 ]; then
    echo "Backup successful: autostock_$DATE.sql.gz"
else
    echo "Backup failed!"
    exit 1
fi

# Remove old backups
find $BACKUP_DIR -name "autostock_*.sql.gz" -mtime +$RETENTION_DAYS -delete

echo "Old backups cleaned up (older than $RETENTION_DAYS days)"
```

**Make Executable:**

```bash
chmod +x /opt/autostock/backup.sh
```

**Setup Cron Job:**

```bash
crontab -e
```

```bash
# Daily backup at 2 AM
0 2 * * * /opt/autostock/backup.sh >> /var/log/autostock-backup.log 2>&1
```

### Restore from Backup

```bash
# Stop services
docker-compose stop backend

# Restore database
gunzip -c backups/autostock_20260704_020000.sql.gz | docker-compose exec -T postgres psql -U autostock autostock

# Start services
docker-compose start backend
```

## Monitoring

### Health Checks

**Backend Health Endpoint:**

```go
// backend/internal/handler/health.go
func (h *HealthHandler) Health(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "healthy",
        "timestamp": time.Now(),
    })
}
```

**Check Health:**

```bash
curl http://localhost:8080/health
```

### Log Management

**View Logs:**

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f postgres

# Last 100 lines
docker-compose logs --tail=100 backend
```

**Log Rotation:**

```bash
nano /etc/logrotate.d/autostock
```

```
/var/log/autostock/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 www-data www-data
}
```

### Resource Monitoring

**Install htop:**

```bash
apt install htop -y
htop
```

**Check Disk Usage:**

```bash
df -h
du -sh /opt/autostock/*
```

**Check Docker Resources:**

```bash
docker stats
```

## Security Best Practices

### 1. Keep System Updated

```bash
apt update && apt upgrade -y
```

### 2. Use Strong Passwords

- Database password: 32+ characters
- JWT secret: 48+ characters
- Admin password: 12+ characters with complexity

### 3. Enable Firewall

```bash
ufw enable
ufw allow 22    # SSH
ufw allow 80    # HTTP
ufw allow 443   # HTTPS
ufw status
```

### 4. Disable Root SSH Login

```bash
nano /etc/ssh/sshd_config
```

```
PermitRootLogin no
PasswordAuthentication no  # If using SSH keys
```

```bash
systemctl restart sshd
```

### 5. Use SSH Keys

```bash
# On your local machine
ssh-keygen -t ed25519 -C "your-email@example.com"

# Copy to server
ssh-copy-id user@your-vps-ip
```

### 6. Regular Backups

- Daily database backups
- Weekly full system backups
- Test restore procedure monthly

### 7. SSL/TLS

- Always use HTTPS in production
- Redirect HTTP to HTTPS
- Use strong TLS protocols (TLS 1.2+)

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker-compose logs backend

# Check container status
docker-compose ps

# Restart container
docker-compose restart backend

# Rebuild container
docker-compose up -d --build backend
```

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Test connection
docker-compose exec postgres psql -U autostock -c "SELECT 1"

# Check logs
docker-compose logs postgres
```

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in .env
```

### Out of Disk Space

```bash
# Check disk usage
df -h

# Clean Docker images
docker system prune -a

# Clean old backups
find /opt/autostock/backups -name "*.sql.gz" -mtime +7 -delete
```

## Performance Optimization

### PostgreSQL Tuning

```bash
# Edit PostgreSQL config
docker-compose exec postgres nano /var/lib/postgresql/data/postgresql.conf
```

```ini
# Memory settings (for 1GB RAM VPS)
shared_buffers = 256MB
effective_cache_size = 768MB
work_mem = 4MB
maintenance_work_mem = 64MB

# Connection settings
max_connections = 100
```

### Backend Optimization

```bash
# Enable connection pooling
# Already configured in backend code with pgxpool
```

### Frontend Optimization

- Enable gzip compression (already in nginx.conf)
- Use CDN for static assets (optional)
- Enable browser caching (already configured)

## Scaling

### Vertical Scaling

Upgrade VPS resources:
- More CPU cores
- More RAM
- Larger SSD

### Horizontal Scaling (Future)

For high traffic scenarios:
- Load balancer (Nginx, HAProxy)
- Multiple backend instances
- Database replication
- Redis for caching
- CDN for static assets

## Cost Optimization

### Cheapest VPS Options (as of 2026)

1. **DigitalOcean**: $4/month (1GB RAM, 1 CPU, 25GB SSD)
2. **Vultr**: $2.50/month (512MB RAM, 1 CPU, 10GB SSD)
3. **Linode**: $5/month (1GB RAM, 1 CPU, 25GB SSD)
4. **Hetzner**: €3.29/month (2GB RAM, 1 CPU, 20GB SSD)

### Split Deployment (Free Tier)

- **Frontend**: Vercel (free tier)
- **Backend + DB**: $5/month VPS
- **Total**: ~$5/month

### Cost Breakdown

| Component | Cost |
|-----------|------|
| VPS (1GB RAM) | $4-5/month |
| Domain (optional) | $10-15/year |
| SSL Certificate | Free (Let's Encrypt) |
| **Total** | **~$5/month** |

## Maintenance Checklist

### Daily
- [ ] Check logs for errors
- [ ] Verify backups completed
- [ ] Check disk space

### Weekly
- [ ] Review system updates
- [ ] Check database size
- [ ] Review error logs

### Monthly
- [ ] Apply security updates
- [ ] Test backup restore
- [ ] Review user access logs
- [ ] Clean old logs and backups

### Quarterly
- [ ] Full system audit
- [ ] Performance review
- [ ] Disaster recovery test
- [ ] Update documentation
