import { useState, useEffect } from 'react';
import { Plus, Edit, Trash2, Loader2, Tag, Utensils, X } from 'lucide-react';
import { Category, FoodType, getCategories, getFoodTypes, createCategory, updateCategory, deleteCategory, createFoodType, updateFoodType, deleteFoodType } from '../services/api';
import { useEscapeKey } from '../hooks/useEscapeKey';

interface SettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function SettingsModal({ isOpen, onClose }: SettingsModalProps) {
  useEscapeKey(onClose, isOpen);
  const [categories, setCategories] = useState<Category[]>([]);
  const [foodTypes, setFoodTypes] = useState<FoodType[]>([]);
  const [loading, setLoading] = useState(true);

  const [newCategoryName, setNewCategoryName] = useState('');
  const [editingCategoryId, setEditingCategoryId] = useState<number | null>(null);
  const [editCategoryName, setEditCategoryName] = useState('');

  const [newFoodTypeName, setNewFoodTypeName] = useState('');
  const [editingFoodTypeId, setEditingFoodTypeId] = useState<number | null>(null);
  const [editFoodTypeName, setEditFoodTypeName] = useState('');

  const fetchData = async () => {
    try {
      const [cats, fts] = await Promise.all([getCategories(), getFoodTypes()]);
      setCategories(cats);
      setFoodTypes(fts);
    } catch (error) {
      // Handle error silently
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchData();
    }
  }, [isOpen]);

  // Category handlers
  const handleCreateCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCategoryName.trim()) return;
    try {
      await createCategory(newCategoryName);
      setNewCategoryName('');
      fetchData();
    } catch (error) {
      // Handle error
    }
  };

  const handleUpdateCategory = async (id: number) => {
    if (!editCategoryName.trim()) return;
    try {
      await updateCategory(id, editCategoryName);
      setEditingCategoryId(null);
      setEditCategoryName('');
      fetchData();
    } catch (error) {
      // Handle error
    }
  };

  const handleDeleteCategory = async (id: number) => {
    if (!confirm('Are you sure you want to delete this category?')) return;
    try {
      await deleteCategory(id);
      fetchData();
    } catch (error) {
      // Handle error
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
      await createFoodType(newFoodTypeName);
      setNewFoodTypeName('');
      fetchData();
    } catch (error) {
      // Handle error
    }
  };

  const handleUpdateFoodType = async (id: number) => {
    if (!editFoodTypeName.trim()) return;
    try {
      await updateFoodType(id, editFoodTypeName);
      setEditingFoodTypeId(null);
      setEditFoodTypeName('');
      fetchData();
    } catch (error) {
      // Handle error
    }
  };

  const handleDeleteFoodType = async (id: number) => {
    if (!confirm('Are you sure you want to delete this food type?')) return;
    try {
      await deleteFoodType(id);
      fetchData();
    } catch (error) {
      // Handle error
    }
  };

  const startEditFoodType = (foodType: FoodType) => {
    setEditingFoodTypeId(foodType.id);
    setEditFoodTypeName(foodType.name);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-4xl p-8 max-h-[90vh] overflow-y-auto">
          {/* Close button */}
          <button
            onClick={onClose}
            className="absolute top-4 right-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors"
          >
            <X className="w-6 h-6" />
          </button>

          {/* Header */}
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">Settings</h2>

          {loading ? (
            <div className="flex items-center justify-center h-64">
              <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Categories Section */}
              <div className="space-y-6">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-blue-500/20">
                    <Tag className="w-5 h-5 text-blue-500" />
                  </div>
                  <h3 className="text-xl font-bold text-gray-900 dark:text-white">Categories</h3>
                </div>

                <form onSubmit={handleCreateCategory} className="flex gap-2">
                  <input
                    type="text"
                    value={newCategoryName}
                    onChange={(e) => setNewCategoryName(e.target.value)}
                    placeholder="New category name..."
                    className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button type="submit" className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors">
                    <Plus className="w-5 h-5" />
                    Add
                  </button>
                </form>

                {categories.length === 0 ? (
                  <p className="text-center text-gray-500 dark:text-gray-400 py-8">
                    No categories yet. Add your first category above.
                  </p>
                ) : (
                  <div className="space-y-2 max-h-[300px] overflow-y-auto">
                    {categories.map((category) => (
                      <div key={category.id} className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg flex items-center justify-between group hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors">
                        {editingCategoryId === category.id ? (
                          <div className="flex items-center gap-2 flex-1">
                            <input
                              type="text"
                              value={editCategoryName}
                              onChange={(e) => setEditCategoryName(e.target.value)}
                              className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                              autoFocus
                            />
                            <button
                              onClick={() => handleUpdateCategory(category.id)}
                              className="px-3 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors"
                            >
                              Save
                            </button>
                            <button
                              onClick={() => {
                                setEditingCategoryId(null);
                                setEditCategoryName('');
                              }}
                              className="px-3 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors"
                            >
                              Cancel
                            </button>
                          </div>
                        ) : (
                          <>
                            <div className="flex items-center gap-3">
                              <div className="w-8 h-8 rounded-full bg-blue-500/20 flex items-center justify-center">
                                <Tag className="w-4 h-4 text-blue-500" />
                              </div>
                              <span className="font-medium text-gray-900 dark:text-white">{category.name}</span>
                            </div>
                            <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                              <button
                                onClick={() => startEditCategory(category)}
                                className="p-2 rounded-lg hover:bg-blue-500/20 transition-colors"
                              >
                                <Edit className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                              </button>
                              <button
                                onClick={() => handleDeleteCategory(category.id)}
                                className="p-2 rounded-lg hover:bg-red-500/20 transition-colors"
                              >
                                <Trash2 className="w-4 h-4 text-red-500" />
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
              <div className="space-y-6">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-green-500/20">
                    <Utensils className="w-5 h-5 text-green-500" />
                  </div>
                  <h3 className="text-xl font-bold text-gray-900 dark:text-white">Food Types</h3>
                </div>

                <form onSubmit={handleCreateFoodType} className="flex gap-2">
                  <input
                    type="text"
                    value={newFoodTypeName}
                    onChange={(e) => setNewFoodTypeName(e.target.value)}
                    placeholder="New food type name..."
                    className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button type="submit" className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors">
                    <Plus className="w-5 h-5" />
                    Add
                  </button>
                </form>

                {foodTypes.length === 0 ? (
                  <p className="text-center text-gray-500 dark:text-gray-400 py-8">
                    No food types yet. Add your first food type above.
                  </p>
                ) : (
                  <div className="space-y-2 max-h-[300px] overflow-y-auto">
                    {foodTypes.map((foodType) => (
                      <div key={foodType.id} className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg flex items-center justify-between group hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors">
                        {editingFoodTypeId === foodType.id ? (
                          <div className="flex items-center gap-2 flex-1">
                            <input
                              type="text"
                              value={editFoodTypeName}
                              onChange={(e) => setEditFoodTypeName(e.target.value)}
                              className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                              autoFocus
                            />
                            <button
                              onClick={() => handleUpdateFoodType(foodType.id)}
                              className="px-3 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors"
                            >
                              Save
                            </button>
                            <button
                              onClick={() => {
                                setEditingFoodTypeId(null);
                                setEditFoodTypeName('');
                              }}
                              className="px-3 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors"
                            >
                              Cancel
                            </button>
                          </div>
                        ) : (
                          <>
                            <div className="flex items-center gap-3">
                              <div className="w-8 h-8 rounded-full bg-green-500/20 flex items-center justify-center">
                                <Utensils className="w-4 h-4 text-green-500" />
                              </div>
                              <span className="font-medium text-gray-900 dark:text-white">{foodType.name}</span>
                            </div>
                            <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                              <button
                                onClick={() => startEditFoodType(foodType)}
                                className="p-2 rounded-lg hover:bg-blue-500/20 transition-colors"
                              >
                                <Edit className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                              </button>
                              <button
                                onClick={() => handleDeleteFoodType(foodType.id)}
                                className="p-2 rounded-lg hover:bg-red-500/20 transition-colors"
                              >
                                <Trash2 className="w-4 h-4 text-red-500" />
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
          )}
        </div>
      </div>
    </div>
  );
}
