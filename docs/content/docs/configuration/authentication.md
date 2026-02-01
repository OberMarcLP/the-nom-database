---
title: "Authentication"
description: "Local, OIDC, and multi-mode auth setup"
weight: 310
icon: "lock"
toc: true
---

# Authentication Implementation Guide

## Overview

The Nom Database now includes a comprehensive authentication system with three modes of operation:

1. **No Auth (`AUTH_MODE=none`)** - For testing and development (current default)
2. **Local Auth (`AUTH_MODE=local`)** - Email/password authentication with JWT tokens
3. **OAuth (`AUTH_MODE=oauth`)** - Google OAuth authentication only
4. **Both (`AUTH_MODE=both`)** - Supports both local and OAuth (recommended for production)

## Quick Start

### For Development/Testing (Current Setup)

The application is currently configured with `AUTH_MODE=none` which disables authentication. All API endpoints are accessible without tokens.

```bash
# Start the backend
make backend

# Or with Docker
docker compose up
```

### Enabling Authentication

#### 1. Local Authentication (Email/Password)

**Step 1:** Update `.env`:
```bash
AUTH_MODE=local
JWT_SECRET_KEY=your_secure_secret_key_here  # Already set for you
```

**Step 2:** Run migrations to create user tables:
```bash
# The migration will run automatically on server start
# Or manually:
make migrate-up
```

**Step 3:** Restart the backend

**Step 4:** Register a user:
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "myusername",
    "password": "securepassword123",
    "full_name": "John Doe"
  }'
```

**Step 5:** Login:
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

You'll receive:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "random_secure_token",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "myusername",
    ...
  }
}
```

**Step 6:** Use the access token for protected endpoints:
```bash
curl -X POST http://localhost:8080/api/restaurants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -d '{
    "name": "New Restaurant",
    "address": "123 Main St"
  }'
```

#### 2. Google OAuth Authentication

**Step 1:** Set up Google OAuth:
- Go to [Google Cloud Console](https://console.cloud.google.com/)
- Create a new project or select existing
- Enable "Google+ API"
- Create OAuth 2.0 credentials (Web application)
- Add authorized redirect URI: `http://localhost:8080/api/auth/google/callback`

**Step 2:** Update `.env`:
```bash
AUTH_MODE=oauth  # or 'both' for local + OAuth
GOOGLE_OAUTH_CLIENT_ID=your_client_id.apps.googleusercontent.com
GOOGLE_OAUTH_CLIENT_SECRET=your_client_secret
GOOGLE_OAUTH_REDIRECT_URL=http://localhost:8080/api/auth/google/callback
```

**Step 3:** Restart backend

**Step 4:** Login flow:
1. Navigate to: `http://localhost:8080/api/auth/google`
2. You'll be redirected to Google login
3. After authorization, you'll be redirected back with tokens

#### 3. Both Local and OAuth (Recommended for Production)

```bash
AUTH_MODE=both
JWT_SECRET_KEY=your_secure_secret_key_here
GOOGLE_OAUTH_CLIENT_ID=your_client_id.apps.googleusercontent.com
GOOGLE_OAUTH_CLIENT_SECRET=your_client_secret
GOOGLE_OAUTH_REDIRECT_URL=https://yourdomain.com/api/auth/google/callback
```

## API Endpoints

### Public Endpoints (No Authentication Required)

#### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login with email/password
- `POST /api/auth/refresh` - Refresh access token
- `POST /api/auth/logout` - Logout (invalidate refresh token)
- `GET /api/auth/google` - Initiate Google OAuth
- `GET /api/auth/google/callback` - Google OAuth callback

#### Read-Only Access
- `GET /api/restaurants` - List restaurants
- `GET /api/restaurants/:id` - Get restaurant details
- `GET /api/categories` - List categories
- `GET /api/food-types` - List food types
- `GET /api/search` - Global search
- `GET /api/places/search` - Google Maps search (proxied)
- `GET /api/places/:placeId` - Google Maps place details

### Protected Endpoints (Authentication Required)

#### User Profile
- `GET /api/auth/me` - Get current user

#### Write Operations
- `POST /api/restaurants` - Create restaurant
- `PUT /api/restaurants/:id` - Update restaurant
- `DELETE /api/restaurants/:id` - Delete restaurant
- `POST /api/ratings` - Create rating
- `DELETE /api/ratings/:id` - Delete rating
- `POST /api/categories` - Create category
- `PUT /api/categories/:id` - Update category
- `DELETE /api/categories/:id` - Delete category
- Similar for food types, suggestions, and photos

## Frontend Integration

### Using Fetch API

```javascript
// Register
const registerResponse = await fetch('http://localhost:8080/api/auth/register', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'user@example.com',
    username: 'username',
    password: 'password123',
    full_name: 'John Doe'
  })
});
const { access_token, refresh_token, user } = await registerResponse.json();

// Store tokens
localStorage.setItem('access_token', access_token);
localStorage.setItem('refresh_token', refresh_token);

// Use token for protected requests
const createRestaurant = await fetch('http://localhost:8080/api/restaurants', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${localStorage.getItem('access_token')}`
  },
  body: JSON.stringify({ name: 'New Restaurant' })
});

// Handle token expiration and refresh
if (createRestaurant.status === 401) {
  // Refresh token
  const refreshResponse = await fetch('http://localhost:8080/api/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      refresh_token: localStorage.getItem('refresh_token')
    })
  });
  const { access_token: newToken } = await refreshResponse.json();
  localStorage.setItem('access_token', newToken);

  // Retry original request
  // ...
}
```

### Token Lifetimes

- **Access Token**: 15 minutes
- **Refresh Token**: 7 days

Access tokens should be used for API requests. When they expire, use the refresh token to get a new access token without requiring the user to log in again.

## Security Features

### Implemented

- **Password Hashing**: Argon2id with secure parameters
- **JWT Tokens**: HS256 algorithm with secure secret
- **Refresh Tokens**: Cryptographically secure random tokens
- **Rate Limiting**: 100 requests/minute per IP
- **Input Sanitization**: XSS protection
- **CORS**: Restricted to allowed origins
- **Security Headers**: XSS, clickjacking, MIME sniffing protection
- **Session Management**: IP and User-Agent tracking
- **Google Maps API Proxying**: API key not exposed to frontend

### Best Practices

1. **Never commit `.env` file** - It contains sensitive secrets
2. **Use strong JWT secrets** - At least 32 characters, generated with `openssl rand -base64 32`
3. **Enable HTTPS in production** - Never send tokens over HTTP
4. **Rotate secrets regularly** - Especially after security incidents
5. **Set appropriate CORS origins** - Only allow trusted domains

## Database Schema

### Users Table
```sql
users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  username VARCHAR(100) UNIQUE NOT NULL,
  password_hash VARCHAR(255),  -- NULL for OAuth users
  provider VARCHAR(50) DEFAULT 'local',  -- 'local', 'google'
  provider_id VARCHAR(255),  -- OAuth provider user ID
  full_name VARCHAR(255),
  avatar_url TEXT,
  is_active BOOLEAN DEFAULT true,
  is_admin BOOLEAN DEFAULT false,
  email_verified BOOLEAN DEFAULT false,
  last_login_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
```

### Sessions Table
```sql
sessions (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
  refresh_token VARCHAR(512) UNIQUE NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ip_address VARCHAR(45),
  user_agent TEXT
)
```

## Troubleshooting

### "JWT_SECRET_KEY environment variable is required"

**Solution**: Ensure `.env` has `JWT_SECRET_KEY` set when using `local` or `both` auth modes:
```bash
JWT_SECRET_KEY=$(openssl rand -base64 32)
```

### "OAuth not configured"

**Solution**: Add Google OAuth credentials to `.env`:
```bash
GOOGLE_OAUTH_CLIENT_ID=your_client_id
GOOGLE_OAUTH_CLIENT_SECRET=your_secret
```

### "Token has expired"

**Solution**: Use the refresh token endpoint to get a new access token:
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "your_refresh_token"}'
```

### "CORS error" when calling from frontend

**Solution**: Add your frontend URL to `.env`:
```bash
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

## Migration Path

### From No Auth to Auth

1. **Update `.env`**: Set `AUTH_MODE=local` or `AUTH_MODE=both`
2. **Run migrations**: Tables will be created automatically
3. **Create first user**: Use `/api/auth/register`
4. **Update frontend**: Add login/logout UI and token management
5. **Test thoroughly**: Ensure all features work with authentication

### Keeping No Auth Mode

If you want to keep testing without authentication:
```bash
AUTH_MODE=none
```

All endpoints remain accessible without tokens. A dummy user (ID: 1, email: test@example.com) is injected into all requests.

## Next Steps

- [ ] Implement frontend login/logout UI (see frontend integration section above)
- [ ] Add email verification workflow
- [ ] Implement password reset flow
- [ ] Add role-based access control (RBAC) for admin features
- [ ] Set up production OAuth redirect URLs
- [ ] Configure production-ready CORS origins
- [ ] Set up monitoring for failed login attempts
- [ ] Implement two-factor authentication (2FA)

## Resources

- [JWT.io](https://jwt.io/) - JWT debugger
- [Google OAuth Setup Guide](https://developers.google.com/identity/protocols/oauth2)
- [Argon2 Specification](https://github.com/P-H-C/phc-winner-argon2)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
