import { useState } from 'react';
import { Plus, Edit, Trash2, Loader2, Utensils } from 'lucide-react';
import { FoodType } from '../services/api';
import { PermissionGuard } from '../components/PermissionGuard';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { useCreateFoodType, useDeleteFoodType, useFoodTypes, useUpdateFoodType } from '../hooks/useApi';

export function FoodTypesPage() {
  const [newName, setNewName] = useState('');
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState('');
  const [confirmDelete, setConfirmDelete] = useState<{ id: number; name: string } | null>(null);

  const { data: foodTypes = [], isLoading: loading } = useFoodTypes();
  const createFoodTypeMutation = useCreateFoodType();
  const updateFoodTypeMutation = useUpdateFoodType();
  const deleteFoodTypeMutation = useDeleteFoodType();

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim()) return;
    try {
      await createFoodTypeMutation.mutateAsync(newName);
      setNewName('');
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const handleUpdate = async (id: number) => {
    if (!editName.trim()) return;
    try {
      await updateFoodTypeMutation.mutateAsync({ id, name: editName });
      setEditingId(null);
      setEditName('');
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) return;
    try {
      await deleteFoodTypeMutation.mutateAsync(confirmDelete.id);
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const startEdit = (foodType: FoodType) => {
    setEditingId(foodType.id);
    setEditName(foodType.name);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-(--info)" />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center gap-3 mb-6">
        <div className="p-3 rounded-xl bg-linear-to-br from-(--accent-dim) to-emerald-500/20">
          <Utensils className="w-6 h-6 text-(--success)" />
        </div>
        <h1 className="text-3xl font-bold text-gradient">Food Types</h1>
      </div>

      <PermissionGuard permissions={['food_types.create', 'food_types.update', 'food_types.delete']}>
        <form onSubmit={handleCreate} className="flex gap-2 mb-6">
          <input
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="New food type name..."
            className="input-glass flex-1"
          />
          <button type="submit" className="btn-glass-primary flex items-center gap-2">
            <Plus className="w-5 h-5" />
            Add
          </button>
        </form>
      </PermissionGuard>

      {foodTypes.length === 0 ? (
        <p className="text-center text-(--text-muted) py-8">
          No food types yet. Add your first food type above.
        </p>
      ) : (
        <div className="space-y-2">
          {foodTypes.map((foodType) => (
            <div key={foodType.id} className="card-glass flex items-center justify-between">
              {editingId === foodType.id ? (
                <div className="flex items-center gap-2 flex-1">
                  <input
                    type="text"
                    value={editName}
                    onChange={(e) => setEditName(e.target.value)}
                    className="input-glass flex-1"
                    autoFocus
                  />
                  <button
                    onClick={() => handleUpdate(foodType.id)}
                    className="btn-glass-primary text-sm"
                  >
                    Save
                  </button>
                  <button
                    onClick={() => setEditingId(null)}
                    className="btn btn-secondary text-sm"
                  >
                    Cancel
                  </button>
                </div>
              ) : (
                <>
                  <span className="font-medium">{foodType.name}</span>
                  <PermissionGuard permissions={['food_types.create', 'food_types.update', 'food_types.delete']}>
                    <div className="flex gap-2">
                      <button
                        onClick={() => startEdit(foodType)}
                        className="p-2 hover:bg-(--surface-hover) rounded-lg"
                      >
                        <Edit className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => setConfirmDelete({ id: foodType.id, name: foodType.name })}
                        className="p-2 hover:bg-(--danger-dim) text-(--danger) rounded-lg"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </PermissionGuard>
                </>
              )}
            </div>
          ))}
        </div>
      )}

      <ConfirmDialog
        isOpen={confirmDelete !== null}
        onClose={() => setConfirmDelete(null)}
        onConfirm={handleDelete}
        title="Delete Food Type"
        message={`Are you sure you want to delete "${confirmDelete?.name}"?`}
        confirmText="Delete"
        isDangerous
      />
    </div>
  );
}
