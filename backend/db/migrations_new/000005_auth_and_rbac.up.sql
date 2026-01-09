-- Users table for authentication and user management
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    provider VARCHAR(50) NOT NULL DEFAULT 'local',
    provider_id VARCHAR(255),
    full_name VARCHAR(255),
    avatar_url VARCHAR(500),
    bio TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_admin BOOLEAN NOT NULL DEFAULT false,
    password_must_change BOOLEAN NOT NULL DEFAULT false,
    email_verified BOOLEAN NOT NULL DEFAULT false,

    -- OIDC fields
    oidc_subject VARCHAR(255),
    oidc_provider VARCHAR(100),

    -- Timestamps
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_oidc ON users(oidc_subject, oidc_provider);
CREATE INDEX idx_users_active ON users(is_active);

-- Roles table for RBAC
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_roles_name ON roles(name);

-- Permissions table for RBAC
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_resource_action UNIQUE (resource, action)
);

CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_name ON permissions(name);

-- User-Role junction table (many-to-many)
CREATE TABLE user_roles (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    assigned_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

-- Role-Permission junction table (many-to-many)
CREATE TABLE role_permissions (
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);

-- Sessions table for refresh tokens
CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(500) NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address VARCHAR(45),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- API Keys table for programmatic access
CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_active ON api_keys(is_active);

-- Audit log table for tracking admin actions
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id INTEGER,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);

-- Add user_id foreign keys to existing tables
ALTER TABLE restaurants ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE ratings ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE restaurant_suggestions ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE menu_photos ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;

-- Add indexes for user_id foreign keys
CREATE INDEX idx_restaurants_user ON restaurants(user_id);
CREATE INDEX idx_ratings_user ON ratings(user_id);
CREATE INDEX idx_suggestions_user ON restaurant_suggestions(user_id);
CREATE INDEX idx_menu_photos_user ON menu_photos(user_id);

-- Insert default roles
INSERT INTO roles (name, description, is_system) VALUES
    ('admin', 'Full system access with all permissions', true),
    ('moderator', 'Can moderate content and manage suggestions', true),
    ('user', 'Standard registered user', true),
    ('guest', 'Limited read-only access', true);

-- Insert all permissions
INSERT INTO permissions (name, resource, action, description) VALUES
    -- Restaurant permissions
    ('restaurants.create', 'restaurants', 'create', 'Create new restaurants'),
    ('restaurants.read', 'restaurants', 'read', 'View restaurants'),
    ('restaurants.update', 'restaurants', 'update', 'Update restaurant details'),
    ('restaurants.delete', 'restaurants', 'delete', 'Delete restaurants'),
    ('restaurants.moderate', 'restaurants', 'moderate', 'Moderate any restaurant content'),

    -- Rating permissions
    ('ratings.create', 'ratings', 'create', 'Create new ratings'),
    ('ratings.read', 'ratings', 'read', 'View ratings'),
    ('ratings.update', 'ratings', 'update', 'Update own ratings'),
    ('ratings.delete', 'ratings', 'delete', 'Delete own ratings'),
    ('ratings.moderate', 'ratings', 'moderate', 'Moderate any rating content'),

    -- Category permissions
    ('categories.create', 'categories', 'create', 'Create new categories'),
    ('categories.read', 'categories', 'read', 'View categories'),
    ('categories.update', 'categories', 'update', 'Update categories'),
    ('categories.delete', 'categories', 'delete', 'Delete categories'),

    -- Food type permissions
    ('food_types.create', 'food_types', 'create', 'Create new food types'),
    ('food_types.read', 'food_types', 'read', 'View food types'),
    ('food_types.update', 'food_types', 'update', 'Update food types'),
    ('food_types.delete', 'food_types', 'delete', 'Delete food types'),

    -- Suggestion permissions
    ('suggestions.create', 'suggestions', 'create', 'Create restaurant suggestions'),
    ('suggestions.read', 'suggestions', 'read', 'View suggestions'),
    ('suggestions.update', 'suggestions', 'update', 'Update own suggestions'),
    ('suggestions.delete', 'suggestions', 'delete', 'Delete own suggestions'),
    ('suggestions.approve', 'suggestions', 'approve', 'Approve suggestions'),
    ('suggestions.reject', 'suggestions', 'reject', 'Reject suggestions'),
    ('suggestions.convert', 'suggestions', 'convert', 'Convert suggestions to restaurants'),

    -- Photo permissions
    ('photos.upload', 'photos', 'upload', 'Upload photos'),
    ('photos.delete', 'photos', 'delete', 'Delete own photos'),
    ('photos.moderate', 'photos', 'moderate', 'Moderate any photos'),

    -- User management permissions
    ('users.create', 'users', 'create', 'Create new users'),
    ('users.read', 'users', 'read', 'View user information'),
    ('users.update', 'users', 'update', 'Update user information'),
    ('users.delete', 'users', 'delete', 'Delete users'),
    ('users.manage_roles', 'users', 'manage_roles', 'Assign/remove user roles'),

    -- Role management permissions
    ('roles.create', 'roles', 'create', 'Create new roles'),
    ('roles.read', 'roles', 'read', 'View roles'),
    ('roles.update', 'roles', 'update', 'Update roles'),
    ('roles.delete', 'roles', 'delete', 'Delete roles'),
    ('roles.manage_permissions', 'roles', 'manage_permissions', 'Assign/remove role permissions'),

    -- System permissions
    ('system.settings', 'system', 'settings', 'Manage system settings'),
    ('system.audit', 'system', 'audit', 'View audit logs'),
    ('system.analytics', 'system', 'analytics', 'View system analytics');

-- Assign permissions to Admin role (all permissions)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin';

-- Assign permissions to Moderator role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
INNER JOIN permissions p ON p.name IN (
    'restaurants.create',
    'restaurants.read',
    'restaurants.update',
    'restaurants.moderate',
    'ratings.create',
    'ratings.read',
    'ratings.update',
    'ratings.moderate',
    'categories.read',
    'food_types.read',
    'suggestions.read',
    'suggestions.approve',
    'suggestions.reject',
    'suggestions.convert',
    'photos.upload',
    'photos.moderate',
    'users.read'
)
WHERE r.name = 'moderator';

-- Assign permissions to User role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
INNER JOIN permissions p ON p.name IN (
    'restaurants.create',
    'restaurants.read',
    'restaurants.update',
    'ratings.create',
    'ratings.read',
    'ratings.update',
    'ratings.delete',
    'categories.read',
    'food_types.read',
    'suggestions.create',
    'suggestions.read',
    'suggestions.update',
    'suggestions.delete',
    'photos.upload',
    'photos.delete'
)
WHERE r.name = 'user';

-- Assign permissions to Guest role (read-only)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
INNER JOIN permissions p ON p.name IN (
    'restaurants.read',
    'ratings.read',
    'categories.read',
    'food_types.read'
)
WHERE r.name = 'guest';
