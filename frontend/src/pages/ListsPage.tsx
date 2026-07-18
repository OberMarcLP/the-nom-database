import { useState, useEffect } from 'react';
import { Plus, Bookmark, Trash2, Eye, EyeOff, MapPin } from 'lucide-react';
import { getUserLists, deleteList, getList, RestaurantList, ListWithRestaurants } from '../services/api';
import { ListFormModal } from '../components/ListFormModal';

export function ListsPage() {
  const [lists, setLists] = useState<RestaurantList[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedList, setSelectedList] = useState<ListWithRestaurants | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  useEffect(() => {
    loadLists();
  }, []);

  const loadLists = async () => {
    try {
      setLoading(true);
      const data = await getUserLists();
      setLists(data);
    } catch (error) {
      console.error('Failed to load lists:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (listId: number) => {
    if (!confirm('Are you sure you want to delete this list?')) return;

    try {
      await deleteList(listId);
      await loadLists();
      if (selectedList?.list.id === listId) {
        setSelectedList(null);
      }
    } catch (error) {
      console.error('Failed to delete list:', error);
      alert('Failed to delete list');
    }
  };

  const handleSelectList = async (listId: number) => {
    if (selectedList?.list.id === listId) {
      setSelectedList(null);
      return;
    }

    try {
      setLoadingDetail(true);
      const data = await getList(listId);
      setSelectedList(data);
    } catch (error) {
      console.error('Failed to load list detail:', error);
      alert('Failed to load list details');
    } finally {
      setLoadingDetail(false);
    }
  };

  return (
    <div className="max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <Bookmark className="w-8 h-8" />
          My Lists
        </h1>
        <button
          onClick={() => setShowCreateModal(true)}
          className="btn btn-primary flex items-center gap-2"
        >
          <Plus className="w-4 h-4" />
          Create List
        </button>
      </div>

      {loading ? (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      ) : lists.length === 0 ? (
        <div className="card text-center py-12">
          <Bookmark className="w-16 h-16 mx-auto mb-4 text-gray-400" />
          <h2 className="text-xl font-semibold mb-2">No Lists Yet</h2>
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            Create your first list to organize restaurants you want to try or your favorites!
          </p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="btn btn-primary flex items-center gap-2 mx-auto"
          >
            <Plus className="w-4 h-4" />
            Create Your First List
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Lists sidebar */}
          <div className="lg:col-span-1 space-y-3">
            {lists.map((list) => (
              <button
                key={list.id}
                onClick={() => handleSelectList(list.id)}
                className={`w-full card p-4 text-left transition-all ${
                  selectedList?.list.id === list.id
                    ? 'ring-2 ring-blue-500 bg-blue-50 dark:bg-blue-900/20'
                    : 'hover:shadow-lg'
                }`}
              >
                <div className="flex items-start justify-between mb-2">
                  <h3 className="font-semibold text-lg">{list.name}</h3>
                  <div className="flex items-center gap-1">
                    {list.is_public ? (
                      <div title="Public"><Eye className="w-4 h-4 text-green-600" /></div>
                    ) : (
                      <div title="Private"><EyeOff className="w-4 h-4 text-gray-400" /></div>
                    )}
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDelete(list.id);
                      }}
                      className="p-1 hover:bg-red-100 dark:hover:bg-red-900 rounded-sm transition-colors"
                      title="Delete list"
                    >
                      <Trash2 className="w-4 h-4 text-red-600" />
                    </button>
                  </div>
                </div>
                {list.description && (
                  <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                    {list.description}
                  </p>
                )}
                <p className="text-xs text-gray-500 dark:text-gray-500">
                  {list.restaurant_count || 0} restaurant{list.restaurant_count !== 1 ? 's' : ''}
                </p>
              </button>
            ))}
          </div>

          {/* List detail */}
          <div className="lg:col-span-2">
            {loadingDetail ? (
              <div className="card text-center py-12">
                <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : selectedList ? (
              <div>
                <div className="card p-6 mb-4">
                  <h2 className="text-2xl font-bold mb-2">{selectedList.list.name}</h2>
                  {selectedList.list.description && (
                    <p className="text-gray-600 dark:text-gray-400 mb-2">
                      {selectedList.list.description}
                    </p>
                  )}
                  <p className="text-sm text-gray-500 dark:text-gray-500">
                    {selectedList.list.is_public ? 'Public list' : 'Private list'} •{' '}
                    {selectedList.restaurants.length} restaurant{selectedList.restaurants.length !== 1 ? 's' : ''}
                  </p>
                </div>

                {selectedList.restaurants.length === 0 ? (
                  <div className="card text-center py-8 text-gray-500 dark:text-gray-400">
                    This list is empty. Add restaurants from their detail pages!
                  </div>
                ) : (
                  <div className="space-y-3">
                    {selectedList.restaurants.map((item) => (
                      <div key={item.id} className="card p-4 hover:shadow-lg transition-shadow">
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <h3 className="font-semibold text-lg mb-1">
                              {item.restaurant?.name}
                            </h3>
                            {item.restaurant?.address && (
                              <p className="text-sm text-gray-600 dark:text-gray-400 flex items-start gap-1 mb-2">
                                <MapPin className="w-4 h-4 mt-0.5 shrink-0" />
                                {item.restaurant.address}
                              </p>
                            )}
                            {item.notes && (
                              <p className="text-sm text-gray-700 dark:text-gray-300 mt-2 p-2 bg-yellow-50 dark:bg-yellow-900/20 rounded-sm">
                                <strong>Note:</strong> {item.notes}
                              </p>
                            )}
                            <p className="text-xs text-gray-500 dark:text-gray-500 mt-2">
                              Added {new Date(item.added_at).toLocaleDateString()}
                            </p>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ) : (
              <div className="card text-center py-12 text-gray-500 dark:text-gray-400">
                <Bookmark className="w-16 h-16 mx-auto mb-4 text-gray-300 dark:text-gray-700" />
                <p>Select a list to view its restaurants</p>
              </div>
            )}
          </div>
        </div>
      )}

      {showCreateModal && (
        <ListFormModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={() => {
            setShowCreateModal(false);
            loadLists();
          }}
        />
      )}
    </div>
  );
}
