---
title: "Production Quick Start"
description: "Fast path to production deployment"
weight: 110
icon: "rocket_launch"
toc: true
---

# Production Deployment Quick Start

This guide provides the fastest path to getting The Nom Database running in production.

## Prerequisites

- Ubuntu 22.04+ server with Docker installed
- Domain name with DNS configured
- Google Maps API key

## 5-Minute Deployment

### 1. Clone Repository

```bash
cd /opt
sudo mkdir -p nomdb && sudo chown $USER:$USER nomdb
cd nomdb
git clone https://github.com/obermarclp/the-nom-database.git .
```

### 2. Configure Environment

```bash
# Copy production example
cp .env.production.example .env.production

# Generate secrets
echo "DB_PASSWORD=$(openssl rand -base64 32)" >> .env.production
echo "JWT_SECRET_KEY=$(openssl rand -base64 64)" >> .env.production

# Edit configuration
nano .env.production
```

**Required changes in `.env.production`:**
- Set `GOOGLE_MAPS_API_KEY=your_actual_api_key`
- Set your domain in `ALLOWED_ORIGINS` and `VITE_API_URL`
- Update `OIDC_*` variables if using OAuth (or set `AUTH_MODE=local`)

### 3. Setup SSL Certificates

**Option A: Let's Encrypt (Recommended)**

```bash
sudo apt install -y certbot
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com
sudo mkdir -p nginx/ssl
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/
sudo chown -R $USER:$USER nginx/ssl
```

**Option B: Self-Signed (Testing Only)**

```bash
./scripts/deploy.sh
# Will prompt to generate self-signed certificate
```

### 4. Update Domain in Nginx

```bash
# Replace yourdomain.com with your actual domain
sed -i 's/yourdomain\.com/your-actual-domain.com/g' nginx/nginx.conf
```

### 5. Deploy

```bash
./scripts/deploy.sh
```

This script will:
- Validate Docker installation
- Check environment configuration
- Verify SSL certificates
- Build Docker images
- Start all services
- Wait for health checks

### 6. Create Admin User

```bash
curl -X POST https://yourdomain.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@yourdomain.com",
    "username": "admin",
    "password": "ChangeThisPassword123!",
    "full_name": "Admin User"
  }'

# Make user an admin
docker compose -f docker-compose.prod.yml exec db psql -U nomdb -d nomdb \
  -c "UPDATE users SET is_admin = true WHERE email = 'admin@yourdomain.com';"
```

### 7. Setup Automated Backups

```bash
# Add daily backup at 2 AM
(crontab -l 2>/dev/null; echo "0 2 * * * cd /opt/nomdb && ./scripts/backup.sh") | crontab -
```

## Verify Deployment

```bash
# Check service health
docker compose -f docker-compose.prod.yml ps

# Test endpoints
curl https://yourdomain.com/health
curl https://yourdomain.com/api/health

# View logs
docker compose -f docker-compose.prod.yml logs -f
```

## Common Commands

```bash
# View logs
docker compose -f docker-compose.prod.yml logs -f

# Restart services
docker compose -f docker-compose.prod.yml restart

# Stop services
docker compose -f docker-compose.prod.yml down

# Update application
git pull origin main
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d

# Manual backup
./scripts/backup.sh

# Restore from backup
./scripts/restore.sh backups/backup_YYYYMMDD_HHMMSS.sql.gz
```

## Troubleshooting

### Services Won't Start

```bash
# Check logs
docker compose -f docker-compose.prod.yml logs

# Validate configuration
docker compose -f docker-compose.prod.yml config

# Check environment variables
cat .env.production
```

### SSL Certificate Issues

```bash
# Verify certificates exist
ls -la nginx/ssl/

# Test nginx config
docker compose -f docker-compose.prod.yml exec nginx nginx -t

# Regenerate certificates
sudo certbot renew --force-renewal
```

### Database Connection Issues

```bash
# Check database is running
docker compose -f docker-compose.prod.yml exec db pg_isready -U nomdb

# View database logs
docker compose -f docker-compose.prod.yml logs db

# Access database
docker compose -f docker-compose.prod.yml exec db psql -U nomdb nomdb
```

## Next Steps

- Review [Deployment Guide](../deployment/guide/) for detailed configuration
- Complete [Production Checklist](../deployment/production-checklist/)
- Set up monitoring and alerts
- Configure CDN (optional)
- Enable OIDC with Authentik (optional)
- Test disaster recovery procedures

## Security Reminders

- Change all default passwords
- Use strong JWT secret (64+ random bytes)
- Restrict Google Maps API key by domain
- Keep `.env.production` secure (never commit to git)
- Enable firewall (only ports 22, 80, 443)
- Set `AUTH_MODE=local` or `both` (never `none` in production)
- Keep Docker and system packages updated

## Support

For detailed documentation, see:
- [Deployment Guide](../deployment/guide/) - Full deployment guide
- [Authentication](../configuration/authentication/) - Auth setup with OIDC
- [Production Checklist](../deployment/production-checklist/) - Security checklist
