import { useState, useEffect } from 'react';
import { Restaurant, Category, FoodType, GooglePlaceResult, CreateRestaurantData, getCategories, getFoodTypes, getPlaceDetails } from '../services/api';
import { PlaceSearch } from './PlaceSearch';
import { X, Loader2 } from 'lucide-react';

interface RestaurantFormProps {
  restaurant?: Restaurant;
  onSubmit: (data: CreateRestaurantData) => void;
  onCancel: () => void;
}

export function RestaurantForm({ restaurant, onSubmit, onCancel }: RestaurantFormProps) {
  const [categories, setCategories] = useState<Category[]>([]);
  const [foodTypes, setFoodTypes] = useState<FoodType[]>([]);
  const [loadingDetails, setLoadingDetails] = useState(false);
  const [formData, setFormData] = useState({
    name: restaurant?.name || '',
    description: restaurant?.description || '',
    address: restaurant?.address || '',
    phone: restaurant?.phone || '',
    website: restaurant?.website || '',
    latitude: restaurant?.latitude || null,
    longitude: restaurant?.longitude || null,
    google_place_id: restaurant?.google_place_id || '',
    category_id: restaurant?.category_id || null,
    food_type_ids: restaurant?.food_types?.map(ft => ft.id) || [] as number[],
  });

  useEffect(() => {
    const fetchData = async () => {
      const [cats, fts] = await Promise.all([getCategories(), getFoodTypes()]);
      setCategories(cats);
      setFoodTypes(fts);
    };
    fetchData();
  }, []);

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
      description: formData.description || null,
      address: formData.address || null,
      phone: formData.phone || null,
      website: formData.website || null,
      latitude: formData.latitude,
      longitude: formData.longitude,
      google_place_id: formData.google_place_id || null,
      category_id: formData.category_id,
      food_type_ids: formData.food_type_ids,
    });
  };

  const selectedFoodTypes = foodTypes.filter(ft => formData.food_type_ids.includes(ft.id));

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {!restaurant && (
        <div>
          <label className="label">Search Google Maps</label>
          <PlaceSearch onSelect={handlePlaceSelect} />
          {loadingDetails && (
            <div className="flex items-center gap-2 mt-2 text-sm text-gray-500">
              <Loader2 className="w-4 h-4 animate-spin" />
              Fetching restaurant details...
            </div>
          )}
        </div>
      )}

      <div>
        <label className="label">Name *</label>
        <input
          type="text"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          className="input-glass"
          required
        />
      </div>

      <div>
        <label className="label">Description</label>
        <textarea
          value={formData.description}
          onChange={(e) => setFormData({ ...formData, description: e.target.value })}
          className="input-glass min-h-[100px]"
          rows={3}
        />
      </div>

      <div>
        <label className="label">Address</label>
        <input
          type="text"
          value={formData.address}
          onChange={(e) => setFormData({ ...formData, address: e.target.value })}
          className="input-glass"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="label">Phone</label>
          <input
            type="tel"
            value={formData.phone}
            onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
            className="input-glass"
            placeholder="+1 234 567 8900"
          />
        </div>
        <div>
          <label className="label">Website</label>
          <input
            type="url"
            value={formData.website}
            onChange={(e) => setFormData({ ...formData, website: e.target.value })}
            className="input-glass"
            placeholder="https://..."
          />
        </div>
      </div>

      <div>
        <label className="label">Category</label>
        <select
          value={formData.category_id || ''}
          onChange={(e) =>
            setFormData({
              ...formData,
              category_id: e.target.value ? parseInt(e.target.value) : null,
            })
          }
          className="input-glass"
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
        <label className="label">Food Types (select multiple)</label>

        {selectedFoodTypes.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-3">
            {selectedFoodTypes.map(ft => (
              <span
                key={ft.id}
                className="badge-food-type"
              >
                {ft.name}
                <button
                  type="button"
                  onClick={() => handleFoodTypeToggle(ft.id)}
                  className="hover:bg-green-200 dark:hover:bg-green-800 rounded-full p-0.5"
                >
                  <X className="w-3 h-3" />
                </button>
              </span>
            ))}
          </div>
        )}

        <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 max-h-48 overflow-y-auto p-3 border border-white/30 dark:border-white/10 rounded-xl bg-white/20 dark:bg-gray-700/20 backdrop-blur-sm">
          {foodTypes.map((ft) => (
            <label
              key={ft.id}
              className={`flex items-center gap-2 p-2 rounded-lg cursor-pointer transition-all duration-200 ${
                formData.food_type_ids.includes(ft.id)
                  ? 'bg-green-500/20 dark:bg-green-500/30 border border-green-500/40'
                  : 'hover:bg-white/40 dark:hover:bg-white/10 border border-transparent'
              }`}
            >
              <input
                type="checkbox"
                checked={formData.food_type_ids.includes(ft.id)}
                onChange={() => handleFoodTypeToggle(ft.id)}
                className="w-4 h-4 text-green-600 rounded-sm focus:ring-green-500"
              />
              <span className="text-sm">{ft.name}</span>
            </label>
          ))}
        </div>
      </div>

      {formData.latitude && formData.longitude && (
        <div className="text-sm text-gray-500 dark:text-gray-400">
          Location: {formData.latitude.toFixed(6)}, {formData.longitude.toFixed(6)}
        </div>
      )}

      <div className="flex gap-3 pt-4">
        <button type="submit" className="btn-glass-primary flex-1">
          {restaurant ? 'Update Restaurant' : 'Add Restaurant'}
        </button>
        <button type="button" onClick={onCancel} className="btn-glass">
          Cancel
        </button>
      </div>
    </form>
  );
}
