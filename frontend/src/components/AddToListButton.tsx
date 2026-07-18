import { useState, useEffect } from 'react';
import { Bookmark, Plus, Check, X } from 'lucide-react';
import { getRestaurantLists, addRestaurantToList, removeRestaurantFromList, createList, ListWithStatus } from '../services/api';

interface AddToListButtonProps {
  restaurantId: number;
  restaurantName: string;
}

export function AddToListButton({ restaurantId, restaurantName }: AddToListButtonProps) {
  const [showModal, setShowModal] = useState(false);
  const [lists, setLists] = useState<ListWithStatus[]>([]);
  const [loading, setLoading] = useState(false);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [newListName, setNewListName] = useState('');
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    if (showModal) {
      loadLists();
    }
  }, [showModal]);

  const loadLists = async () => {
    try {
      setLoading(true);
      const data = await getRestaurantLists(restaurantId);
      setLists(data);
    } catch (error) {
      console.error('Failed to load lists:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleToggleList = async (listId: number, isInList: boolean) => {
    try {
      if (isInList) {
        await removeRestaurantFromList(listId, restaurantId);
      } else {
        await addRestaurantToList(listId, restaurantId);
      }
      await loadLists();
    } catch (error) {
      console.error('Failed to toggle list:', error);
      alert('Failed to update list');
    }
  };

  const handleCreateList = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newListName.trim()) return;

    try {
      setCreating(true);
      await createList({
        name: newListName,
        is_public: false,
      });
      setNewListName('');
      setShowCreateForm(false);
      await loadLists();
    } catch (error) {
      console.error('Failed to create list:', error);
      alert('Failed to create list');
    } finally {
      setCreating(false);
    }
  };

  return (
    <>
      <button
        onClick={() => setShowModal(true)}
        className="btn-glass flex items-center gap-2"
      >
        <Bookmark className="w-4 h-4" />
        Add to List
      </button>

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-glass admin-modal-lg" onClick={(e) => e.stopPropagation()}>
            <div className="admin-modal-header">
              <h2 className="admin-modal-title">
                Add "{restaurantName}" to List
              </h2>
              <button onClick={() => setShowModal(false)} className="admin-modal-close">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="admin-modal-body">
              <p className="text-sm text-(--text-muted) mb-6">
                Select lists to add this restaurant to
              </p>

              {loading ? (
                <div className="text-center py-8 text-(--text-muted)">
                  Loading lists...
                </div>
              ) : (
                <>
                  {lists.length === 0 && !showCreateForm ? (
                    <div className="text-center py-8">
                      <p className="mb-4 text-(--text-muted)">You don't have any lists yet.</p>
                      <button
                        onClick={() => setShowCreateForm(true)}
                        className="admin-btn-primary inline-flex items-center gap-2"
                      >
                        <Plus className="w-4 h-4" />
                        Create Your First List
                      </button>
                    </div>
                  ) : (
                    <>
                      <div className="space-y-2 mb-4 max-h-96 overflow-y-auto">
                        {lists.map((list) => (
                          <button
                            key={list.id}
                            onClick={() => handleToggleList(list.id, list.contains_restaurant)}
                            className={`w-full p-3 rounded border-2 transition-all text-left flex items-center justify-between ${
                              list.contains_restaurant
                                ? 'border-(--accent) bg-(--accent)/10'
                                : 'border-(--border) hover:border-(--accent)/50'
                            }`}
                          >
                            <div>
                              <h3 className="font-semibold text-(--text)">{list.name}</h3>
                              {list.description && (
                                <p className="text-sm text-(--text-muted)">
                                  {list.description}
                                </p>
                              )}
                              <p className="text-xs text-(--text-muted) mt-1">
                                {list.restaurant_count || 0} restaurant{list.restaurant_count !== 1 ? 's' : ''}
                              </p>
                            </div>
                            {list.contains_restaurant && (
                              <Check className="w-5 h-5 text-(--accent) shrink-0" />
                            )}
                          </button>
                        ))}
                      </div>

                      {!showCreateForm ? (
                        <button
                          onClick={() => setShowCreateForm(true)}
                          className="admin-btn w-full flex items-center justify-center gap-2"
                        >
                          <Plus className="w-4 h-4" />
                          Create New List
                        </button>
                      ) : (
                        <div className="admin-card p-4">
                          <h3 className="admin-card-title mb-4">Create New List</h3>
                          <form onSubmit={handleCreateList} className="space-y-4">
                            <input
                              type="text"
                              value={newListName}
                              onChange={(e) => setNewListName(e.target.value)}
                              placeholder="List name (e.g., Want to Try)"
                              className="admin-input"
                              autoFocus
                              required
                            />
                            <div className="flex gap-2">
                              <button
                                type="submit"
                                disabled={creating || !newListName.trim()}
                                className="admin-btn-primary flex-1"
                              >
                                {creating ? 'Creating...' : 'Create'}
                              </button>
                              <button
                                type="button"
                                onClick={() => {
                                  setShowCreateForm(false);
                                  setNewListName('');
                                }}
                                className="admin-btn"
                              >
                                Cancel
                              </button>
                            </div>
                          </form>
                        </div>
                      )}
                    </>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
