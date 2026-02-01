---
title: "Deployment Summary"
description: "Production deployment package overview"
weight: 220
icon: "description"
toc: true
---

# Deployment Package Summary

This document summarizes the complete production deployment package for The Nom Database.

## What's Been Created

### Production Configuration Files

1. **docker-compose.prod.yml**
   - Production-ready Docker Compose configuration
   - All 4 services: PostgreSQL, Backend, Frontend, Nginx
   - Health checks for all containers
   - Restart policies (unless-stopped)
   - Log rotation (10MB max, 3 files)
   - Proper networking and volumes
   - Environment variable configuration

2. **.env.production.example**
   - Complete production environment template
   - All required variables documented
   - Secure defaults with clear instructions
   - Instructions for generating secrets
   - OIDC/Authentik configuration examples

3. **nginx/nginx.conf**
   - Reverse proxy configuration
   - HTTP to HTTPS redirect
   - SSL/TLS with modern ciphers (TLSv1.2, TLSv1.3)
   - Rate limiting (10 req/s API, 5 req/m auth)
   - Security headers (HSTS, X-Frame-Options, CSP-style headers)
   - Gzip compression
   - Static file caching (1 year)
   - Health check endpoint

### Deployment Scripts

1. **scripts/deploy.sh**
   - Automated deployment script
   - Environment validation
   - SSL certificate checks
   - Domain configuration helper
   - Interactive deployment workflow
   - Service health monitoring
   - Usage modes: full, check-only, build-only, start-only

2. **scripts/backup.sh**
   - Automated PostgreSQL backup
   - Gzip compression
   - 30-day retention policy
   - Optional S3 upload
   - Cron-friendly output
   - Backup size reporting

3. **scripts/restore.sh**
   - Database restoration from backup
   - Safety confirmations
   - Service orchestration (stop/restore/restart)
   - Automatic decompression
   - Connection cleanup

### Documentation

1. **Deployment Guide** ([guide](/docs/deployment/guide/))
   - Step-by-step deployment instructions
   - Server setup and prerequisites
   - SSL/TLS certificate setup (Let's Encrypt + self-signed)
   - DNS configuration
   - Environment configuration
   - Post-deployment steps
   - Monitoring and maintenance
   - Troubleshooting guide
   - Security checklist
   - Performance optimization
   - Disaster recovery procedures

2. **Quick Start** ([quickstart-production](/docs/getting-started/quickstart-production/))
   - 5-minute deployment guide
   - Minimal steps to production
   - Quick command reference
   - Common troubleshooting
   - Essential security reminders

3. **Production Checklist** ([production-checklist](/docs/deployment/production-checklist/))
   - Pre-deployment checklist
   - Configuration verification
   - Security hardening checklist
   - Post-deployment validation
   - Regular maintenance tasks
   - Emergency procedures
   - Compliance considerations

4. **scripts/README.md** (Script Documentation)
   - Detailed script usage
   - Examples and workflows
   - Best practices
   - Troubleshooting
   - Advanced usage patterns

5. **Authentication** ([authentication](/docs/configuration/authentication/))
   - OIDC/Authentik setup guide
   - Multi-mode authentication docs
   - API endpoint reference

## Deployment Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Internet                         │
└─────────────────┬───────────────────────────────────┘
                  │
         ┌────────▼────────┐
         │  Nginx (443)    │  ← SSL/TLS termination
         │  Rate Limiting  │  ← Security headers
         │  Compression    │  ← Static caching
         └────────┬────────┘
                  │
     ┌────────────┴────────────┐
     │                         │
┌────▼─────┐           ┌──────▼──────┐
│ Frontend │           │   Backend   │
│  (3000)  │           │    (8080)   │
│  React   │           │     Go      │
└──────────┘           └──────┬──────┘
                              │
                       ┌──────▼──────┐
                       │  PostgreSQL │
                       │    (5432)   │
                       │   Database  │
                       └─────────────┘
```

### Network Security

- Frontend: Only accessible via Nginx (localhost:3000)
- Backend: Only accessible via Nginx (localhost:8080)
- Database: Only accessible from backend (localhost:5432)
- Nginx: Public facing (ports 80, 443)

## Key Features

### Security

- **SSL/TLS**: Modern cipher suites, HSTS enabled
- **Rate Limiting**: API (10 req/s) and Auth (5 req/m)
- **Security Headers**: X-Frame-Options, X-Content-Type-Options, etc.
- **Network Isolation**: Backend and DB not publicly accessible
- **Secret Management**: Environment-based configuration
- **Authentication**: Multi-mode (none/local/oauth/both)
- **Password Hashing**: Argon2id with secure parameters

### Performance

- **Gzip Compression**: All text content
- **Static Caching**: 1-year cache for immutable assets
- **HTTP/2**: Enabled in Nginx
- **Connection Pooling**: PostgreSQL connection pool
- **Database Indexes**: Optimized queries

### Reliability

- **Health Checks**: All services monitored
- **Restart Policies**: Automatic recovery
- **Log Rotation**: Prevents disk filling
- **Automated Backups**: Daily PostgreSQL dumps
- **Backup Retention**: 30-day history
- **Disaster Recovery**: Tested restore procedures

### Maintainability

- **Infrastructure as Code**: All config in version control
- **Automated Scripts**: Deployment, backup, restore
- **Comprehensive Docs**: Multiple guide levels
- **Monitoring**: Service status and logs
- **Update Process**: Documented and scripted

## Quick Deployment Commands

```bash
# Initial setup (one time)
cp .env.production.example .env.production
nano .env.production  # Configure variables
sudo certbot certonly --standalone -d yourdomain.com
cp /etc/letsencrypt/live/yourdomain.com/*.pem nginx/ssl/
sed -i 's/yourdomain\.com/actual-domain.com/g' nginx/nginx.conf

# Deploy
./scripts/deploy.sh

# Setup automated backups
(crontab -l; echo "0 2 * * * cd /opt/nomdb && ./scripts/backup.sh") | crontab -

# Create admin user
curl -X POST https://yourdomain.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@yourdomain.com","username":"admin","password":"SecurePass123!","full_name":"Admin"}'

docker compose -f docker-compose.prod.yml exec db psql -U nomdb -d nomdb \
  -c "UPDATE users SET is_admin = true WHERE email = 'admin@yourdomain.com';"
```

## Environment Variables Reference

### Required (Always)

```bash
# Database
DB_PASSWORD=<generate with: openssl rand -base64 32>
DATABASE_URL=postgres://nomdb:${DB_PASSWORD}@db:5432/nomdb?sslmode=disable

# Google Maps
GOOGLE_MAPS_API_KEY=<your production API key>

# Domain
ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
VITE_API_URL=https://yourdomain.com
```

### Required (If using local/both auth mode)

```bash
AUTH_MODE=local  # or 'both' for local + OIDC
JWT_SECRET_KEY=<generate with: openssl rand -base64 64>
```

### Required (If using oauth/both auth mode)

```bash
AUTH_MODE=oauth  # or 'both' for local + OIDC
OIDC_ISSUER_URL=https://authentik.company/application/o/your-app/
OIDC_CLIENT_ID=<from Authentik>
OIDC_CLIENT_SECRET=<from Authentik>
OIDC_REDIRECT_URL=https://yourdomain.com/api/auth/oidc/callback
```

### Optional

```bash
# AWS S3 (for photo storage)
AWS_ACCESS_KEY_ID=<your key>
AWS_SECRET_ACCESS_KEY=<your secret>
AWS_REGION=us-east-1
S3_BUCKET_NAME=<your bucket>

# Backup to S3
S3_BACKUP_BUCKET=<backup bucket name>
```

## File Checklist

Before deploying, ensure these files exist:

```
✅ docker-compose.prod.yml
✅ .env.production (configured)
✅ nginx/nginx.conf (domain updated)
✅ nginx/ssl/fullchain.pem
✅ nginx/ssl/privkey.pem
✅ scripts/deploy.sh (executable)
✅ scripts/backup.sh (executable)
✅ scripts/restore.sh (executable)
```

## Post-Deployment Verification

```bash
# 1. Check all services are healthy
docker compose -f docker-compose.prod.yml ps

# 2. Test health endpoints
curl https://yourdomain.com/health        # Should return: healthy
curl https://yourdomain.com/api/health    # Should return: {"status":"ok"}

# 3. Verify SSL
curl -I https://yourdomain.com            # Should return 200 OK

# 4. Test HTTP redirect
curl -I http://yourdomain.com             # Should return 301/302 to HTTPS

# 5. Check authentication
curl -X POST https://yourdomain.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"test","password":"Test123!"}'

# 6. View logs
docker compose -f docker-compose.prod.yml logs -f

# 7. Test backup
./scripts/backup.sh
ls -lh backups/
```

## Monitoring Checklist

Daily:
- [ ] Check service health: `docker compose -f docker-compose.prod.yml ps`
- [ ] Review error logs: `docker compose -f docker-compose.prod.yml logs --tail=100`
- [ ] Verify backups ran: `ls -lt backups/ | head -5`

Weekly:
- [ ] Check disk space: `df -h`
- [ ] Review resource usage: `docker stats`
- [ ] Test restore: `./scripts/restore.sh <recent-backup>`

Monthly:
- [ ] Update Docker images: `docker compose pull && docker compose up -d`
- [ ] Review SSL expiry: `openssl x509 -in nginx/ssl/fullchain.pem -noout -dates`
- [ ] Security audit: Review access logs for suspicious activity

## Troubleshooting Quick Reference

**Services not starting:**
```bash
docker compose -f docker-compose.prod.yml logs
docker compose -f docker-compose.prod.yml restart
```

**502 Bad Gateway:**
```bash
docker compose -f docker-compose.prod.yml logs backend
docker compose -f docker-compose.prod.yml restart backend
```

**Database connection errors:**
```bash
docker compose -f docker-compose.prod.yml exec db pg_isready -U nomdb
docker compose -f docker-compose.prod.yml restart db backend
```

**SSL certificate expired:**
```bash
sudo certbot renew
sudo cp /etc/letsencrypt/live/yourdomain.com/*.pem nginx/ssl/
docker compose -f docker-compose.prod.yml restart nginx
```

## Security Hardening (Post-Deployment)

1. **Firewall**
   ```bash
   sudo ufw allow OpenSSH
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```

2. **Fail2ban** (optional)
   ```bash
   sudo apt install fail2ban
   sudo systemctl enable fail2ban
   ```

3. **Automatic Updates**
   ```bash
   sudo apt install unattended-upgrades
   sudo dpkg-reconfigure -plow unattended-upgrades
   ```

4. **Google Maps API Restrictions**
   - Restrict by domain in Google Cloud Console
   - Set spending limits
   - Enable only required APIs

5. **Regular Security Updates**
   ```bash
   # Update system
   sudo apt update && sudo apt upgrade -y

   # Update Docker images
   docker compose -f docker-compose.prod.yml pull
   docker compose -f docker-compose.prod.yml up -d
   ```

## Support Resources

- **Full Deployment Guide**: [Deployment Guide](/docs/deployment/guide/)
- **Quick Start**: [Quick Start Production](/docs/getting-started/quickstart-production/)
- **Security Checklist**: [Production Checklist](/docs/deployment/production-checklist/)
- **Authentication Setup**: [Authentication](/docs/configuration/authentication/)
- **Script Documentation**: See [scripts/README.md](https://github.com/obermarclp/the-nom-database/blob/main/scripts/README.md) in the repository
- **Project Overview**: [README](https://github.com/obermarclp/the-nom-database#readme)

## License

See LICENSE file in repository.

---

**Ready to deploy?** Start with [Quick Start Production](/docs/getting-started/quickstart-production/) for the fastest path to production!
