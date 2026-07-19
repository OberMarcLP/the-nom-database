import { useEffect, useState } from 'react';
import { api, Role, Permission } from '../services/api';
import { Shield, Plus, Edit2, Trash2, X, CheckCircle } from 'lucide-react';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { useToast } from '../hooks/useToast';

interface RoleWithPermissions extends Role {
  permissions: Permission[];
}

export function AdminRoles() {
  const [roles, setRoles] = useState<RoleWithPermissions[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedRole, setSelectedRole] = useState<RoleWithPermissions | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showPermissionsModal, setShowPermissionsModal] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<{ roleId: number; roleName: string } | null>(null);
  const { showError } = useToast();

  useEffect(() => {
    loadRoles();
    loadPermissions();
  }, []);

  // Handle ESC key to close modals
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (confirmDelete) {
          setConfirmDelete(null);
        } else if (showPermissionsModal) {
          setShowPermissionsModal(false);
        } else if (showEditModal) {
          setShowEditModal(false);
        } else if (showCreateModal) {
          setShowCreateModal(false);
        }
      }
    };

    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [showCreateModal, showEditModal, showPermissionsModal, confirmDelete]);

  const loadRoles = async () => {
    try {
      setLoading(true);
      const response = await api.get<RoleWithPermissions[]>('/admin/roles');
      setRoles(response);
    } catch (error) {
      console.error('Failed to load roles:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadPermissions = async () => {
    try {
      const response = await api.get<Permission[]>('/admin/permissions');
      setPermissions(response);
    } catch (error) {
      console.error('Failed to load permissions:', error);
    }
  };

  const handleCreateRole = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);

    try {
      await api.post('/admin/roles', {
        name: formData.get('name'),
        description: formData.get('description') || null,
      });
      setShowCreateModal(false);
      loadRoles();
    } catch (error) {
      console.error('Failed to create role:', error);
      showError('Failed to create role');
    }
  };

  const handleUpdateRole = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selectedRole) return;

    const formData = new FormData(e.currentTarget);

    try {
      await api.put(`/admin/roles/${selectedRole.id}`, {
        name: formData.get('name'),
        description: formData.get('description') || null,
      });
      setShowEditModal(false);
      setSelectedRole(null);
      loadRoles();
    } catch (error) {
      console.error('Failed to update role:', error);
      showError('Failed to update role');
    }
  };

  const handleDeleteRole = async () => {
    if (!confirmDelete) return;

    try {
      await api.delete(`/admin/roles/${confirmDelete.roleId}`);
      loadRoles();
      setConfirmDelete(null);
    } catch (error) {
      console.error('Failed to delete role:', error);
      showError('Failed to delete role. System roles cannot be deleted.');
      setConfirmDelete(null);
    }
  };

  const handleAssignPermission = async (permissionId: number) => {
    if (!selectedRole) return;

    try {
      await api.post(`/admin/roles/${selectedRole.id}/permissions`, { permission_id: permissionId });
      const response = await api.get<RoleWithPermissions>(`/admin/roles/${selectedRole.id}`);
      setSelectedRole(response);

      // Update roles list without reloading
      setRoles(prevRoles =>
        prevRoles.map(role =>
          role.id === selectedRole.id ? response : role
        )
      );
    } catch (error) {
      console.error('Failed to assign permission:', error);
      showError('Failed to assign permission');
    }
  };

  const handleRemovePermission = async (permissionId: number) => {
    if (!selectedRole) return;

    try {
      await api.delete(`/admin/roles/${selectedRole.id}/permissions/${permissionId}`);
      const response = await api.get<RoleWithPermissions>(`/admin/roles/${selectedRole.id}`);
      setSelectedRole(response);

      // Update roles list without reloading
      setRoles(prevRoles =>
        prevRoles.map(role =>
          role.id === selectedRole.id ? response : role
        )
      );
    } catch (error) {
      console.error('Failed to remove permission:', error);
      showError('Failed to remove permission');
    }
  };

  const groupPermissionsByResource = (perms: Permission[]) => {
    const grouped: { [key: string]: Permission[] } = {};
    perms.forEach(perm => {
      if (!grouped[perm.resource]) {
        grouped[perm.resource] = [];
      }
      grouped[perm.resource].push(perm);
    });
    return grouped;
  };

  if (loading) {
    return (
      <div className="admin-loading">
        <div className="admin-spinner" />
      </div>
    );
  }

  return (
    <div>
      <div className="admin-page-header">
        <h1 className="admin-page-title">Role Management</h1>
        <p className="admin-page-description">
          Manage roles and their permissions
        </p>
        <button className="admin-btn" onClick={() => setShowCreateModal(true)}>
          <Plus size={16} />
          Create Role
        </button>
      </div>

      <div className="admin-card">
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <Shield size={20} />
            System Roles
          </h2>
        </div>

        <div className="admin-table-container">
          <table className="admin-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Permissions</th>
                <th>Type</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {roles.map(role => (
                <tr key={role.id}>
                  <td>
                    <span className="admin-badge admin-badge-primary">{role.name}</span>
                  </td>
                  <td style={{ color: 'var(--admin-text-muted)' }}>
                    {role.description || 'No description'}
                  </td>
                  <td>
                    <span className="admin-badge">{role.permissions?.length || 0} permissions</span>
                  </td>
                  <td>
                    {role.is_system ? (
                      <span className="admin-badge admin-badge-warning">System</span>
                    ) : (
                      <span className="admin-badge admin-badge-success">Custom</span>
                    )}
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: '8px' }}>
                      <button
                        className="admin-btn-icon"
                        onClick={() => {
                          setSelectedRole(role);
                          setShowPermissionsModal(true);
                        }}
                        title="Manage Permissions"
                      >
                        <Shield size={16} />
                      </button>
                      <button
                        className="admin-btn-icon"
                        onClick={() => {
                          setSelectedRole(role);
                          setShowEditModal(true);
                        }}
                        title="Edit Role"
                      >
                        <Edit2 size={16} />
                      </button>
                      {!role.is_system && (
                        <button
                          className="admin-btn-icon admin-btn-danger"
                          onClick={() => setConfirmDelete({ roleId: role.id, roleName: role.name })}
                          title="Delete Role"
                        >
                          <Trash2 size={16} />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create Role Modal */}
      {showCreateModal && (
        <div className="admin-modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="admin-modal" onClick={(e) => e.stopPropagation()}>
            <div className="admin-modal-header">
              <h2>Create New Role</h2>
              <button className="admin-modal-close" onClick={() => setShowCreateModal(false)}>
                <X size={20} />
              </button>
            </div>
            <form onSubmit={handleCreateRole}>
              <div className="admin-modal-body">
                <div className="admin-form-group">
                  <label className="admin-label">Role Name</label>
                  <input
                    type="text"
                    name="name"
                    className="admin-input"
                    placeholder="e.g., content-moderator"
                    required
                  />
                </div>
                <div className="admin-form-group">
                  <label className="admin-label">Description</label>
                  <textarea
                    name="description"
                    className="admin-input"
                    rows={3}
                    placeholder="Brief description of this role"
                  />
                </div>
              </div>
              <div className="admin-modal-footer">
                <button type="button" className="admin-btn admin-btn-secondary" onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="admin-btn">
                  Create Role
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Role Modal */}
      {showEditModal && selectedRole && (
        <div className="admin-modal-overlay" onClick={() => setShowEditModal(false)}>
          <div className="admin-modal" onClick={(e) => e.stopPropagation()}>
            <div className="admin-modal-header">
              <h2>Edit Role: {selectedRole.name}</h2>
              <button className="admin-modal-close" onClick={() => setShowEditModal(false)}>
                <X size={20} />
              </button>
            </div>
            <form onSubmit={handleUpdateRole}>
              <div className="admin-modal-body">
                <div className="admin-form-group">
                  <label className="admin-label">Role Name</label>
                  <input
                    type="text"
                    name="name"
                    className="admin-input"
                    defaultValue={selectedRole.name}
                    required
                  />
                </div>
                <div className="admin-form-group">
                  <label className="admin-label">Description</label>
                  <textarea
                    name="description"
                    className="admin-input"
                    rows={3}
                    defaultValue={selectedRole.description || ''}
                  />
                </div>
              </div>
              <div className="admin-modal-footer">
                <button type="button" className="admin-btn admin-btn-secondary" onClick={() => setShowEditModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="admin-btn">
                  Update Role
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Manage Permissions Modal */}
      {showPermissionsModal && selectedRole && (
        <div className="modal-overlay" onClick={() => setShowPermissionsModal(false)}>
          <div className="modal-glass admin-modal-lg" onClick={(e) => e.stopPropagation()}>
            <div className="admin-modal-header">
              <h2 className="admin-modal-title">Manage Permissions: {selectedRole.name}</h2>
              <button className="admin-modal-close" onClick={() => setShowPermissionsModal(false)}>
                <X size={20} />
              </button>
            </div>
            <div className="admin-modal-body">
              {Object.entries(groupPermissionsByResource(permissions)).map(([resource, perms]) => (
                <div key={resource} className="admin-permission-group">
                  <h3 className="admin-permission-resource">{resource}</h3>
                  <div className="admin-permission-grid">
                    {perms.map(perm => {
                      const hasPermission = selectedRole.permissions?.some(p => p.id === perm.id);
                      return (
                        <button
                          key={perm.id}
                          onClick={() => hasPermission ? handleRemovePermission(perm.id) : handleAssignPermission(perm.id)}
                          className={`admin-permission-toggle ${hasPermission ? 'admin-permission-active' : ''}`}
                        >
                          {hasPermission ? (
                            <CheckCircle size={14} />
                          ) : (
                            <div className="admin-permission-checkbox" />
                          )}
                          {perm.action}
                        </button>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
            <div className="admin-modal-footer">
              <button className="admin-btn" onClick={() => setShowPermissionsModal(false)}>
                Done
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Confirm Delete Dialog */}
      <ConfirmDialog
        isOpen={!!confirmDelete}
        onClose={() => setConfirmDelete(null)}
        onConfirm={handleDeleteRole}
        title="localhost:3000"
        message={`Are you sure you want to delete the role "${confirmDelete?.roleName}"?`}
        confirmText="OK"
        cancelText="Cancel"
        isDangerous={true}
      />
    </div>
  );
}
