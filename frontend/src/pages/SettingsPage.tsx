import { useState } from 'react';
import { Plus, Edit, Trash2, Loader2, Tag, Utensils, Settings as SettingsIcon } from 'lucide-react';
import { Category, FoodType } from '../services/api';
import { ConfirmDialog } from '../components/ConfirmDialog';
import {
  useCategories,
  useCreateCategory,
  useCreateFoodType,
  useDeleteCategory,
  useDeleteFoodType,
  useFoodTypes,
  useUpdateCategory,
  useUpdateFoodType,
} from '../hooks/useApi';

export function SettingsPage() {
  const [newCategoryName, setNewCategoryName] = useState('');
  const [editingCategoryId, setEditingCategoryId] = useState<number | null>(null);
  const [editCategoryName, setEditCategoryName] = useState('');

  const [newFoodTypeName, setNewFoodTypeName] = useState('');
  const [editingFoodTypeId, setEditingFoodTypeId] = useState<number | null>(null);
  const [editFoodTypeName, setEditFoodTypeName] = useState('');

  const [confirmDeleteCategory, setConfirmDeleteCategory] = useState<{ id: number; name: string } | null>(null);
  const [confirmDeleteFoodType, setConfirmDeleteFoodType] = useState<{ id: number; name: string } | null>(null);

  const { data: categories = [], isLoading: loadingCategories } = useCategories();
  const { data: foodTypes = [], isLoading: loadingFoodTypes } = useFoodTypes();
  const loading = loadingCategories || loadingFoodTypes;

  const createCategoryMutation = useCreateCategory();
  const updateCategoryMutation = useUpdateCategory();
  const deleteCategoryMutation = useDeleteCategory();
  const createFoodTypeMutation = useCreateFoodType();
  const updateFoodTypeMutation = useUpdateFoodType();
  const deleteFoodTypeMutation = useDeleteFoodType();

  // Category handlers
  const handleCreateCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCategoryName.trim()) return;
    try {
      await createCategoryMutation.mutateAsync(newCategoryName);
      setNewCategoryName('');
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const handleUpdateCategory = async (id: number) => {
    if (!editCategoryName.trim()) return;
    try {
      await updateCategoryMutation.mutateAsync({ id, name: editCategoryName });
      setEditingCategoryId(null);
      setEditCategoryName('');
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const handleDeleteCategory = async () => {
    if (!confirmDeleteCategory) return;
    try {
      await deleteCategoryMutation.mutateAsync(confirmDeleteCategory.id);
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const startEditCategory = (category: Category) => {
    setEditingCategoryId(category.id);
    setEditCategoryName(category.name);
  };

  // Food Type handlers
  const handleCreateFoodType = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newFoodTypeName.trim()) return;
    try {
      await createFoodTypeMutation.mutateAsync(newFoodTypeName);
      setNewFoodTypeName('');
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const handleUpdateFoodType = async (id: number) => {
    if (!editFoodTypeName.trim()) return;
    try {
      await updateFoodTypeMutation.mutateAsync({ id, name: editFoodTypeName });
      setEditingFoodTypeId(null);
      setEditFoodTypeName('');
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const handleDeleteFoodType = async () => {
    if (!confirmDeleteFoodType) return;
    try {
      await deleteFoodTypeMutation.mutateAsync(confirmDeleteFoodType.id);
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    }
  };

  const startEditFoodType = (foodType: FoodType) => {
    setEditingFoodTypeId(foodType.id);
    setEditFoodTypeName(foodType.name);
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
      <div className="flex items-center gap-3 mb-8">
        <div className="p-3 rounded-xl bg-linear-to-br from-(--accent-dim) to-(--accent-dim)">
          <SettingsIcon className="w-6 h-6 text-(--accent)" />
        </div>
        <h1 className="text-3xl font-bold text-gradient">Settings</h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Categories Section */}
        <div className="card-glass p-6 space-y-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-lg bg-(--accent-dim)">
              <Tag className="w-5 h-5 text-(--info)" />
            </div>
            <h2 className="text-xl font-bold">Categories</h2>
          </div>

          <form onSubmit={handleCreateCategory} className="flex gap-2">
            <input
              type="text"
              value={newCategoryName}
              onChange={(e) => setNewCategoryName(e.target.value)}
              placeholder="New category name..."
              className="input-glass flex-1"
            />
            <button type="submit" className="btn-glass-primary flex items-center gap-2">
              <Plus className="w-5 h-5" />
              Add
            </button>
          </form>

          {categories.length === 0 ? (
            <p className="text-center text-(--text-muted) py-8">
              No categories yet. Add your first category above.
            </p>
          ) : (
            <div className="space-y-2 max-h-[400px] overflow-y-auto">
              {categories.map((category) => (
                <div key={category.id} className="card-glass p-4 flex items-center justify-between group hover:shadow-lg transition-all duration-200">
                  {editingCategoryId === category.id ? (
                    <div className="flex items-center gap-2 flex-1">
                      <input
                        type="text"
                        value={editCategoryName}
                        onChange={(e) => setEditCategoryName(e.target.value)}
                        className="input-glass flex-1"
                        autoFocus
                      />
                      <button
                        onClick={() => handleUpdateCategory(category.id)}
                        className="btn-glass-success px-3 py-2"
                      >
                        Save
                      </button>
                      <button
                        onClick={() => {
                          setEditingCategoryId(null);
                          setEditCategoryName('');
                        }}
                        className="btn-glass px-3 py-2"
                      >
                        Cancel
                      </button>
                    </div>
                  ) : (
                    <>
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-full bg-(--accent-dim) flex items-center justify-center">
                          <Tag className="w-4 h-4 text-(--info)" />
                        </div>
                        <span className="font-medium">{category.name}</span>
                      </div>
                      <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                          onClick={() => startEditCategory(category)}
                          className="p-2 rounded-lg btn-glass hover:bg-(--accent-dim)"
                        >
                          <Edit className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => setConfirmDeleteCategory({ id: category.id, name: category.name })}
                          className="p-2 rounded-lg btn-glass hover:bg-(--danger-dim)"
                        >
                          <Trash2 className="w-4 h-4 text-(--danger)" />
                        </button>
                      </div>
                    </>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Food Types Section */}
        <div className="card-glass p-6 space-y-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-lg bg-(--accent-dim)">
              <Utensils className="w-5 h-5 text-(--success)" />
            </div>
            <h2 className="text-xl font-bold">Food Types</h2>
          </div>

          <form onSubmit={handleCreateFoodType} className="flex gap-2">
            <input
              type="text"
              value={newFoodTypeName}
              onChange={(e) => setNewFoodTypeName(e.target.value)}
              placeholder="New food type name..."
              className="input-glass flex-1"
            />
            <button type="submit" className="btn-glass-primary flex items-center gap-2">
              <Plus className="w-5 h-5" />
              Add
            </button>
          </form>

          {foodTypes.length === 0 ? (
            <p className="text-center text-(--text-muted) py-8">
              No food types yet. Add your first food type above.
            </p>
          ) : (
            <div className="space-y-2 max-h-[400px] overflow-y-auto">
              {foodTypes.map((foodType) => (
                <div key={foodType.id} className="card-glass p-4 flex items-center justify-between group hover:shadow-lg transition-all duration-200">
                  {editingFoodTypeId === foodType.id ? (
                    <div className="flex items-center gap-2 flex-1">
                      <input
                        type="text"
                        value={editFoodTypeName}
                        onChange={(e) => setEditFoodTypeName(e.target.value)}
                        className="input-glass flex-1"
                        autoFocus
                      />
                      <button
                        onClick={() => handleUpdateFoodType(foodType.id)}
                        className="btn-glass-success px-3 py-2"
                      >
                        Save
                      </button>
                      <button
                        onClick={() => {
                          setEditingFoodTypeId(null);
                          setEditFoodTypeName('');
                        }}
                        className="btn-glass px-3 py-2"
                      >
                        Cancel
                      </button>
                    </div>
                  ) : (
                    <>
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-full bg-(--accent-dim) flex items-center justify-center">
                          <Utensils className="w-4 h-4 text-(--success)" />
                        </div>
                        <span className="font-medium">{foodType.name}</span>
                      </div>
                      <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                          onClick={() => startEditFoodType(foodType)}
                          className="p-2 rounded-lg btn-glass hover:bg-(--accent-dim)"
                        >
                          <Edit className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => setConfirmDeleteFoodType({ id: foodType.id, name: foodType.name })}
                          className="p-2 rounded-lg btn-glass hover:bg-(--danger-dim)"
                        >
                          <Trash2 className="w-4 h-4 text-(--danger)" />
                        </button>
                      </div>
                    </>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <ConfirmDialog
        isOpen={confirmDeleteCategory !== null}
        onClose={() => setConfirmDeleteCategory(null)}
        onConfirm={handleDeleteCategory}
        title="Delete Category"
        message={`Are you sure you want to delete "${confirmDeleteCategory?.name}"?`}
        confirmText="Delete"
        isDangerous
      />

      <ConfirmDialog
        isOpen={confirmDeleteFoodType !== null}
        onClose={() => setConfirmDeleteFoodType(null)}
        onConfirm={handleDeleteFoodType}
        title="Delete Food Type"
        message={`Are you sure you want to delete "${confirmDeleteFoodType?.name}"?`}
        confirmText="Delete"
        isDangerous
      />
    </div>
  );
}
