import { useState, useEffect, useRef } from 'react';
import { X, Camera } from 'lucide-react';
import { updateUserProfile, uploadAvatar } from '../services/api';
import { useAuth } from '../contexts/AuthContext';
import { Avatar } from './Avatar';

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
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen && user) {
      setUsername(user.username);
      setFullName(user.full_name || '');
      setEmail(user.email);
      setError('');
    }
  }, [isOpen, user]);

  const handleAvatarClick = () => {
    fileInputRef.current?.click();
  };

  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.type.startsWith('image/')) {
      setError('Please select an image file');
      return;
    }

    // Validate file size (5MB)
    if (file.size > 5 * 1024 * 1024) {
      setError('Image must be smaller than 5MB');
      return;
    }

    try {
      setUploading(true);
      setError('');
      const response = await uploadAvatar(file);

      // Update user with new avatar URL
      if (user) {
        const updatedUser = { ...user, avatar_url: response.avatar_url };
        updateUser(updatedUser);
        onSuccess();
      }
    } catch (err: any) {
      setError(err.message || 'Failed to upload avatar');
    } finally {
      setUploading(false);
    }
  };

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

  const isOIDCUser = user?.provider === 'oidc';

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

          {/* Avatar Upload Section */}
          <div className="flex flex-col items-center mb-6">
            <div className="relative group">
              <Avatar
                src={user?.avatar_url}
                alt={user?.username}
                size="xl"
                fallbackText={user?.full_name || user?.username}
                className="ring-4 ring-[var(--border)]"
              />
              <button
                type="button"
                onClick={handleAvatarClick}
                disabled={uploading}
                className="absolute inset-0 flex items-center justify-center bg-black/50 rounded-full opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
                title="Upload new avatar"
              >
                <Camera className="w-8 h-8 text-white" />
              </button>
            </div>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/jpeg,image/png,image/webp"
              onChange={handleAvatarUpload}
              className="hidden"
            />
            <p className="text-xs text-[var(--text-muted)] mt-3 text-center">
              {uploading ? 'Uploading...' : 'Click avatar to upload new picture (max 5MB)'}
            </p>
          </div>

          {isOIDCUser && (
            <div className="mb-5 p-4 border-2 border-[var(--info)] bg-[var(--info)]/10 rounded">
              <p className="text-sm text-[var(--info)] font-semibold">
                Username cannot be changed for OIDC accounts
              </p>
            </div>
          )}

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
                className={`admin-input ${isOIDCUser ? 'opacity-50 cursor-not-allowed' : ''}`}
                required
                disabled={isOIDCUser}
                title={isOIDCUser ? "Username cannot be changed for OIDC accounts" : ""}
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
