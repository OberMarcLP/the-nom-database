import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { changePassword } from '../services/api';
import { useAuth } from '../contexts/AuthContext';

export default function ChangePasswordPage() {
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, updateUser } = useAuth();

  const isRequired = (location.state as any)?.required === true;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (newPassword !== confirmPassword) {
      setError('New passwords do not match');
      return;
    }

    if (newPassword.length < 8) {
      setError('New password must be at least 8 characters');
      return;
    }

    if (oldPassword === newPassword) {
      setError('New password must be different from old password');
      return;
    }

    setLoading(true);

    try {
      await changePassword({ old_password: oldPassword, new_password: newPassword });

      // Update user's password_must_change flag
      if (user) {
        updateUser({ ...user, password_must_change: false });
      }

      navigate('/', { replace: true });
    } catch (err: any) {
      setError(err.message || 'Failed to change password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full">
        <div className="card-glass">
          <div className="mb-6">
            <h2 className="admin-page-title text-center mb-4">
              Change Password
            </h2>
            {isRequired && (
              <div className="p-4 border-2 border-warning rounded-lg" style={{ background: 'var(--surface-hover)' }}>
                <p className="text-sm" style={{ color: 'var(--warning)' }}>
                  ⚠️ You must change your password before continuing
                </p>
              </div>
            )}
          </div>

          <form className="space-y-6" onSubmit={handleSubmit}>
            {error && (
              <div className="p-4 border-2 border-danger rounded-lg" style={{ background: 'var(--surface-hover)' }}>
                <p className="text-sm" style={{ color: 'var(--danger)' }}>{error}</p>
              </div>
            )}

            <div className="admin-form-group">
              <label htmlFor="oldPassword" className="admin-label">
                Current Password
              </label>
              <input
                id="oldPassword"
                name="oldPassword"
                type="password"
                autoComplete="current-password"
                required
                className="admin-input"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
              />
            </div>

            <div className="admin-form-group">
              <label htmlFor="newPassword" className="admin-label">
                New Password
              </label>
              <input
                id="newPassword"
                name="newPassword"
                type="password"
                autoComplete="new-password"
                required
                className="admin-input"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
              />
              <p className="mt-2 text-xs" style={{ color: 'var(--text-muted)' }}>
                Must be at least 8 characters
              </p>
            </div>

            <div className="admin-form-group">
              <label htmlFor="confirmPassword" className="admin-label">
                Confirm New Password
              </label>
              <input
                id="confirmPassword"
                name="confirmPassword"
                type="password"
                autoComplete="new-password"
                required
                className="admin-input"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
              />
            </div>

            <div className="flex gap-3 pt-4">
              {!isRequired && (
                <button
                  type="button"
                  onClick={() => navigate(-1)}
                  className="btn flex-1"
                >
                  Cancel
                </button>
              )}
              <button
                type="submit"
                disabled={loading}
                className="btn-primary flex-1 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? 'Changing...' : 'Change Password'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
