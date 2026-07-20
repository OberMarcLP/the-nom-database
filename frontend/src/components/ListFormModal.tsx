import { useState } from 'react';
import { useCreateList } from '../hooks/useApi';
import { useToast } from '../hooks/useToast';
import { Modal } from './Modal';

interface ListFormModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

export function ListFormModal({ onClose, onSuccess }: ListFormModalProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isPublic, setIsPublic] = useState(false);
  const { showError } = useToast();
  const createListMutation = useCreateList();
  const loading = createListMutation.isPending;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    try {
      await createListMutation.mutateAsync({
        name: name.trim(),
        description: description.trim() || undefined,
        is_public: isPublic,
      });
      onSuccess();
    } catch (error) {
      console.error('Failed to create list:', error);
      showError('Failed to create list');
    }
  };

  return (
    <Modal isOpen onClose={onClose} title="Create New List">
      <p className="text-sm text-(--text-muted) mb-6">
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
            data-autofocus
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
            className="w-4 h-4 rounded-sm border-(--border) bg-(--surface) text-(--accent) focus:ring-(--accent) focus:ring-offset-0"
          />
          <label htmlFor="isPublic" className="text-sm text-(--text)">
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
    </Modal>
  );
}
