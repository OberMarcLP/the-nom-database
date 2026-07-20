import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { Avatar } from './Avatar';
import { Modal } from './Modal';

interface UserProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function UserProfileModal({ isOpen, onClose }: UserProfileModalProps) {
  const { user } = useAuth();
  const navigate = useNavigate();

  if (!user) return null;

  const handleChangePassword = () => {
    onClose();
    navigate('/change-password');
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="User Profile" size="lg">
      <div className="space-y-6">
        {/* Avatar Section */}
        <div className="flex justify-center">
          <Avatar
            src={user.avatar_url}
            alt={user.username}
            size="xl"
            fallbackText={user.full_name || user.username}
          />
        </div>

        {/* User Info */}
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
          <div>
            <label className="admin-label">Username</label>
            <div className="mt-1 text-sm text-(--text)">{user.username}</div>
          </div>

          <div>
            <label className="admin-label">Email</label>
            <div className="mt-1 text-sm text-(--text)">{user.email}</div>
          </div>

          {user.full_name && (
            <div>
              <label className="admin-label">Full Name</label>
              <div className="mt-1 text-sm text-(--text)">{user.full_name}</div>
            </div>
          )}

          <div>
            <label className="admin-label">Provider</label>
            <div className="mt-1 text-sm text-(--text) capitalize">{user.provider}</div>
          </div>

          <div>
            <label className="admin-label">Account Status</label>
            <div className="mt-1">
              <span className={`admin-badge ${
                user.is_active ? 'admin-badge-success' : 'admin-badge-danger'
              }`}>
                {user.is_active ? 'Active' : 'Inactive'}
              </span>
            </div>
          </div>

          {user.is_admin && (
            <div>
              <label className="admin-label">Role</label>
              <div className="mt-1">
                <span className="admin-badge admin-badge-info">
                  Administrator
                </span>
              </div>
            </div>
          )}

          <div>
            <label className="admin-label">Email Verified</label>
            <div className="mt-1">
              <span className={`admin-badge ${
                user.email_verified ? 'admin-badge-success' : 'admin-badge-warning'
              }`}>
                {user.email_verified ? 'Verified' : 'Not Verified'}
              </span>
            </div>
          </div>

          {user.last_login_at && (
            <div>
              <label className="admin-label">Last Login</label>
              <div className="mt-1 text-sm text-(--text)">
                {new Date(user.last_login_at).toLocaleString()}
              </div>
            </div>
          )}

          <div>
            <label className="admin-label">Member Since</label>
            <div className="mt-1 text-sm text-(--text)">
              {new Date(user.created_at).toLocaleDateString()}
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-3 pt-6 border-t border-(--border)">
          {user.provider === 'local' && (
            <button
              onClick={handleChangePassword}
              className="admin-btn-primary"
            >
              Change Password
            </button>
          )}
          <button
            onClick={onClose}
            className="admin-btn"
          >
            Close
          </button>
        </div>
      </div>
    </Modal>
  );
}
