import { useEffect, useState } from 'react';
import { Plus, Bookmark, Trash2, Eye, EyeOff, MapPin } from 'lucide-react';
import { ListFormModal } from '../components/ListFormModal';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { useToast } from '../hooks/useToast';
import { useDeleteList, useListDetail, useUserLists } from '../hooks/useApi';

export function ListsPage() {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedListId, setSelectedListId] = useState<number | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<{ id: number; name: string } | null>(null);
  const { showError } = useToast();

  const { data: lists = [], isLoading: loading } = useUserLists();
  const {
    data: selectedList = null,
    isLoading: loadingDetail,
    error: detailError,
  } = useListDetail(selectedListId ?? 0);
  const deleteListMutation = useDeleteList();

  useEffect(() => {
    if (detailError) {
      showError('Failed to load list details');
    }
  }, [detailError, showError]);

  const handleDelete = (listId: number, name: string) => {
    setConfirmDelete({ id: listId, name });
  };

  const confirmDeleteList = async () => {
    if (!confirmDelete) return;

    try {
      await deleteListMutation.mutateAsync(confirmDelete.id);
      if (selectedListId === confirmDelete.id) {
        setSelectedListId(null);
      }
    } catch (error) {
      console.error('Failed to delete list:', error);
      showError('Failed to delete list');
    }
  };

  const handleSelectList = (listId: number) => {
    setSelectedListId((prev) => (prev === listId ? null : listId));
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
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-(--accent)"></div>
        </div>
      ) : lists.length === 0 ? (
        <div className="card text-center py-12">
          <Bookmark className="w-16 h-16 mx-auto mb-4 text-(--text-muted)" />
          <h2 className="text-xl font-semibold mb-2">No Lists Yet</h2>
          <p className="text-(--text-muted) mb-4">
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
                  selectedListId === list.id
                    ? 'ring-2 ring-(--accent) bg-(--accent-dim)'
                    : 'hover:shadow-lg'
                }`}
              >
                <div className="flex items-start justify-between mb-2">
                  <h3 className="font-semibold text-lg">{list.name}</h3>
                  <div className="flex items-center gap-1">
                    {list.is_public ? (
                      <div title="Public"><Eye className="w-4 h-4 text-(--success)" /></div>
                    ) : (
                      <div title="Private"><EyeOff className="w-4 h-4 text-(--text-muted)" /></div>
                    )}
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDelete(list.id, list.name);
                      }}
                      className="p-1 hover:bg-(--danger-dim) rounded-sm transition-colors"
                      title="Delete list"
                    >
                      <Trash2 className="w-4 h-4 text-(--danger)" />
                    </button>
                  </div>
                </div>
                {list.description && (
                  <p className="text-sm text-(--text-muted) mb-2">
                    {list.description}
                  </p>
                )}
                <p className="text-xs text-(--text-muted)">
                  {list.restaurant_count || 0} restaurant{list.restaurant_count !== 1 ? 's' : ''}
                </p>
              </button>
            ))}
          </div>

          {/* List detail */}
          <div className="lg:col-span-2">
            {loadingDetail ? (
              <div className="card text-center py-12">
                <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-(--accent)"></div>
              </div>
            ) : selectedList ? (
              <div>
                <div className="card p-6 mb-4">
                  <h2 className="text-2xl font-bold mb-2">{selectedList.list.name}</h2>
                  {selectedList.list.description && (
                    <p className="text-(--text-muted) mb-2">
                      {selectedList.list.description}
                    </p>
                  )}
                  <p className="text-sm text-(--text-muted)">
                    {selectedList.list.is_public ? 'Public list' : 'Private list'} •{' '}
                    {selectedList.restaurants.length} restaurant{selectedList.restaurants.length !== 1 ? 's' : ''}
                  </p>
                </div>

                {selectedList.restaurants.length === 0 ? (
                  <div className="card text-center py-8 text-(--text-muted)">
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
                              <p className="text-sm text-(--text-muted) flex items-start gap-1 mb-2">
                                <MapPin className="w-4 h-4 mt-0.5 shrink-0" />
                                {item.restaurant.address}
                              </p>
                            )}
                            {item.notes && (
                              <p className="text-sm text-(--text) mt-2 p-2 bg-(--warning)/10 rounded-sm">
                                <strong>Note:</strong> {item.notes}
                              </p>
                            )}
                            <p className="text-xs text-(--text-muted) mt-2">
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
              <div className="card text-center py-12 text-(--text-muted)">
                <Bookmark className="w-16 h-16 mx-auto mb-4 text-(--text-muted) dark:text-(--text)" />
                <p>Select a list to view its restaurants</p>
              </div>
            )}
          </div>
        </div>
      )}

      {showCreateModal && (
        <ListFormModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={() => setShowCreateModal(false)}
        />
      )}

      <ConfirmDialog
        isOpen={confirmDelete !== null}
        onClose={() => setConfirmDelete(null)}
        onConfirm={confirmDeleteList}
        title="Delete List"
        message={`Are you sure you want to delete "${confirmDelete?.name}"?`}
        confirmText="Delete"
        isDangerous
      />
    </div>
  );
}
