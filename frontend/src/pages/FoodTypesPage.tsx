import { useState, useEffect } from 'react';
import { Plus, Edit, Trash2, Loader2, Utensils } from 'lucide-react';
import { FoodType, getFoodTypes, createFoodType, updateFoodType, deleteFoodType } from '../services/api';
import { PermissionGuard } from '../components/PermissionGuard';

export function FoodTypesPage() {
  const [foodTypes, setFoodTypes] = useState<FoodType[]>([]);
  const [loading, setLoading] = useState(true);
  const [newName, setNewName] = useState('');
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState('');

  const fetchFoodTypes = async () => {
    try {
      const data = await getFoodTypes();
      setFoodTypes(data);
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFoodTypes();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim()) return;
    try {
      await createFoodType(newName);
      setNewName('');
      fetchFoodTypes();
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const handleUpdate = async (id: number) => {
    if (!editName.trim()) return;
    try {
      await updateFoodType(id, editName);
      setEditingId(null);
      setEditName('');
      fetchFoodTypes();
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this food type?')) return;
    try {
      await deleteFoodType(id);
      fetchFoodTypes();
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
        <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center gap-3 mb-6">
        <div className="p-3 rounded-xl bg-linear-to-br from-green-500/20 to-emerald-500/20 backdrop-blur-sm">
          <Utensils className="w-6 h-6 text-green-500" />
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
        <p className="text-center text-gray-500 dark:text-gray-400 py-8">
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
                        className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
                      >
                        <Edit className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => handleDelete(foodType.id)}
                        className="p-2 hover:bg-red-100 dark:hover:bg-red-900/30 text-red-600 dark:text-red-400 rounded-lg"
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
    </div>
  );
}
