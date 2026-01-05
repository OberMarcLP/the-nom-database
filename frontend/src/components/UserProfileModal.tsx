import { useNavigate } from 'react-router-dom';
import { X } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import { useEscapeKey } from '../hooks/useEscapeKey';
import { Avatar } from './Avatar';

interface UserProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function UserProfileModal({ isOpen, onClose }: UserProfileModalProps) {
  useEscapeKey(onClose, isOpen);
  const { user } = useAuth();
  const navigate = useNavigate();

  if (!isOpen || !user) return null;

  const handleChangePassword = () => {
    onClose();
    navigate('/change-password');
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-2xl p-8">
          {/* Close button */}
          <button
            onClick={onClose}
            className="absolute top-4 right-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors"
          >
            <X className="w-6 h-6" />
          </button>

          {/* Header */}
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">User Profile</h2>

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
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Username</label>
                <div className="mt-1 text-sm text-gray-900 dark:text-white">{user.username}</div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Email</label>
                <div className="mt-1 text-sm text-gray-900 dark:text-white">{user.email}</div>
              </div>

              {user.full_name && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Full Name</label>
                  <div className="mt-1 text-sm text-gray-900 dark:text-white">{user.full_name}</div>
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Provider</label>
                <div className="mt-1 text-sm text-gray-900 dark:text-white capitalize">{user.provider}</div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Account Status</label>
                <div className="mt-1">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    user.is_active
                      ? 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-200'
                      : 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-200'
                  }`}>
                    {user.is_active ? 'Active' : 'Inactive'}
                  </span>
                </div>
              </div>

              {user.is_admin && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Role</label>
                  <div className="mt-1">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900/20 dark:text-purple-200">
                      Administrator
                    </span>
                  </div>
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Email Verified</label>
                <div className="mt-1">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    user.email_verified
                      ? 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-200'
                      : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-200'
                  }`}>
                    {user.email_verified ? 'Verified' : 'Not Verified'}
                  </span>
                </div>
              </div>

              {user.last_login_at && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Last Login</label>
                  <div className="mt-1 text-sm text-gray-900 dark:text-white">
                    {new Date(user.last_login_at).toLocaleString()}
                  </div>
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Member Since</label>
                <div className="mt-1 text-sm text-gray-900 dark:text-white">
                  {new Date(user.created_at).toLocaleDateString()}
                </div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex gap-3 pt-6 border-t border-gray-200 dark:border-gray-700">
              {user.provider === 'local' && (
                <button
                  onClick={handleChangePassword}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
                >
                  Change Password
                </button>
              )}
              <button
                onClick={onClose}
                className="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
