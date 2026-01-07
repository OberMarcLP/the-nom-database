import { useState } from 'react';
import { X } from 'lucide-react';
import { createList } from '../services/api';

interface ListFormModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

export function ListFormModal({ onClose, onSuccess }: ListFormModalProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isPublic, setIsPublic] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    try {
      setLoading(true);
      await createList({
        name: name.trim(),
        description: description.trim() || undefined,
        is_public: isPublic,
      });
      onSuccess();
    } catch (error) {
      console.error('Failed to create list:', error);
      alert('Failed to create list');
    } finally {
      setLoading(false);
    }
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
        <div className="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md p-8" onClick={(e) => e.stopPropagation()}>
          {/* Close button */}
          <button
            onClick={onClose}
            className="absolute top-4 right-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors"
          >
            <X className="w-6 h-6" />
          </button>

          {/* Header */}
          <div className="mb-6">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Create New List</h2>
            <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
              Organize restaurants you want to try or your favorites
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="listName" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                List Name *
              </label>
              <input
                id="listName"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Want to Try, Favorites, Date Night"
                className="appearance-none relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-700"
                required
                autoFocus
              />
            </div>

            <div>
              <label htmlFor="listDescription" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Description (optional)
              </label>
              <textarea
                id="listDescription"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Add a description for this list..."
                className="appearance-none relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-700 min-h-[80px]"
                rows={3}
              />
            </div>

            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="isPublic"
                checked={isPublic}
                onChange={(e) => setIsPublic(e.target.checked)}
                className="w-4 h-4 text-blue-600 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500 bg-white dark:bg-gray-700"
              />
              <label htmlFor="isPublic" className="text-sm text-gray-700 dark:text-gray-300">
                Make this list public (others can view it)
              </label>
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="submit"
                disabled={loading || !name.trim()}
                className="flex-1 flex justify-center py-2.5 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {loading ? 'Creating...' : 'Create List'}
              </button>
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2.5 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
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
