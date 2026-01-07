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
        className="px-4 py-2.5 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors flex items-center gap-2"
      >
        <Bookmark className="w-4 h-4" />
        Add to List
      </button>

      {showModal && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black/50 backdrop-blur-sm"
            onClick={() => setShowModal(false)}
          />

          {/* Modal */}
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md p-8" onClick={(e) => e.stopPropagation()}>
              {/* Close button */}
              <button
                onClick={() => setShowModal(false)}
                className="absolute top-4 right-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors"
              >
                <X className="w-6 h-6" />
              </button>

              {/* Header */}
              <div className="mb-6">
                <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
                  Add "{restaurantName}" to List
                </h2>
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                  Select lists to add this restaurant to
                </p>
              </div>

              {loading ? (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                  Loading lists...
                </div>
              ) : (
                <>
                  {lists.length === 0 && !showCreateForm ? (
                    <div className="text-center py-8">
                      <p className="mb-4 text-gray-600 dark:text-gray-400">You don't have any lists yet.</p>
                      <button
                        onClick={() => setShowCreateForm(true)}
                        className="inline-flex items-center gap-2 px-4 py-2.5 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
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
                            className={`w-full p-3 rounded-lg border-2 transition-all text-left flex items-center justify-between ${
                              list.contains_restaurant
                                ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 text-gray-900 dark:text-white'
                                : 'border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-600 text-gray-900 dark:text-gray-100'
                            }`}
                          >
                            <div>
                              <h3 className="font-semibold text-gray-900 dark:text-white">{list.name}</h3>
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
                          className="w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors flex items-center justify-center gap-2"
                        >
                          <Plus className="w-4 h-4" />
                          Create New List
                        </button>
                      ) : (
                        <div className="bg-gray-50 dark:bg-gray-900/50 p-4 rounded-lg border border-gray-200 dark:border-gray-700">
                          <h3 className="font-semibold mb-3 text-gray-900 dark:text-white">Create New List</h3>
                          <form onSubmit={handleCreateList} className="space-y-3">
                            <input
                              type="text"
                              value={newListName}
                              onChange={(e) => setNewListName(e.target.value)}
                              placeholder="List name (e.g., Want to Try)"
                              className="appearance-none relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-700"
                              autoFocus
                              required
                            />
                            <div className="flex gap-2">
                              <button
                                type="submit"
                                disabled={creating || !newListName.trim()}
                                className="flex-1 flex justify-center py-2.5 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                              >
                                {creating ? 'Creating...' : 'Create'}
                              </button>
                              <button
                                type="button"
                                onClick={() => {
                                  setShowCreateForm(false);
                                  setNewListName('');
                                }}
                                className="px-4 py-2.5 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
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
