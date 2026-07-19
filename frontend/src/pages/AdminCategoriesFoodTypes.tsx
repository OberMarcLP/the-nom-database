import { useEffect, useState } from 'react';
import { Category, FoodType, getCategories, getFoodTypes, createCategory, updateCategory, deleteCategory, createFoodType, updateFoodType, deleteFoodType } from '../services/api';
import { Tag, Utensils, Plus, Edit2, Trash2, X, Save } from 'lucide-react';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { useToast } from '../hooks/useToast';

export function AdminCategoriesFoodTypes() {
  const { showError } = useToast();

  // Categories state
  const [categories, setCategories] = useState<Category[]>([]);
  const [newCategoryName, setNewCategoryName] = useState('');
  const [editingCategoryId, setEditingCategoryId] = useState<number | null>(null);
  const [editCategoryName, setEditCategoryName] = useState('');
  const [confirmDeleteCategory, setConfirmDeleteCategory] = useState<{ id: number; name: string } | null>(null);

  // Food Types state
  const [foodTypes, setFoodTypes] = useState<FoodType[]>([]);
  const [newFoodTypeName, setNewFoodTypeName] = useState('');
  const [editingFoodTypeId, setEditingFoodTypeId] = useState<number | null>(null);
  const [editFoodTypeName, setEditFoodTypeName] = useState('');
  const [confirmDeleteFoodType, setConfirmDeleteFoodType] = useState<{ id: number; name: string } | null>(null);

  useEffect(() => {
    loadCategories();
    loadFoodTypes();
  }, []);

  // Handle ESC key to close modals and cancel editing
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (confirmDeleteCategory) {
          setConfirmDeleteCategory(null);
        } else if (confirmDeleteFoodType) {
          setConfirmDeleteFoodType(null);
        } else if (editingCategoryId !== null) {
          setEditingCategoryId(null);
          setEditCategoryName('');
        } else if (editingFoodTypeId !== null) {
          setEditingFoodTypeId(null);
          setEditFoodTypeName('');
        }
      }
    };

    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [confirmDeleteCategory, confirmDeleteFoodType, editingCategoryId, editingFoodTypeId]);

  const loadCategories = async () => {
    try {
      const cats = await getCategories();
      setCategories(cats);
    } catch (error) {
      console.error('Failed to load categories:', error);
    }
  };

  const loadFoodTypes = async () => {
    try {
      const fts = await getFoodTypes();
      setFoodTypes(fts);
    } catch (error) {
      console.error('Failed to load food types:', error);
    }
  };

  // Category handlers
  const handleCreateCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCategoryName.trim()) return;
    try {
      await createCategory(newCategoryName);
      setNewCategoryName('');
      loadCategories();
    } catch (error) {
      console.error('Failed to create category:', error);
      showError('Failed to create category');
    }
  };

  const handleUpdateCategory = async (id: number) => {
    if (!editCategoryName.trim()) return;
    try {
      await updateCategory(id, editCategoryName);
      setEditingCategoryId(null);
      setEditCategoryName('');
      loadCategories();
    } catch (error) {
      console.error('Failed to update category:', error);
      showError('Failed to update category');
    }
  };

  const handleDeleteCategory = async () => {
    if (!confirmDeleteCategory) return;
    try {
      await deleteCategory(confirmDeleteCategory.id);
      loadCategories();
      setConfirmDeleteCategory(null);
    } catch (error) {
      console.error('Failed to delete category:', error);
      showError('Failed to delete category');
      setConfirmDeleteCategory(null);
    }
  };

  // Food Type handlers
  const handleCreateFoodType = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newFoodTypeName.trim()) return;
    try {
      await createFoodType(newFoodTypeName);
      setNewFoodTypeName('');
      loadFoodTypes();
    } catch (error) {
      console.error('Failed to create food type:', error);
      showError('Failed to create food type');
    }
  };

  const handleUpdateFoodType = async (id: number) => {
    if (!editFoodTypeName.trim()) return;
    try {
      await updateFoodType(id, editFoodTypeName);
      setEditingFoodTypeId(null);
      setEditFoodTypeName('');
      loadFoodTypes();
    } catch (error) {
      console.error('Failed to update food type:', error);
      showError('Failed to update food type');
    }
  };

  const handleDeleteFoodType = async () => {
    if (!confirmDeleteFoodType) return;
    try {
      await deleteFoodType(confirmDeleteFoodType.id);
      loadFoodTypes();
      setConfirmDeleteFoodType(null);
    } catch (error) {
      console.error('Failed to delete food type:', error);
      showError('Failed to delete food type');
      setConfirmDeleteFoodType(null);
    }
  };

  return (
    <div>
      <div className="admin-page-header">
        <h1 className="admin-page-title">Categories & Food Types</h1>
        <p className="admin-page-description">
          Manage restaurant categories and food type classifications
        </p>
      </div>

      {/* Categories Section */}
      <div className="admin-card" style={{ marginBottom: '20px' }}>
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <Tag size={20} />
            Categories
          </h2>
        </div>
        <div style={{ padding: '20px', display: 'grid', gap: '20px' }}>
          <form onSubmit={handleCreateCategory} style={{ display: 'flex', gap: '12px' }}>
            <input
              type="text"
              value={newCategoryName}
              onChange={(e) => setNewCategoryName(e.target.value)}
              placeholder="New category name..."
              className="admin-input"
              style={{ flex: 1 }}
            />
            <button type="submit" className="admin-btn admin-btn-primary">
              <Plus size={16} />
              Add
            </button>
          </form>

          {categories.length === 0 ? (
            <div className="admin-empty" style={{ padding: '40px 20px' }}>
              <Tag size={48} style={{ opacity: 0.3, marginBottom: '16px' }} />
              <p className="admin-empty-text">No categories yet. Add your first category above.</p>
            </div>
          ) : (
            <div style={{ display: 'grid', gap: '8px', maxHeight: '300px', overflowY: 'auto' }}>
              {categories.map((category) => (
                <div key={category.id} style={{
                  padding: '12px 16px',
                  background: 'var(--admin-bg)',
                  border: '1px solid var(--admin-border)',
                  borderRadius: '4px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  transition: 'all 0.2s'
                }}>
                  {editingCategoryId === category.id ? (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flex: 1 }}>
                      <input
                        type="text"
                        value={editCategoryName}
                        onChange={(e) => setEditCategoryName(e.target.value)}
                        className="admin-input"
                        style={{ flex: 1, marginBottom: 0 }}
                        autoFocus
                      />
                      <button
                        onClick={() => handleUpdateCategory(category.id)}
                        className="admin-btn-icon"
                        title="Save"
                      >
                        <Save size={14} />
                      </button>
                      <button
                        onClick={() => {
                          setEditingCategoryId(null);
                          setEditCategoryName('');
                        }}
                        className="admin-btn-icon"
                        title="Cancel"
                      >
                        <X size={14} />
                      </button>
                    </div>
                  ) : (
                    <>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <div style={{
                          width: '32px',
                          height: '32px',
                          borderRadius: '50%',
                          background: 'rgba(0, 170, 255, 0.1)',
                          border: '1px solid var(--admin-info)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center'
                        }}>
                          <Tag size={14} style={{ color: 'var(--admin-info)' }} />
                        </div>
                        <span style={{ fontWeight: 500 }}>{category.name}</span>
                      </div>
                      <div style={{ display: 'flex', gap: '4px' }}>
                        <button
                          onClick={() => {
                            setEditingCategoryId(category.id);
                            setEditCategoryName(category.name);
                          }}
                          className="admin-btn-icon"
                          title="Edit category"
                        >
                          <Edit2 size={14} />
                        </button>
                        <button
                          onClick={() => setConfirmDeleteCategory({ id: category.id, name: category.name })}
                          className="admin-btn-icon admin-btn-danger"
                          title="Delete category"
                        >
                          <Trash2 size={14} />
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

      {/* Food Types Section */}
      <div className="admin-card" style={{ marginBottom: '20px' }}>
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <Utensils size={20} />
            Food Types
          </h2>
        </div>
        <div style={{ padding: '20px', display: 'grid', gap: '20px' }}>
          <form onSubmit={handleCreateFoodType} style={{ display: 'flex', gap: '12px' }}>
            <input
              type="text"
              value={newFoodTypeName}
              onChange={(e) => setNewFoodTypeName(e.target.value)}
              placeholder="New food type name..."
              className="admin-input"
              style={{ flex: 1 }}
            />
            <button type="submit" className="admin-btn admin-btn-primary">
              <Plus size={16} />
              Add
            </button>
          </form>

          {foodTypes.length === 0 ? (
            <div className="admin-empty" style={{ padding: '40px 20px' }}>
              <Utensils size={48} style={{ opacity: 0.3, marginBottom: '16px' }} />
              <p className="admin-empty-text">No food types yet. Add your first food type above.</p>
            </div>
          ) : (
            <div style={{ display: 'grid', gap: '8px', maxHeight: '300px', overflowY: 'auto' }}>
              {foodTypes.map((foodType) => (
                <div key={foodType.id} style={{
                  padding: '12px 16px',
                  background: 'var(--admin-bg)',
                  border: '1px solid var(--admin-border)',
                  borderRadius: '4px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  transition: 'all 0.2s'
                }}>
                  {editingFoodTypeId === foodType.id ? (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flex: 1 }}>
                      <input
                        type="text"
                        value={editFoodTypeName}
                        onChange={(e) => setEditFoodTypeName(e.target.value)}
                        className="admin-input"
                        style={{ flex: 1, marginBottom: 0 }}
                        autoFocus
                      />
                      <button
                        onClick={() => handleUpdateFoodType(foodType.id)}
                        className="admin-btn-icon"
                        title="Save"
                      >
                        <Save size={14} />
                      </button>
                      <button
                        onClick={() => {
                          setEditingFoodTypeId(null);
                          setEditFoodTypeName('');
                        }}
                        className="admin-btn-icon"
                        title="Cancel"
                      >
                        <X size={14} />
                      </button>
                    </div>
                  ) : (
                    <>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <div style={{
                          width: '32px',
                          height: '32px',
                          borderRadius: '50%',
                          background: 'var(--admin-accent-dim)',
                          border: '1px solid var(--admin-accent)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center'
                        }}>
                          <Utensils size={14} style={{ color: 'var(--admin-accent)' }} />
                        </div>
                        <span style={{ fontWeight: 500 }}>{foodType.name}</span>
                      </div>
                      <div style={{ display: 'flex', gap: '4px' }}>
                        <button
                          onClick={() => {
                            setEditingFoodTypeId(foodType.id);
                            setEditFoodTypeName(foodType.name);
                          }}
                          className="admin-btn-icon"
                          title="Edit food type"
                        >
                          <Edit2 size={14} />
                        </button>
                        <button
                          onClick={() => setConfirmDeleteFoodType({ id: foodType.id, name: foodType.name })}
                          className="admin-btn-icon admin-btn-danger"
                          title="Delete food type"
                        >
                          <Trash2 size={14} />
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

      {/* Confirm Delete Category Dialog */}
      <ConfirmDialog
        isOpen={!!confirmDeleteCategory}
        onClose={() => setConfirmDeleteCategory(null)}
        onConfirm={handleDeleteCategory}
        title="localhost:3000"
        message={`Are you sure you want to delete the category "${confirmDeleteCategory?.name}"?`}
        confirmText="OK"
        cancelText="Cancel"
        isDangerous={true}
      />

      {/* Confirm Delete Food Type Dialog */}
      <ConfirmDialog
        isOpen={!!confirmDeleteFoodType}
        onClose={() => setConfirmDeleteFoodType(null)}
        onConfirm={handleDeleteFoodType}
        title="localhost:3000"
        message={`Are you sure you want to delete the food type "${confirmDeleteFoodType?.name}"?`}
        confirmText="OK"
        cancelText="Cancel"
        isDangerous={true}
      />
    </div>
  );
}
