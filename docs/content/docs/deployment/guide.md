---
title: "Deployment Guide"
description: "Production deployment instructions"
weight: 210
icon: "rocket_launch"
toc: true
---

# Production Deployment Guide

This guide walks you through deploying The Nom Database to a production server.

## Prerequisites

### Server Requirements
- Ubuntu 22.04 LTS or later (recommended)
- Minimum 2GB RAM, 2 CPU cores
- 20GB+ storage
- Public IP address
- Domain name pointed to your server

### Required Software
- Docker Engine 24.0+
- Docker Compose v2.0+
- Git
- OpenSSL (for generating secrets)

## Step 1: Server Setup

### Install Docker and Docker Compose

```bash
# Update package index
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common

# Add Docker's official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Add Docker repository
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Add your user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify installation
docker --version
docker compose version
```

### Configure Firewall

```bash
# Allow SSH, HTTP, and HTTPS
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
sudo ufw status
```

## Step 2: Clone Repository

```bash
cd /opt
sudo mkdir -p nomdb
sudo chown $USER:$USER nomdb
cd nomdb
git clone https://github.com/obermarclp/the-nom-database.git .
```

## Step 3: Configure Environment

### Generate Secrets

```bash
# Generate JWT secret key
JWT_SECRET=$(openssl rand -base64 64)
echo "JWT_SECRET_KEY=$JWT_SECRET"

# Generate strong database password
DB_PASSWORD=$(openssl rand -base64 32)
echo "DB_PASSWORD=$DB_PASSWORD"
```

### Create Production Environment File

```bash
cp .env.production.example .env.production
```

Edit `.env.production` with your values:

```bash
nano .env.production
```

**Required Configuration:**

1. **Database**:
   - `DB_PASSWORD`: Use the generated password above
   - Keep `DB_USER=nomdb` and `DB_NAME=nomdb`

2. **JWT Secret**:
   - `JWT_SECRET_KEY`: Use the generated JWT secret above

3. **Google Maps API**:
   - `GOOGLE_MAPS_API_KEY`: Your production Google Maps API key
   - Enable Places API in Google Cloud Console

4. **OIDC/Authentik** (if using):
   - `OIDC_ISSUER_URL`: Your Authentik issuer URL
   - `OIDC_CLIENT_ID`: Client ID from Authentik
   - `OIDC_CLIENT_SECRET`: Client secret from Authentik
   - `OIDC_REDIRECT_URL`: https://yourdomain.com/api/auth/oidc/callback

5. **Domain Configuration**:
   - `ALLOWED_ORIGINS`: https://yourdomain.com,https://www.yourdomain.com
   - `VITE_API_URL`: https://yourdomain.com

6. **AWS S3** (optional):
   - Configure if using S3 for photo storage
   - Otherwise, leave blank for local storage

## Step 4: SSL/TLS Certificate Setup

### Option A: Let's Encrypt with Certbot (Recommended)

```bash
# Install Certbot
sudo apt install -y certbot

# Stop any services using port 80
docker compose -f docker-compose.prod.yml down

# Generate certificate (replace with your domain)
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com

# Certificates will be in /etc/letsencrypt/live/yourdomain.com/
# Copy to nginx directory
sudo mkdir -p nginx/ssl
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/
sudo chown -R $USER:$USER nginx/ssl
```

**Auto-renewal setup:**

```bash
# Add renewal cron job
echo "0 3 * * * root certbot renew --quiet --deploy-hook 'cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem /opt/nomdb/nginx/ssl/ && cp /etc/letsencrypt/live/yourdomain.com/privkey.pem /opt/nomdb/nginx/ssl/ && docker compose -f /opt/nomdb/docker-compose.prod.yml restart nginx'" | sudo tee -a /etc/crontab > /dev/null
```

### Option B: Self-Signed Certificate (Development Only)

```bash
sudo mkdir -p nginx/ssl
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/privkey.pem \
  -out nginx/ssl/fullchain.pem \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=yourdomain.com"
sudo chown -R $USER:$USER nginx/ssl
```

### Update Nginx Configuration

Edit `nginx/nginx.conf` and replace all instances of `yourdomain.com` with your actual domain:

```bash
sed -i 's/yourdomain\.com/your-actual-domain.com/g' nginx/nginx.conf
```

## Step 5: DNS Configuration

Point your domain to your server's IP address:

**A Records:**
- `yourdomain.com` → `your-server-ip`
- `www.yourdomain.com` → `your-server-ip`

Verify DNS propagation:
```bash
dig yourdomain.com +short
```

## Step 6: Build and Deploy

### Build Images

```bash
# Build with production environment
docker compose -f docker-compose.prod.yml build
```

### Start Services

```bash
# Start in detached mode
docker compose -f docker-compose.prod.yml up -d

# View logs
docker compose -f docker-compose.prod.yml logs -f

# Check service status
docker compose -f docker-compose.prod.yml ps
```

### Verify Deployment

```bash
# Check all services are healthy
docker compose -f docker-compose.prod.yml ps

# Test health endpoint
curl https://yourdomain.com/health

# Test API
curl https://yourdomain.com/api/health

# Check nginx logs
docker compose -f docker-compose.prod.yml logs nginx

# Check backend logs
docker compose -f docker-compose.prod.yml logs backend
```

## Step 7: Post-Deployment

### Create Admin User

If using local authentication:

```bash
curl -X POST https://yourdomain.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@yourdomain.com",
    "username": "admin",
    "password": "YourSecurePassword123!",
    "full_name": "Admin User"
  }'
```

Then manually set the user as admin in the database:

```bash
docker compose -f docker-compose.prod.yml exec db psql -U nomdb -d nomdb -c \
  "UPDATE users SET is_admin = true WHERE email = 'admin@yourdomain.com';"
```

### Setup Database Backups

Create a backup cron job:

```bash
# Create backup script (see scripts/backup.sh)
chmod +x scripts/backup.sh

# Add daily backup cron job at 2 AM
echo "0 2 * * * /opt/nomdb/scripts/backup.sh" | crontab -
```

## Step 8: Monitoring and Maintenance

### View Logs

```bash
# All services
docker compose -f docker-compose.prod.yml logs -f

# Specific service
docker compose -f docker-compose.prod.yml logs -f backend
docker compose -f docker-compose.prod.yml logs -f frontend
docker compose -f docker-compose.prod.yml logs -f nginx
docker compose -f docker-compose.prod.yml logs -f db
```

### Restart Services

```bash
# Restart all
docker compose -f docker-compose.prod.yml restart

# Restart specific service
docker compose -f docker-compose.prod.yml restart backend
```

### Update Application

```bash
# Pull latest changes
git pull origin main

# Rebuild and restart
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d

# Remove old images
docker image prune -f
```

### Database Management

```bash
# Access PostgreSQL
docker compose -f docker-compose.prod.yml exec db psql -U nomdb -d nomdb

# Backup database manually
docker compose -f docker-compose.prod.yml exec db pg_dump -U nomdb nomdb > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore database
docker compose -f docker-compose.prod.yml exec -T db psql -U nomdb nomdb < backup.sql
```

## Troubleshooting

### Services Not Starting

```bash
# Check logs for errors
docker compose -f docker-compose.prod.yml logs

# Check container status
docker compose -f docker-compose.prod.yml ps

# Restart services
docker compose -f docker-compose.prod.yml restart
```

### SSL Certificate Issues

```bash
# Verify certificate files exist
ls -la nginx/ssl/

# Check nginx configuration
docker compose -f docker-compose.prod.yml exec nginx nginx -t

# Check certificate validity
openssl x509 -in nginx/ssl/fullchain.pem -text -noout
```

### Database Connection Issues

```bash
# Check database is running
docker compose -f docker-compose.prod.yml exec db pg_isready -U nomdb

# Test connection
docker compose -f docker-compose.prod.yml exec db psql -U nomdb -d nomdb -c "SELECT 1;"

# Check backend environment
docker compose -f docker-compose.prod.yml exec backend env | grep DATABASE_URL
```

### 502 Bad Gateway

This usually means the backend is not running or not healthy:

```bash
# Check backend health
curl http://localhost:8080/api/health

# Check backend logs
docker compose -f docker-compose.prod.yml logs backend

# Restart backend
docker compose -f docker-compose.prod.yml restart backend
```

## Security Checklist

- [ ] Strong database password set
- [ ] JWT secret key generated with sufficient entropy
- [ ] SSL/TLS certificate installed and valid
- [ ] Firewall configured (only 80, 443, 22 open)
- [ ] Environment files (.env.production) not committed to git
- [ ] Google Maps API key restricted by domain/IP
- [ ] OIDC client secret kept secure
- [ ] Regular database backups configured
- [ ] Container restart policies enabled
- [ ] Log rotation configured
- [ ] Rate limiting enabled in nginx
- [ ] Security headers configured
- [ ] Database only accessible from localhost
- [ ] Backend only accessible through nginx
- [ ] Admin user password is strong

## Performance Optimization

### Enable Docker BuildKit

```bash
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
```

### Optimize Database

```bash
docker compose -f docker-compose.prod.yml exec db psql -U nomdb -d nomdb
```

```sql
-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_restaurants_created_at ON restaurants(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ratings_restaurant_id ON ratings(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token);

-- Vacuum and analyze
VACUUM ANALYZE;
```

### Monitor Resource Usage

```bash
# Container stats
docker stats

# Disk usage
docker system df

# Clean up
docker system prune -a --volumes
```

## Backup and Disaster Recovery

### Automated Backups

The backup script (`scripts/backup.sh`) runs daily and:
- Creates PostgreSQL dump
- Compresses backup
- Rotates old backups (keeps last 30 days)
- Optionally uploads to S3

### Manual Backup

```bash
./scripts/backup.sh
```

### Restore from Backup

```bash
# Stop services
docker compose -f docker-compose.prod.yml down

# Restore database
docker compose -f docker-compose.prod.yml up -d db
docker compose -f docker-compose.prod.yml exec -T db psql -U nomdb nomdb < backups/backup_20250115_020000.sql

# Restart all services
docker compose -f docker-compose.prod.yml up -d
```

## Scaling Considerations

### Horizontal Scaling

To run multiple backend instances behind nginx:

```yaml
# In docker-compose.prod.yml
backend:
  # ... existing config
  deploy:
    replicas: 3
```

### Database Connection Pooling

Already configured in backend with pgx connection pool.

### CDN Integration

Consider using CloudFlare or AWS CloudFront for:
- Static asset caching
- DDoS protection
- Geographic distribution

## Support and Resources

- Project Repository: https://github.com/obermarclp/the-nom-database
- Documentation: See [documentation](../) and [Authentication](../configuration/authentication/)
- Issue Tracker: GitHub Issues

## License

See LICENSE file in the repository.
