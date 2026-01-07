import { useState, useEffect } from 'react';
import { MapPin, X, Loader2, Navigation } from 'lucide-react';
import { Category, FoodType, RestaurantFilters, geocodeCities, GooglePlaceResult } from '../services/api';
import { AlertDialog } from './AlertDialog';

interface SearchFiltersProps {
  categories: Category[];
  foodTypes: FoodType[];
  filters: RestaurantFilters;
  onFiltersChange: (filters: RestaurantFilters) => void;
}

export function SearchFilters({ categories, foodTypes, filters, onFiltersChange }: SearchFiltersProps) {
  const [locationSearch, setLocationSearch] = useState('');
  const [locationResults, setLocationResults] = useState<GooglePlaceResult[]>([]);
  const [searchingLocation, setSearchingLocation] = useState(false);
  const [selectedLocation, setSelectedLocation] = useState<{ name: string; lat: number; lng: number } | null>(null);
  const [gettingCurrentLocation, setGettingCurrentLocation] = useState(false);
  const [alertMessage, setAlertMessage] = useState('');

  const hasActiveFilters = filters.category_id || (filters.food_type_ids && filters.food_type_ids.length > 0) || filters.radius || filters.price_range || filters.min_rating || filters.sort;

  // Search for locations
  useEffect(() => {
    if (locationSearch.length < 2) {
      setLocationResults([]);
      return;
    }

    const timer = setTimeout(async () => {
      setSearchingLocation(true);
      try {
        const results = await geocodeCities(locationSearch);
        setLocationResults(results);
      } catch (error) {
      } finally {
        setSearchingLocation(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [locationSearch]);

  const handleCategoryChange = (categoryId: number | undefined) => {
    onFiltersChange({ ...filters, category_id: categoryId });
  };

  const handleFoodTypeToggle = (foodTypeId: number) => {
    const currentIds = filters.food_type_ids || [];
    const newIds = currentIds.includes(foodTypeId)
      ? currentIds.filter(id => id !== foodTypeId)
      : [...currentIds, foodTypeId];
    onFiltersChange({ ...filters, food_type_ids: newIds.length > 0 ? newIds : undefined });
  };

  const handleRadiusChange = (radius: number | undefined) => {
    if (radius && selectedLocation) {
      onFiltersChange({
        ...filters,
        radius,
        lat: selectedLocation.lat,
        lng: selectedLocation.lng,
      });
    } else {
      onFiltersChange({
        ...filters,
        radius: undefined,
        lat: undefined,
        lng: undefined,
      });
    }
  };

  const handleLocationSelect = (place: GooglePlaceResult) => {
    setSelectedLocation({ name: place.name, lat: place.latitude, lng: place.longitude });
    setLocationSearch('');
    setLocationResults([]);
    if (filters.radius) {
      onFiltersChange({
        ...filters,
        lat: place.latitude,
        lng: place.longitude,
      });
    }
  };

  const handleUseCurrentLocation = () => {
    if (!navigator.geolocation) {
      setAlertMessage('Geolocation is not supported by your browser');
      return;
    }

    setGettingCurrentLocation(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        const { latitude, longitude } = position.coords;
        setSelectedLocation({ name: 'Current Location', lat: latitude, lng: longitude });
        if (filters.radius) {
          onFiltersChange({
            ...filters,
            lat: latitude,
            lng: longitude,
          });
        }
        setGettingCurrentLocation(false);
      },
      (_error) => {
        setAlertMessage('Unable to get your location. Please search for a location instead.');
        setGettingCurrentLocation(false);
      }
    );
  };

  const clearLocationFilter = () => {
    setSelectedLocation(null);
    onFiltersChange({
      ...filters,
      lat: undefined,
      lng: undefined,
      radius: undefined,
    });
  };

  const clearAllFilters = () => {
    setSelectedLocation(null);
    onFiltersChange({});
  };

  return (
    <div className="card-glass mb-6 p-6 animate-slide-down">
      {hasActiveFilters && (
        <div className="flex items-center justify-end mb-4">
          <button
            onClick={clearAllFilters}
            className="text-sm btn-glass-danger px-3 py-1"
          >
            Clear all
          </button>
        </div>
      )}

      <div className="space-y-4">
          {/* Category Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Category
            </label>
            <select
              value={filters.category_id || ''}
              onChange={(e) => handleCategoryChange(e.target.value ? Number(e.target.value) : undefined)}
              className="input-glass"
            >
              <option value="">All Categories</option>
              {categories.map((cat) => (
                <option key={cat.id} value={cat.id}>{cat.name}</option>
              ))}
            </select>
          </div>

          {/* Food Types Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Food Types
            </label>
            <div className="flex flex-wrap gap-2">
              {foodTypes.map((ft) => (
                <button
                  key={ft.id}
                  onClick={() => handleFoodTypeToggle(ft.id)}
                  className={`px-3 py-1.5 rounded-full text-sm font-medium transition-all duration-300 ${
                    filters.food_type_ids?.includes(ft.id)
                      ? 'badge-food-type shadow-lg scale-105'
                      : 'btn-glass'
                  }`}
                >
                  {ft.name}
                </button>
              ))}
            </div>
          </div>

          {/* Price Range Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Price Range
            </label>
            <div className="flex flex-wrap gap-2">
              {[
                { value: 1, label: '$' },
                { value: 2, label: '$$' },
                { value: 3, label: '$$$' },
                { value: 4, label: '$$$$' }
              ].map((price) => (
                <button
                  key={price.value}
                  onClick={() => onFiltersChange({ ...filters, price_range: filters.price_range === price.value ? undefined : price.value })}
                  className={`px-4 py-1.5 rounded-full text-sm font-medium transition-all duration-300 ${
                    filters.price_range === price.value
                      ? 'bg-gradient-to-r from-green-500/20 to-emerald-500/20 backdrop-blur-md border border-green-500/30 text-green-700 dark:text-green-300 shadow-lg'
                      : 'btn-glass'
                  }`}
                >
                  {price.label}
                </button>
              ))}
            </div>
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Shows restaurants up to and including selected price range</p>
          </div>

          {/* Rating Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Minimum Rating
            </label>
            <div className="flex flex-wrap gap-2">
              {[1, 2, 3, 4, 5].map((rating) => (
                <button
                  key={rating}
                  onClick={() => onFiltersChange({ ...filters, min_rating: filters.min_rating === rating ? undefined : rating })}
                  className={`px-4 py-1.5 rounded-full text-sm font-medium transition-all duration-300 ${
                    filters.min_rating === rating
                      ? 'bg-gradient-to-r from-yellow-400/30 to-orange-400/30 backdrop-blur-md border border-yellow-500/40 text-yellow-700 dark:text-yellow-200 shadow-lg'
                      : 'btn-glass'
                  }`}
                >
                  {rating}+ ⭐
                </button>
              ))}
            </div>
          </div>

          {/* Sort Options */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Sort By
            </label>
            <div className="flex flex-wrap gap-2">
              {[
                { value: 'date', label: 'Recently Added' },
                { value: 'name', label: 'Name (A-Z)' },
                { value: 'rating', label: 'Highest Rated' }
              ].map((sort) => (
                <button
                  key={sort.value}
                  onClick={() => onFiltersChange({ ...filters, sort: filters.sort === sort.value ? undefined : sort.value as any })}
                  className={`px-4 py-1.5 rounded-full text-sm font-medium transition-all duration-300 ${
                    filters.sort === sort.value
                      ? 'bg-gradient-to-r from-purple-500/20 to-pink-500/20 backdrop-blur-md border border-purple-500/30 text-purple-700 dark:text-purple-300 shadow-lg'
                      : 'btn-glass'
                  }`}
                >
                  {sort.label}
                </button>
              ))}
            </div>
          </div>

          {/* Location Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Location & Radius
            </label>

            {selectedLocation ? (
              <div className="flex items-center gap-2 mb-3 p-3 bg-blue-500/20 dark:bg-blue-500/30 backdrop-blur-md border border-blue-500/40 rounded-xl">
                <MapPin className="w-4 h-4 text-blue-500" />
                <span className="text-sm text-gray-700 dark:text-gray-300 flex-1 font-medium">
                  {selectedLocation.name}
                </span>
                <button
                  onClick={clearLocationFilter}
                  className="p-1 hover:bg-red-500/20 rounded-full transition-colors"
                >
                  <X className="w-4 h-4 text-gray-400 hover:text-red-500" />
                </button>
              </div>
            ) : (
              <div className="space-y-2 mb-3">
                <div className="relative">
                  <input
                    type="text"
                    value={locationSearch}
                    onChange={(e) => setLocationSearch(e.target.value)}
                    placeholder="Search city..."
                    className="input-glass pl-10"
                  />
                  <MapPin className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  {searchingLocation && (
                    <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 animate-spin" />
                  )}
                </div>

                {locationResults.length > 0 && (
                  <div className="bg-white/95 dark:bg-gray-800/95 backdrop-blur-2xl border border-white/30 dark:border-white/10 rounded-2xl shadow-2xl shadow-black/20 max-h-48 overflow-y-auto animate-slide-down">
                    {locationResults.map((place) => (
                      <button
                        key={place.place_id}
                        onClick={() => handleLocationSelect(place)}
                        className="w-full px-4 py-3 text-left hover:bg-white/40 dark:hover:bg-white/10 text-sm transition-all duration-200 first:rounded-t-2xl last:rounded-b-2xl"
                      >
                        <div className="font-medium text-gray-900 dark:text-white">{place.name}</div>
                        <div className="text-gray-500 dark:text-gray-400 text-xs">{place.address}</div>
                      </button>
                    ))}
                  </div>
                )}

                <button
                  onClick={handleUseCurrentLocation}
                  disabled={gettingCurrentLocation}
                  className="flex items-center gap-2 text-sm text-blue-500 hover:text-blue-600 disabled:opacity-50"
                >
                  {gettingCurrentLocation ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <Navigation className="w-4 h-4" />
                  )}
                  Use my current location
                </button>
              </div>
            )}

            {/* Radius Selection */}
            <div className="flex flex-wrap gap-2">
              {[1, 5, 10, 25, 50].map((km) => (
                <button
                  key={km}
                  onClick={() => handleRadiusChange(filters.radius === km ? undefined : km)}
                  disabled={!selectedLocation}
                  className={`px-4 py-1.5 rounded-full text-sm font-medium transition-all duration-300 ${
                    filters.radius === km
                      ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-blue-500/30 shadow-lg shadow-blue-500/20'
                      : 'btn-glass disabled:opacity-50 disabled:cursor-not-allowed'
                  }`}
                >
                  {km} km
                </button>
              ))}
            </div>
            {!selectedLocation && (
              <p className="text-xs text-gray-500 mt-1">Select a location first to filter by radius</p>
            )}
          </div>
      </div>
      <AlertDialog
        isOpen={alertMessage !== ''}
        onClose={() => setAlertMessage('')}
        message={alertMessage}
      />
    </div>
  );
}
