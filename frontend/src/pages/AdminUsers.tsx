import { useEffect, useState } from 'react';
import { api, User, Role } from '../services/api';
import {
  Users as UsersIcon,
  Plus,
  Search,
  Edit2,
  Trash2,
  Shield,
  Key,
  CheckCircle,
  XCircle,
} from 'lucide-react';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { Modal } from '../components/Modal';
import { useToast } from '../hooks/useToast';

interface UserListResponse {
  users: User[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export function AdminUsers() {
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [search, setSearch] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showRolesModal, setShowRolesModal] = useState(false);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<{ userId: number; username: string } | null>(null);
  const { showSuccess, showError } = useToast();

  useEffect(() => {
    loadUsers();
    loadRoles();
  }, [page, search]);

  const loadUsers = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams({
        page: page.toString(),
        limit: '20',
      });
      if (search) params.append('search', search);

      const response = await api.get<UserListResponse>(`/admin/users?${params}`);
      setUsers(response.users);
      setTotal(response.pagination.total);
      setTotalPages(response.pagination.totalPages);
    } catch (error) {
      console.error('Failed to load users:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadRoles = async () => {
    try {
      const response = await api.get<Role[]>('/admin/roles');
      setRoles(response);
    } catch (error) {
      console.error('Failed to load roles:', error);
    }
  };

  const handleCreateUser = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);

    try {
      await api.post('/admin/users', {
        username: formData.get('username'),
        email: formData.get('email'),
        password: formData.get('password'),
        full_name: formData.get('full_name'),
      });
      setShowCreateModal(false);
      loadUsers();
    } catch (error) {
      console.error('Failed to create user:', error);
      showError('Failed to create user');
    }
  };

  const handleUpdateUser = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selectedUser) return;

    const formData = new FormData(e.currentTarget);

    try {
      await api.put(`/admin/users/${selectedUser.id}`, {
        username: formData.get('username'),
        email: formData.get('email'),
        full_name: formData.get('full_name'),
        bio: formData.get('bio'),
        is_active: formData.get('is_active') === 'on',
        email_verified: formData.get('email_verified') === 'on',
      });
      setShowEditModal(false);
      setSelectedUser(null);
      loadUsers();
    } catch (error) {
      console.error('Failed to update user:', error);
      showError('Failed to update user');
    }
  };

  const handleDeleteUser = async () => {
    if (!confirmDelete) return;

    try {
      await api.delete(`/admin/users/${confirmDelete.userId}`);
      loadUsers();
      setConfirmDelete(null);
    } catch (error) {
      console.error('Failed to delete user:', error);
      showError('Failed to delete user');
      setConfirmDelete(null);
    }
  };

  const handleAssignRole = async (roleId: number) => {
    if (!selectedUser) return;

    try {
      await api.post(`/admin/users/${selectedUser.id}/roles`, { role_id: roleId });
      const response = await api.get<User>(`/admin/users/${selectedUser.id}`);
      setSelectedUser(response);

      // Update users list without reloading
      setUsers(prevUsers =>
        prevUsers.map(user =>
          user.id === selectedUser.id ? response : user
        )
      );
    } catch (error) {
      console.error('Failed to assign role:', error);
      showError('Failed to assign role');
    }
  };

  const handleRemoveRole = async (roleId: number) => {
    if (!selectedUser) return;

    try {
      await api.delete(`/admin/users/${selectedUser.id}/roles/${roleId}`);
      const response = await api.get<User>(`/admin/users/${selectedUser.id}`);
      setSelectedUser(response);

      // Update users list without reloading
      setUsers(prevUsers =>
        prevUsers.map(user =>
          user.id === selectedUser.id ? response : user
        )
      );
    } catch (error) {
      console.error('Failed to remove role:', error);
      showError('Failed to remove role');
    }
  };

  const handleResetPassword = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selectedUser) return;

    const formData = new FormData(e.currentTarget);

    try {
      await api.post(`/admin/users/${selectedUser.id}/reset-password`, {
        new_password: formData.get('new_password'),
      });
      setShowPasswordModal(false);
      setSelectedUser(null);
      showSuccess('Password reset successfully. User will be required to change password on next login.');
    } catch (error) {
      console.error('Failed to reset password:', error);
      showError('Failed to reset password');
    }
  };

  const openEditModal = (user: User) => {
    setSelectedUser(user);
    setShowEditModal(true);
  };

  const openRolesModal = (user: User) => {
    setSelectedUser(user);
    setShowRolesModal(true);
  };

  const openPasswordModal = (user: User) => {
    setSelectedUser(user);
    setShowPasswordModal(true);
  };

  return (
    <div>
      <div className="admin-page-header">
        <h1 className="admin-page-title">User Management</h1>
        <p className="admin-page-description">
          Manage user accounts, roles, and permissions
        </p>
        <div className="admin-page-actions">
          <div className="admin-search">
            <Search className="admin-search-icon" size={18} />
            <input
              type="text"
              placeholder="Search users..."
              className="admin-input admin-search-input"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
            />
          </div>
          <button
            className="admin-btn admin-btn-primary"
            onClick={() => setShowCreateModal(true)}
          >
            <Plus size={18} />
            Create User
          </button>
        </div>
      </div>

      {loading ? (
        <div className="admin-loading">
          <div className="admin-spinner" />
        </div>
      ) : users.length === 0 ? (
        <div className="admin-empty">
          <UsersIcon className="admin-empty-icon" size={48} />
          <p className="admin-empty-text">No users found</p>
        </div>
      ) : (
        <>
          <div className="admin-table-container">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>User</th>
                  <th>Email</th>
                  <th>Roles</th>
                  <th>Status</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id}>
                    <td>
                      <div className="admin-user-cell">
                        {user.avatar_url && (
                          <img
                            src={user.avatar_url}
                            alt={user.username}
                            className="admin-user-avatar"
                          />
                        )}
                        <div>
                          <div className="admin-user-name">{user.username}</div>
                          {user.full_name && (
                            <div className="admin-user-fullname">
                              {user.full_name}
                            </div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td>
                      <div className="admin-email-cell">
                        {user.email}
                        {user.email_verified && (
                          <CheckCircle size={14} className="text-accent" />
                        )}
                      </div>
                    </td>
                    <td>
                      <div className="admin-roles-cell">
                        {user.roles?.map((role) => (
                          <span key={role.id} className="admin-badge admin-badge-info">
                            {role.name}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td>
                      {user.is_active ? (
                        <span className="admin-badge admin-badge-success">Active</span>
                      ) : (
                        <span className="admin-badge admin-badge-danger">Inactive</span>
                      )}
                    </td>
                    <td className="admin-date-cell">
                      {new Date(user.created_at).toLocaleDateString()}
                    </td>
                    <td>
                      <div className="admin-actions-cell">
                        <button
                          className="admin-btn admin-btn-sm"
                          onClick={() => openEditModal(user)}
                          title="Edit user"
                        >
                          <Edit2 size={14} />
                        </button>
                        <button
                          className="admin-btn admin-btn-sm"
                          onClick={() => openRolesModal(user)}
                          title="Manage roles"
                        >
                          <Shield size={14} />
                        </button>
                        <button
                          className="admin-btn admin-btn-sm"
                          onClick={() => openPasswordModal(user)}
                          title="Reset password"
                        >
                          <Key size={14} />
                        </button>
                        <button
                          className="admin-btn admin-btn-sm admin-btn-danger"
                          onClick={() => setConfirmDelete({ userId: user.id, username: user.username })}
                          title="Delete user"
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="admin-pagination">
            <div className="admin-pagination-info">
              Showing {((page - 1) * 20) + 1} to {Math.min(page * 20, total)} of {total} users
            </div>
            <div className="admin-pagination-controls">
              <button
                className="admin-pagination-btn"
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                Previous
              </button>
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                const pageNum = i + 1;
                return (
                  <button
                    key={pageNum}
                    className={`admin-pagination-btn ${page === pageNum ? 'active' : ''}`}
                    onClick={() => setPage(pageNum)}
                  >
                    {pageNum}
                  </button>
                );
              })}
              <button
                className="admin-pagination-btn"
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
              >
                Next
              </button>
            </div>
          </div>
        </>
      )}

      {/* Create User Modal */}
      <Modal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        title="Create New User"
      >
        <form onSubmit={handleCreateUser}>
          <div className="admin-form-group">
            <label className="admin-label">Username *</label>
            <input type="text" name="username" className="admin-input" required />
          </div>
          <div className="admin-form-group">
            <label className="admin-label">Email *</label>
            <input type="email" name="email" className="admin-input" required />
          </div>
          <div className="admin-form-group">
            <label className="admin-label">Password *</label>
            <input type="password" name="password" className="admin-input" required minLength={8} />
          </div>
          <div className="admin-form-group">
            <label className="admin-label">Full Name</label>
            <input type="text" name="full_name" className="admin-input" />
          </div>
          <div className="admin-modal-footer">
            <button type="button" className="admin-btn" onClick={() => setShowCreateModal(false)}>
              Cancel
            </button>
            <button type="submit" className="admin-btn admin-btn-primary">
              Create User
            </button>
          </div>
        </form>
      </Modal>

      {/* Edit User Modal */}
      {selectedUser && (
        <Modal
          isOpen={showEditModal}
          onClose={() => setShowEditModal(false)}
          title={`Edit User: ${selectedUser.username}`}
        >
          <form onSubmit={handleUpdateUser}>
            <div className="admin-form-group">
              <label className="admin-label">Username</label>
              <input type="text" name="username" className="admin-input" defaultValue={selectedUser.username} required />
            </div>
            <div className="admin-form-group">
              <label className="admin-label">Email</label>
              <input type="email" name="email" className="admin-input" defaultValue={selectedUser.email} required />
            </div>
            <div className="admin-form-group">
              <label className="admin-label">Full Name</label>
              <input type="text" name="full_name" className="admin-input" defaultValue={selectedUser.full_name || ''} />
            </div>
            <div className="admin-form-group">
              <label className="admin-label">Bio</label>
              <textarea name="bio" className="admin-textarea" rows={3} defaultValue={selectedUser.bio || ''} />
            </div>
            <div className="admin-form-group">
              <label className="admin-checkbox-label">
                <input type="checkbox" name="is_active" defaultChecked={selectedUser.is_active} />
                <span className="admin-label">Account Active</span>
              </label>
            </div>
            <div className="admin-form-group">
              <label className="admin-checkbox-label">
                <input type="checkbox" name="email_verified" defaultChecked={selectedUser.email_verified} />
                <span className="admin-label">Email Verified</span>
              </label>
            </div>
            <div className="admin-modal-footer">
              <button type="button" className="admin-btn" onClick={() => setShowEditModal(false)}>
                Cancel
              </button>
              <button type="submit" className="admin-btn admin-btn-primary">
                Save Changes
              </button>
            </div>
          </form>
        </Modal>
      )}

      {/* Manage Roles Modal */}
      {selectedUser && (
        <Modal
          isOpen={showRolesModal}
          onClose={() => setShowRolesModal(false)}
          title={`Manage Roles: ${selectedUser.username}`}
        >
          <div className="admin-form-group">
            <label className="admin-label">Current Roles</label>
            <div className="admin-roles-cell" style={{ marginBottom: '20px' }}>
              {selectedUser.roles && selectedUser.roles.length > 0 ? (
                selectedUser.roles.map((role) => (
                  <div key={role.id} className="admin-badge admin-badge-info admin-badge-removable">
                    {role.name}
                    <button
                      onClick={() => handleRemoveRole(role.id)}
                      className="admin-badge-remove-btn"
                    >
                      <XCircle size={14} />
                    </button>
                  </div>
                ))
              ) : (
                <span className="text-muted">No roles assigned</span>
              )}
            </div>
          </div>
          <div className="admin-form-group">
            <label className="admin-label">Add Role</label>
            <div className="admin-grid-list">
              {roles.filter(role => !selectedUser.roles?.find(r => r.id === role.id)).map((role) => (
                <button
                  key={role.id}
                  className="admin-btn admin-btn-between"
                  onClick={() => handleAssignRole(role.id)}
                >
                  <span>{role.name}</span>
                  <Plus size={16} />
                </button>
              ))}
            </div>
          </div>
          <div className="admin-modal-footer">
            <button className="admin-btn" onClick={() => setShowRolesModal(false)}>
              Close
            </button>
          </div>
        </Modal>
      )}

      {/* Reset Password Modal */}
      {selectedUser && (
        <Modal
          isOpen={showPasswordModal}
          onClose={() => setShowPasswordModal(false)}
          title={`Reset Password: ${selectedUser.username}`}
        >
          <form onSubmit={handleResetPassword}>
            <p className="admin-info-text">
              User will be required to change their password on next login.
            </p>
            <div className="admin-form-group">
              <label className="admin-label">New Password</label>
              <input type="password" name="new_password" className="admin-input" required minLength={8} />
            </div>
            <div className="admin-modal-footer">
              <button type="button" className="admin-btn" onClick={() => setShowPasswordModal(false)}>
                Cancel
              </button>
              <button type="submit" className="admin-btn admin-btn-danger">
                Reset Password
              </button>
            </div>
          </form>
        </Modal>
      )}

      {/* Confirm Delete Dialog */}
      <ConfirmDialog
        isOpen={!!confirmDelete}
        onClose={() => setConfirmDelete(null)}
        onConfirm={handleDeleteUser}
        title="localhost:3000"
        message={`Are you sure you want to delete this user? This action cannot be undone.`}
        confirmText="OK"
        cancelText="Cancel"
        isDangerous={true}
      />
    </div>
  );
}
