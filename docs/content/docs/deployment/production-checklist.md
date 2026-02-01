---
title: "Production Checklist"
description: "Pre- and post-deployment checklist"
weight: 230
icon: "checklist"
toc: true
---

# Production Deployment Checklist

Use this checklist to ensure your production deployment is secure and properly configured.

## Pre-Deployment

### Server Setup
- [ ] Server meets minimum requirements (2GB RAM, 2 CPU, 20GB storage)
- [ ] Ubuntu 22.04 LTS or later installed
- [ ] Docker Engine 24.0+ installed
- [ ] Docker Compose v2.0+ installed
- [ ] Firewall configured (ports 22, 80, 443 open)
- [ ] SSH key-based authentication enabled
- [ ] Root login disabled
- [ ] Automatic security updates enabled

### Domain & DNS
- [ ] Domain name registered
- [ ] DNS A records configured for domain and www subdomain
- [ ] DNS propagated (verified with `dig yourdomain.com`)
- [ ] Domain pointed to correct server IP

### SSL/TLS Certificates
- [ ] Certbot installed (for Let's Encrypt)
- [ ] SSL certificates generated
- [ ] Certificates copied to `nginx/ssl/` directory
- [ ] Certificate permissions set correctly
- [ ] Auto-renewal configured (cron job)
- [ ] Certificate expiry monitoring set up

## Configuration

### Environment Variables
- [ ] Copied `.env.production.example` to `.env.production`
- [ ] Generated strong database password (`openssl rand -base64 32`)
- [ ] Generated JWT secret key (`openssl rand -base64 64`)
- [ ] Configured Google Maps API key (with domain restrictions)
- [ ] Set correct domain in `ALLOWED_ORIGINS`
- [ ] Set correct domain in `VITE_API_URL`
- [ ] Configured OIDC settings (if using Authentik/OAuth)
- [ ] Set `AUTH_MODE` appropriately (none/local/oauth/both)
- [ ] AWS S3 configured (if using for photo storage)
- [ ] Verified `.env.production` is in `.gitignore`

### Nginx Configuration
- [ ] Updated `nginx/nginx.conf` with actual domain name
- [ ] SSL certificate paths are correct
- [ ] Rate limiting configured appropriately
- [ ] Security headers enabled
- [ ] CORS origins match your frontend domain
- [ ] Gzip compression enabled
- [ ] Static file caching configured

### Database
- [ ] Database password is strong and unique
- [ ] Database port only accessible from localhost
- [ ] Database backups configured
- [ ] Connection pool size appropriate for load

## Security

### Secrets Management
- [ ] All default passwords changed
- [ ] JWT secret is randomly generated (min 32 bytes)
- [ ] Database password is strong (min 16 characters)
- [ ] OIDC client secret is secure
- [ ] `.env.production` has restricted permissions (600)
- [ ] No secrets committed to git repository

### API Keys
- [ ] Google Maps API key restricted by domain
- [ ] Google Maps API key has spending limits
- [ ] AWS credentials use IAM with minimal permissions
- [ ] AWS S3 bucket has CORS properly configured

### Authentication
- [ ] AUTH_MODE set appropriately (not 'none' in production)
- [ ] JWT token expiration configured (15 minutes recommended)
- [ ] Refresh token expiration configured (7 days recommended)
- [ ] Password hashing uses Argon2id (already configured)
- [ ] OIDC issuer URL is HTTPS
- [ ] OIDC redirect URL matches exactly

### Network Security
- [ ] Firewall enabled (ufw)
- [ ] Only necessary ports open (22, 80, 443)
- [ ] Database not exposed to internet
- [ ] Backend not directly exposed to internet
- [ ] Nginx rate limiting enabled
- [ ] Fail2ban installed and configured (optional)

### Container Security
- [ ] Docker daemon socket not exposed
- [ ] Containers run as non-root users where possible
- [ ] Container restart policies set
- [ ] Resource limits configured (CPU, memory)
- [ ] Log rotation enabled

## Deployment

### Initial Deployment
- [ ] Code repository cloned to `/opt/nomdb`
- [ ] Dependencies installed
- [ ] Environment file validated (`scripts/deploy.sh --check-only`)
- [ ] Docker images built successfully
- [ ] Services started successfully
- [ ] All containers are healthy
- [ ] Health endpoints responding (https://yourdomain.com/health)

### Database Initialization
- [ ] Database migrations applied
- [ ] Database schema verified
- [ ] Indexes created for performance
- [ ] Admin user created
- [ ] Admin user privileges verified

### Application Testing
- [ ] Frontend accessible at https://yourdomain.com
- [ ] API accessible at https://yourdomain.com/api/health
- [ ] SSL certificate valid (no browser warnings)
- [ ] HTTP redirects to HTTPS
- [ ] Authentication flow works (login/register)
- [ ] OIDC flow works (if configured)
- [ ] Restaurant search works
- [ ] Photo upload works
- [ ] Google Maps integration works
- [ ] Dark/light theme toggle works

### Monitoring
- [ ] Application logs accessible
- [ ] Error logs configured
- [ ] Log rotation configured
- [ ] Disk space monitoring set up
- [ ] Uptime monitoring configured
- [ ] SSL certificate expiry monitoring
- [ ] Database backup monitoring

## Post-Deployment

### Backups
- [ ] Backup script tested (`scripts/backup.sh`)
- [ ] Automated daily backups configured (cron)
- [ ] Backup retention policy set (30 days default)
- [ ] Restore procedure tested (`scripts/restore.sh`)
- [ ] Backups stored off-site (S3 or equivalent)
- [ ] Backup integrity verified

### Performance
- [ ] Database indexes created
- [ ] Static assets cached properly
- [ ] Gzip compression verified
- [ ] CDN configured (optional)
- [ ] Image optimization in place
- [ ] Query performance acceptable

### Documentation
- [ ] Admin credentials documented (securely)
- [ ] Deployment procedures documented
- [ ] Backup/restore procedures documented
- [ ] Emergency contact information documented
- [ ] Runbook created for common issues

### Maintenance
- [ ] Update procedure documented
- [ ] Rollback procedure documented
- [ ] Monitoring alerts configured
- [ ] On-call rotation set up (if applicable)
- [ ] Incident response plan documented

## Regular Maintenance Tasks

### Daily
- [ ] Check application health
- [ ] Review error logs
- [ ] Verify backups completed successfully

### Weekly
- [ ] Review resource usage (CPU, memory, disk)
- [ ] Check for failed login attempts
- [ ] Review rate limiting logs
- [ ] Test restore from backup

### Monthly
- [ ] Update Docker images
- [ ] Update dependencies
- [ ] Review SSL certificate expiry
- [ ] Review and rotate logs
- [ ] Security audit
- [ ] Performance review

### Quarterly
- [ ] Disaster recovery drill
- [ ] Review and update documentation
- [ ] Capacity planning review
- [ ] Security penetration testing (if applicable)

## Emergency Procedures

### Service Down
```bash
# Check service status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs -f

# Restart specific service
docker compose -f docker-compose.prod.yml restart backend

# Restart all services
docker compose -f docker-compose.prod.yml restart
```

### Database Issues
```bash
# Access database
docker compose -f docker-compose.prod.yml exec db psql -U nomdb nomdb

# Check connections
docker compose -f docker-compose.prod.yml exec db psql -U nomdb -c \
  "SELECT count(*) FROM pg_stat_activity WHERE datname = 'nomdb';"

# Restore from backup
./scripts/restore.sh backups/backup_YYYYMMDD_HHMMSS.sql.gz
```

### SSL Certificate Expired
```bash
# Renew certificate
sudo certbot renew

# Copy new certificates
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/

# Restart nginx
docker compose -f docker-compose.prod.yml restart nginx
```

### Out of Disk Space
```bash
# Check disk usage
df -h

# Clean up Docker
docker system prune -a --volumes

# Remove old backups
find backups/ -name "backup_*.sql.gz" -mtime +30 -delete

# Check log sizes
du -sh /var/lib/docker/containers/*/*.log
```

## Compliance & Legal

- [ ] Privacy policy published (if collecting user data)
- [ ] Terms of service published
- [ ] GDPR compliance verified (if serving EU users)
- [ ] Cookie consent implemented (if required)
- [ ] Data retention policy documented
- [ ] User data deletion procedure documented

## Sign-Off

**Deployed by:** ____________________

**Date:** ____________________

**Version:** ____________________

**Notes:**
___________________________________________
___________________________________________
___________________________________________
