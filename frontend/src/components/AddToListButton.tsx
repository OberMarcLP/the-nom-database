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
        className="btn btn-secondary flex items-center gap-2"
      >
        <Bookmark className="w-4 h-4" />
        Add to List
      </button>

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-content max-w-md" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold">Add "{restaurantName}" to List</h2>
              <button
                onClick={() => setShowModal(false)}
                className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {loading ? (
              <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                Loading lists...
              </div>
            ) : (
              <>
                {lists.length === 0 && !showCreateForm ? (
                  <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                    <p className="mb-4">You don't have any lists yet.</p>
                    <button
                      onClick={() => setShowCreateForm(true)}
                      className="btn btn-primary flex items-center gap-2 mx-auto"
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
                          className={`w-full p-3 rounded-lg border-2 transition-colors text-left flex items-center justify-between ${
                            list.contains_restaurant
                              ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                              : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                          }`}
                        >
                          <div>
                            <h3 className="font-semibold">{list.name}</h3>
                            {list.description && (
                              <p className="text-sm text-gray-600 dark:text-gray-400">
                                {list.description}
                              </p>
                            )}
                            <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                              {list.restaurant_count || 0} restaurant{list.restaurant_count !== 1 ? 's' : ''}
                            </p>
                          </div>
                          {list.contains_restaurant && (
                            <Check className="w-5 h-5 text-blue-500 flex-shrink-0" />
                          )}
                        </button>
                      ))}
                    </div>

                    {!showCreateForm ? (
                      <button
                        onClick={() => setShowCreateForm(true)}
                        className="btn btn-secondary w-full flex items-center justify-center gap-2"
                      >
                        <Plus className="w-4 h-4" />
                        Create New List
                      </button>
                    ) : (
                      <form onSubmit={handleCreateList} className="card p-4 bg-gray-50 dark:bg-gray-800">
                        <h3 className="font-semibold mb-3">Create New List</h3>
                        <input
                          type="text"
                          value={newListName}
                          onChange={(e) => setNewListName(e.target.value)}
                          placeholder="List name (e.g., Want to Try)"
                          className="input mb-3"
                          autoFocus
                          required
                        />
                        <div className="flex gap-2">
                          <button
                            type="submit"
                            disabled={creating || !newListName.trim()}
                            className="btn btn-primary flex-1 disabled:opacity-50"
                          >
                            {creating ? 'Creating...' : 'Create'}
                          </button>
                          <button
                            type="button"
                            onClick={() => {
                              setShowCreateForm(false);
                              setNewListName('');
                            }}
                            className="btn btn-secondary"
                          >
                            Cancel
                          </button>
                        </div>
                      </form>
                    )}
                  </>
                )}
              </>
            )}
          </div>
        </div>
      )}
    </>
  );
}
