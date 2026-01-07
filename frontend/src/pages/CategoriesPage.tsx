import { useState, useEffect } from 'react';
import { Plus, Edit, Trash2, Loader2, Tag } from 'lucide-react';
import { Category, getCategories, createCategory, updateCategory, deleteCategory } from '../services/api';
import { PermissionGuard } from '../components/PermissionGuard';

export function CategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [newName, setNewName] = useState('');
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState('');

  const fetchCategories = async () => {
    try {
      const data = await getCategories();
      setCategories(data);
    } catch (error) {
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCategories();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim()) return;
    try {
      await createCategory(newName);
      setNewName('');
      fetchCategories();
    } catch (error) {
    }
  };

  const handleUpdate = async (id: number) => {
    if (!editName.trim()) return;
    try {
      await updateCategory(id, editName);
      setEditingId(null);
      setEditName('');
      fetchCategories();
    } catch (error) {
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this category?')) return;
    try {
      await deleteCategory(id);
      fetchCategories();
    } catch (error) {
    }
  };

  const startEdit = (category: Category) => {
    setEditingId(category.id);
    setEditName(category.name);
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
        <div className="p-3 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 backdrop-blur-sm">
          <Tag className="w-6 h-6 text-blue-500" />
        </div>
        <h1 className="text-3xl font-bold text-gradient">Categories</h1>
      </div>

      <PermissionGuard permissions={['categories.create', 'categories.update', 'categories.delete']}>
        <form onSubmit={handleCreate} className="flex gap-2 mb-6">
          <input
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="New category name..."
            className="input-glass flex-1"
          />
          <button type="submit" className="btn-glass-primary flex items-center gap-2">
            <Plus className="w-5 h-5" />
            Add
          </button>
        </form>
      </PermissionGuard>

      {categories.length === 0 ? (
        <p className="text-center text-gray-500 dark:text-gray-400 py-8">
          No categories yet. Add your first category above.
        </p>
      ) : (
        <div className="space-y-2">
          {categories.map((category) => (
            <div key={category.id} className="card-glass flex items-center justify-between">
              {editingId === category.id ? (
                <div className="flex items-center gap-2 flex-1">
                  <input
                    type="text"
                    value={editName}
                    onChange={(e) => setEditName(e.target.value)}
                    className="input-glass flex-1"
                    autoFocus
                  />
                  <button
                    onClick={() => handleUpdate(category.id)}
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
                  <span className="font-medium">{category.name}</span>
                  <PermissionGuard permissions={['categories.create', 'categories.update', 'categories.delete']}>
                    <div className="flex gap-2">
                      <button
                        onClick={() => startEdit(category)}
                        className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
                      >
                        <Edit className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => handleDelete(category.id)}
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
