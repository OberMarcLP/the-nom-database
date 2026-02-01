---
title: "Database Migrations"
description: "Schema migrations with golang-migrate"
weight: 330
icon: "storage"
toc: true
---

# Database Migrations Guide

## Overview

The Nom Database uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema version control. Migrations are automatically run when the backend server starts, ensuring the database schema is always up-to-date.

## Migration Files

Migration files are located in `backend/db/migrations_new/` and follow this naming convention:

```
{version}_{description}.up.sql    # Applied when migrating forward
{version}_{description}.down.sql  # Applied when rolling back
```

Example:
```
000001_init_schema.up.sql
000001_init_schema.down.sql
```

## Current Migrations

1. **000001_init_schema** - Initial database schema with core tables
   - Creates: categories, food_types, restaurants, ratings, restaurant_food_types
   - Includes default seed data for categories and food types

2. **000002_suggestions_and_photos** - Restaurant suggestions and menu photos
   - Creates: restaurant_suggestions, menu_photos, suggestion_food_types
   - Adds suggestion workflow support

3. **000003_unique_constraints** - Prevent duplicate restaurants
   - Adds unique indexes on google_place_id and name+address combinations
   - Prevents duplicate restaurant entries

4. **000004_performance_indexes** - Query optimization
   - Adds indexes for common search patterns
   - Improves performance for filtering and sorting

## Automatic Migrations

Migrations run automatically when the backend server starts:

```bash
# Using Docker Compose
docker compose up

# Using make
make db    # Start database
make backend  # Migrations run on startup
```

## Manual Migration Commands

The migration CLI tool provides commands for manual migration management.

### Prerequisites

Set the DATABASE_URL environment variable:

```bash
export DATABASE_URL="postgres://nomdb:nomdb_secret@localhost:5432/nomdb?sslmode=disable"
```

Or use the `.env` file (recommended):

```bash
# .env file
DATABASE_URL=postgres://nomdb:nomdb_secret@localhost:5432/nomdb?sslmode=disable
```

### Available Commands

#### View Current Migration Version

```bash
make migrate-version
```

Or directly:

```bash
cd backend
go run cmd/migrate/main.go version
```

Output:
```
Current migration version: 4
```

#### Run Migrations

Apply all pending migrations:

```bash
make migrate-up
```

Or directly:

```bash
cd backend
go run cmd/migrate/main.go up
```

#### Rollback Migration

Roll back the last applied migration:

```bash
make migrate-down
```

Or directly:

```bash
cd backend
go run cmd/migrate/main.go down
```

**Warning**: This will undo the last migration using the `.down.sql` file. Make sure you have backups if working with production data.

#### Create New Migration

Create a new migration file pair:

```bash
make migrate-create NAME=add_user_profiles
```

Or directly:

```bash
cd backend
go run cmd/migrate/main.go create add_user_profiles
```

This creates:
- `backend/db/migrations_new/000005_add_user_profiles.up.sql`
- `backend/db/migrations_new/000005_add_user_profiles.down.sql`

#### Force Migration Version

If migrations are in a "dirty" state, you can force a specific version:

```bash
make migrate-force VERSION=4
```

Or directly:

```bash
cd backend
go run cmd/migrate/main.go force 4
```

**Warning**: This should only be used to fix migration state issues. It doesn't run any migrations, just sets the version number.

## Creating New Migrations

### Best Practices

1. **Atomic Changes**: Each migration should represent a single, atomic change
2. **Reversible**: Always write the corresponding `.down.sql` file
3. **Idempotent**: Use `IF EXISTS` and `IF NOT EXISTS` where appropriate
4. **Test Locally**: Test both up and down migrations before committing

### Example Migration

**000005_add_favorites.up.sql**:
```sql
-- Add user favorites table
CREATE TABLE IF NOT EXISTS user_favorites (
    id SERIAL PRIMARY KEY,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(restaurant_id, user_id)
);

-- Add index for faster queries
CREATE INDEX IF NOT EXISTS idx_favorites_user ON user_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_restaurant ON user_favorites(restaurant_id);
```

**000005_add_favorites.down.sql**:
```sql
-- Remove indexes
DROP INDEX IF EXISTS idx_favorites_restaurant;
DROP INDEX IF EXISTS idx_favorites_user;

-- Remove table
DROP TABLE IF EXISTS user_favorites;
```

### Writing Migration SQL

#### UP Migration (Forward)
- Create tables with `CREATE TABLE IF NOT EXISTS`
- Add columns with error handling
- Create indexes with `CREATE INDEX IF NOT EXISTS`
- Insert seed data with `ON CONFLICT DO NOTHING` if applicable

#### DOWN Migration (Rollback)
- Drop in reverse order (indexes first, then tables)
- Use `DROP ... IF EXISTS` to handle missing objects gracefully
- Remove seed data if necessary

## Migration State Management

### Schema Migrations Table

golang-migrate tracks migration state in the `schema_migrations` table:

```sql
-- View migration state
SELECT * FROM schema_migrations;
```

Output:
```
 version | dirty
---------+-------
       4 | f
```

- **version**: Current migration version
- **dirty**: `true` if migration failed mid-execution, `false` if clean

### Handling Dirty State

If migrations fail mid-execution, the database enters a "dirty" state:

```
Failed to run migrations: Dirty database version 1. Fix and force version.
```

To fix:

1. **Inspect the database** - Check which changes were applied
2. **Manually fix the schema** - Complete or rollback the partial migration
3. **Force the version** - Set the correct version number

```bash
# If migration completed, force to current version
make migrate-force VERSION=4

# If migration needs rollback, force to previous version
make migrate-force VERSION=3
```

## Development Workflow

### Starting Fresh

To reset the database completely:

```bash
# Stop containers and remove volumes
docker compose down -v

# Start with clean database (migrations run automatically)
docker compose up
```

### Adding a New Feature

1. Create migration files:
```bash
make migrate-create NAME=add_feature_name
```

2. Edit the `.up.sql` file with your schema changes

3. Edit the `.down.sql` file with rollback logic

4. Test locally:
```bash
# Apply migration
make migrate-up

# Test the feature
# ...

# Test rollback
make migrate-down

# Re-apply
make migrate-up
```

5. Commit both files:
```bash
git add backend/db/migrations_new/00000X_add_feature_name.*
git commit -m "feat: add feature_name schema"
```

## Production Deployment

### Pre-Deployment Checklist

- [ ] All migrations tested locally
- [ ] Both up and down migrations verified
- [ ] Database backup created
- [ ] Migrations reviewed by team
- [ ] Migration order is correct

### Deployment Process

1. **Backup the database**:
```bash
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql
```

2. **Deploy new code** with migrations

3. **Verify migration success**:
```bash
make migrate-version
```

4. **Monitor application logs** for errors

### Rollback Strategy

If deployment fails:

1. **Rollback code** to previous version

2. **Rollback migration** if necessary:
```bash
make migrate-down
```

3. **Restore from backup** if needed:
```bash
psql $DATABASE_URL < backup_YYYYMMDD_HHMMSS.sql
```

## Troubleshooting

### Error: "Dirty database version X"

**Cause**: Migration failed mid-execution

**Solution**:
```bash
# Check database state
docker compose exec db psql -U nomdb -d nomdb -c "\dt"

# Force to correct version after manual fix
make migrate-force VERSION=X
```

### Error: "relation already exists"

**Cause**: Migration files conflict with existing schema

**Solution**:
- Use `IF NOT EXISTS` in CREATE statements
- Use `IF EXISTS` in DROP statements
- Check for old migration scripts in `db/migrations/` (should be removed)

### Error: "no such file or directory"

**Cause**: Migration files not found

**Solution**:
- Check `backend/db/migrations_new/` directory exists
- Verify Dockerfile copies migrations correctly
- Ensure files are named correctly (XXXXXX_name.up.sql / .down.sql)

### Migrations not running in Docker

**Cause**: Migration files not copied to container

**Solution**:
Check `backend/Dockerfile` includes:
```dockerfile
COPY --from=builder /app/db/migrations_new ./db/migrations_new
```

## References

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Migration Best Practices](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md)
