import { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import { updateUserProfile } from '../services/api';
import { useAuth } from '../contexts/AuthContext';

interface EditProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export function EditProfileModal({ isOpen, onClose, onSuccess }: EditProfileModalProps) {
  const { user, updateUser } = useAuth();
  const [username, setUsername] = useState('');
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (isOpen && user) {
      setUsername(user.username);
      setFullName(user.full_name || '');
      setEmail(user.email);
      setError('');
    }
  }, [isOpen, user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!username.trim()) {
      setError('Username is required');
      return;
    }

    if (!email.trim()) {
      setError('Email is required');
      return;
    }

    try {
      setSaving(true);
      const updatedUser = await updateUserProfile({
        username: username.trim(),
        full_name: fullName.trim() || undefined,
        email: email.trim(),
      });
      updateUser(updatedUser);
      onSuccess();
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to update profile');
    } finally {
      setSaving(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-glass" onClick={(e) => e.stopPropagation()}>
        <div className="admin-modal-header">
          <h2 className="admin-modal-title">Edit Profile</h2>
          <button onClick={onClose} className="admin-modal-close">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="admin-modal-body">
          <p className="text-sm text-[var(--text-muted)] mb-6">
            Update your profile information
          </p>

          {error && (
            <div className="mb-5 p-4 border-2 border-[var(--danger)] bg-[var(--danger)]/10 rounded">
              <p className="text-sm text-[var(--danger)] font-semibold">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-5">
            <div className="admin-form-group">
              <label htmlFor="username" className="admin-label">
                Username *
              </label>
              <input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="admin-input"
                required
              />
            </div>

            <div className="admin-form-group">
              <label htmlFor="fullName" className="admin-label">
                Full Name
              </label>
              <input
                id="fullName"
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                className="admin-input"
              />
            </div>

            <div className="admin-form-group">
              <label htmlFor="email" className="admin-label">
                Email *
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="admin-input"
                required
              />
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="submit"
                disabled={saving}
                className="admin-btn-primary flex-1"
              >
                {saving ? 'Saving...' : 'Save Changes'}
              </button>
              <button
                type="button"
                onClick={onClose}
                className="admin-btn"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
