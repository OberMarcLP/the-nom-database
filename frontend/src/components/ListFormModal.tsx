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
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-glass" onClick={(e) => e.stopPropagation()}>
        <div className="admin-modal-header">
          <h2 className="admin-modal-title">Create New List</h2>
          <button onClick={onClose} className="admin-modal-close">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="admin-modal-body">
          <p className="text-sm text-[var(--text-muted)] mb-6">
            Organize restaurants you want to try or your favorites
          </p>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div className="admin-form-group">
              <label htmlFor="listName" className="admin-label">
                List Name *
              </label>
              <input
                id="listName"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Want to Try, Favorites, Date Night"
                className="admin-input"
                required
                autoFocus
              />
            </div>

            <div className="admin-form-group">
              <label htmlFor="listDescription" className="admin-label">
                Description (optional)
              </label>
              <textarea
                id="listDescription"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Add a description for this list..."
                className="admin-textarea"
                rows={3}
              />
            </div>

            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="isPublic"
                checked={isPublic}
                onChange={(e) => setIsPublic(e.target.checked)}
                className="w-4 h-4 rounded border-[var(--border)] bg-[var(--surface)] text-[var(--accent)] focus:ring-[var(--accent)] focus:ring-offset-0"
              />
              <label htmlFor="isPublic" className="text-sm text-[var(--text)]">
                Make this list public (others can view it)
              </label>
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="submit"
                disabled={loading || !name.trim()}
                className="admin-btn-primary flex-1"
              >
                {loading ? 'Creating...' : 'Create List'}
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
