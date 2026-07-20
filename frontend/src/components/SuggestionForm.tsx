import { useState } from 'react';
import { GooglePlaceResult, CreateSuggestionData, getPlaceDetails } from '../services/api';
import { PlaceSearch } from './PlaceSearch';
import { X, Loader2 } from 'lucide-react';
import { useCategories, useFoodTypes } from '../hooks/useApi';

interface SuggestionFormProps {
  onSubmit: (data: CreateSuggestionData) => void;
  onCancel: () => void;
}

export function SuggestionForm({ onSubmit, onCancel }: SuggestionFormProps) {
  const { data: categories = [] } = useCategories();
  const { data: foodTypes = [] } = useFoodTypes();
  const [loadingDetails, setLoadingDetails] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    address: '',
    phone: '',
    website: '',
    latitude: null as number | null,
    longitude: null as number | null,
    google_place_id: '',
    suggested_category_id: null as number | null,
    food_type_ids: [] as number[],
    notes: '',
  });

  const handlePlaceSelect = async (place: GooglePlaceResult) => {
    // First set basic info from search results
    setFormData(prev => ({
      ...prev,
      name: place.name,
      address: place.address,
      latitude: place.latitude,
      longitude: place.longitude,
      google_place_id: place.place_id,
    }));

    // Then fetch additional details (phone, website)
    setLoadingDetails(true);
    try {
      const details = await getPlaceDetails(place.place_id);
      setFormData(prev => ({
        ...prev,
        phone: details.phone || '',
        website: details.website || '',
      }));
    } catch {
      // intentionally ignored - user-facing error handling tracked in review backlog
    } finally {
      setLoadingDetails(false);
    }
  };

  const handleFoodTypeToggle = (ftId: number) => {
    setFormData(prev => ({
      ...prev,
      food_type_ids: prev.food_type_ids.includes(ftId)
        ? prev.food_type_ids.filter(id => id !== ftId)
        : [...prev.food_type_ids, ftId],
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      name: formData.name,
      address: formData.address || null,
      phone: formData.phone || null,
      website: formData.website || null,
      latitude: formData.latitude,
      longitude: formData.longitude,
      google_place_id: formData.google_place_id || null,
      suggested_category_id: formData.suggested_category_id,
      food_type_ids: formData.food_type_ids,
      notes: formData.notes || null,
    });
  };

  const selectedFoodTypes = foodTypes.filter(ft => formData.food_type_ids.includes(ft.id));

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="admin-label">Search Google Maps</label>
        <PlaceSearch onSelect={handlePlaceSelect} />
        {loadingDetails && (
          <div className="flex items-center gap-2 mt-2 text-sm text-(--text-muted)">
            <Loader2 className="w-4 h-4 animate-spin" />
            Fetching restaurant details...
          </div>
        )}
      </div>

      <div>
        <label className="admin-label">Name *</label>
        <input
          type="text"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          className="admin-input"
          required
        />
      </div>

      <div>
        <label className="admin-label">Address</label>
        <input
          type="text"
          value={formData.address}
          onChange={(e) => setFormData({ ...formData, address: e.target.value })}
          className="admin-input"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="admin-label">Phone</label>
          <input
            type="tel"
            value={formData.phone}
            onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
            className="admin-input"
            placeholder="+1 234 567 8900"
          />
        </div>
        <div>
          <label className="admin-label">Website</label>
          <input
            type="url"
            value={formData.website}
            onChange={(e) => setFormData({ ...formData, website: e.target.value })}
            className="admin-input"
            placeholder="https://..."
          />
        </div>
      </div>

      <div>
        <label className="admin-label">Suggested Category</label>
        <select
          value={formData.suggested_category_id || ''}
          onChange={(e) =>
            setFormData({
              ...formData,
              suggested_category_id: e.target.value ? parseInt(e.target.value) : null,
            })
          }
          className="admin-select"
        >
          <option value="">Select category</option>
          {categories.map((cat) => (
            <option key={cat.id} value={cat.id}>
              {cat.name}
            </option>
          ))}
        </select>
      </div>

      <div>
        <label className="admin-label">Food Types (select multiple)</label>

        {selectedFoodTypes.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-3">
            {selectedFoodTypes.map(ft => (
              <span
                key={ft.id}
                className="inline-flex items-center gap-1 px-3 py-1 bg-(--surface-hover) text-(--success) rounded-full text-sm"
              >
                {ft.name}
                <button
                  type="button"
                  onClick={() => handleFoodTypeToggle(ft.id)}
                  className="hover:bg-(--surface-hover) rounded-full p-0.5"
                >
                  <X className="w-3 h-3" />
                </button>
              </span>
            ))}
          </div>
        )}

        <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 max-h-48 overflow-y-auto p-2 border border-(--border) rounded-lg">
          {foodTypes.map((ft) => (
            <label
              key={ft.id}
              className={`flex items-center gap-2 p-2 rounded cursor-pointer transition-colors ${
                formData.food_type_ids.includes(ft.id)
                  ? 'bg-(--accent-dim)'
                  : 'hover:bg-(--surface-hover)'
              }`}
            >
              <input
                type="checkbox"
                checked={formData.food_type_ids.includes(ft.id)}
                onChange={() => handleFoodTypeToggle(ft.id)}
                className="w-4 h-4 text-(--success) rounded-sm focus:ring-(--success)"
              />
              <span className="text-sm">{ft.name}</span>
            </label>
          ))}
        </div>
      </div>

      <div>
        <label className="admin-label">Notes</label>
        <textarea
          value={formData.notes}
          onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
          className="admin-textarea min-h-[100px]"
          rows={3}
          placeholder="Any additional notes about this restaurant..."
        />
      </div>

      {formData.latitude && formData.longitude && (
        <div className="text-sm text-(--text-muted)">
          Location: {formData.latitude.toFixed(6)}, {formData.longitude.toFixed(6)}
        </div>
      )}

      <div className="flex gap-3 pt-4">
        <button type="submit" className="btn-glass-primary flex-1">
          Submit Suggestion
        </button>
        <button type="button" onClick={onCancel} className="btn-glass">
          Cancel
        </button>
      </div>
    </form>
  );
}
