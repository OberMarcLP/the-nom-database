-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL, -- e.g., 'restaurants', 'users', 'suggestions'
    action VARCHAR(50) NOT NULL,   -- e.g., 'create', 'read', 'update', 'delete'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create role_permissions junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Create user_roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    PRIMARY KEY (user_id, role_id)
);

-- Create indexes
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);

-- Add trigger for roles table
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default roles
INSERT INTO roles (name, description) VALUES
    ('admin', 'Full system access - can manage users, settings, and all content'),
    ('moderator', 'Can review and moderate suggestions, manage content'),
    ('user', 'Standard user - can create restaurants, ratings, and suggestions'),
    ('guest', 'Limited access - can only view content')
ON CONFLICT (name) DO NOTHING;

-- Insert permissions
INSERT INTO permissions (name, description, resource, action) VALUES
    -- Restaurant permissions
    ('restaurants.create', 'Create new restaurants', 'restaurants', 'create'),
    ('restaurants.read', 'View restaurants', 'restaurants', 'read'),
    ('restaurants.update', 'Update restaurants', 'restaurants', 'update'),
    ('restaurants.delete', 'Delete restaurants', 'restaurants', 'delete'),
    ('restaurants.update.own', 'Update own restaurants', 'restaurants', 'update_own'),
    ('restaurants.delete.own', 'Delete own restaurants', 'restaurants', 'delete_own'),

    -- Rating permissions
    ('ratings.create', 'Create ratings', 'ratings', 'create'),
    ('ratings.read', 'View ratings', 'ratings', 'read'),
    ('ratings.update.own', 'Update own ratings', 'ratings', 'update_own'),
    ('ratings.delete.own', 'Delete own ratings', 'ratings', 'delete_own'),
    ('ratings.delete', 'Delete any rating', 'ratings', 'delete'),

    -- Suggestion permissions
    ('suggestions.create', 'Create suggestions', 'suggestions', 'create'),
    ('suggestions.read', 'View suggestions', 'suggestions', 'read'),
    ('suggestions.update', 'Update any suggestion', 'suggestions', 'update'),
    ('suggestions.delete', 'Delete any suggestion', 'suggestions', 'delete'),
    ('suggestions.approve', 'Approve suggestions', 'suggestions', 'approve'),
    ('suggestions.reject', 'Reject suggestions', 'suggestions', 'reject'),
    ('suggestions.convert', 'Convert suggestions to restaurants', 'suggestions', 'convert'),

    -- User permissions
    ('users.read', 'View user profiles', 'users', 'read'),
    ('users.update.own', 'Update own profile', 'users', 'update_own'),
    ('users.manage', 'Manage all users', 'users', 'manage'),
    ('users.assign_roles', 'Assign roles to users', 'users', 'assign_roles'),

    -- Category and Food Type permissions
    ('categories.create', 'Create categories', 'categories', 'create'),
    ('categories.update', 'Update categories', 'categories', 'update'),
    ('categories.delete', 'Delete categories', 'categories', 'delete'),
    ('food_types.create', 'Create food types', 'food_types', 'create'),
    ('food_types.update', 'Update food types', 'food_types', 'update'),
    ('food_types.delete', 'Delete food types', 'food_types', 'delete'),

    -- Photo permissions
    ('photos.upload', 'Upload photos', 'photos', 'create'),
    ('photos.delete.own', 'Delete own photos', 'photos', 'delete_own'),
    ('photos.delete', 'Delete any photo', 'photos', 'delete')
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to roles

-- Admin role - full access
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Moderator role - content management
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'moderator' AND p.name IN (
    'restaurants.read', 'restaurants.update', 'restaurants.delete',
    'ratings.read', 'ratings.delete',
    'suggestions.read', 'suggestions.update', 'suggestions.delete',
    'suggestions.approve', 'suggestions.reject', 'suggestions.convert',
    'users.read',
    'photos.delete'
)
ON CONFLICT DO NOTHING;

-- User role - standard permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user' AND p.name IN (
    'restaurants.create', 'restaurants.read', 'restaurants.update.own', 'restaurants.delete.own',
    'ratings.create', 'ratings.read', 'ratings.update.own', 'ratings.delete.own',
    'suggestions.create', 'suggestions.read',
    'users.read', 'users.update.own',
    'photos.upload', 'photos.delete.own'
)
ON CONFLICT DO NOTHING;

-- Guest role - read only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'guest' AND p.name IN (
    'restaurants.read',
    'ratings.read',
    'users.read'
)
ON CONFLICT DO NOTHING;

-- Migrate existing users to have appropriate roles
-- All existing users get 'user' role by default
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE r.name = 'user'
ON CONFLICT DO NOTHING;

-- Existing admin users also get 'admin' role
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE r.name = 'admin' AND u.is_admin = true
ON CONFLICT DO NOTHING;
