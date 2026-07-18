import { useState, useEffect, useRef } from 'react';
import { Search, X, Lightbulb, Filter } from 'lucide-react';
import { globalSearch, Restaurant, Category, FoodType, RestaurantFilters } from '../services/api';
import { useNavigate } from 'react-router-dom';
import { SearchFilters } from './SearchFilters';

interface GlobalSearchProps {
  categories: Category[];
  foodTypes: FoodType[];
  filters: RestaurantFilters;
  onFiltersChange: (filters: RestaurantFilters) => void;
}

export function GlobalSearch({ categories, foodTypes, filters, onFiltersChange }: GlobalSearchProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Restaurant[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [showFilters, setShowFilters] = useState(false);
  const searchRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  useEffect(() => {
    const searchDebounce = setTimeout(async () => {
      if (query.trim().length >= 2) {
        setIsLoading(true);
        try {
          const data = await globalSearch(query);
          setResults(data);
          setIsOpen(true);
        } catch (error) {
          setResults([]);
        } finally {
          setIsLoading(false);
        }
      } else {
        setResults([]);
        setIsOpen(false);
      }
    }, 300);

    return () => clearTimeout(searchDebounce);
  }, [query]);

  const handleResultClick = (restaurant: Restaurant) => {
    if (restaurant.is_suggestion) {
      navigate(`/`);
    } else {
      navigate(`/restaurants/${restaurant.id}`);
    }
    setIsOpen(false);
    setQuery('');
  };

  const handleClear = () => {
    setQuery('');
    setResults([]);
    setIsOpen(false);
  };

  const hasActiveFilters = filters.category_id || (filters.food_type_ids && filters.food_type_ids.length > 0) || filters.radius;

  return (
    <div ref={searchRef} className="relative w-full">
      <div className="flex gap-2 items-center mb-2">
        <div className="relative flex-1 flex items-center gap-2 rounded-md border-2 border-(--border) bg-(--bg) focus-within:border-(--accent) focus-within:shadow-[0_0_0_3px_var(--accent-dim)] transition-all">
          <Search className="w-5 h-5 shrink-0 ml-3 text-(--text-muted)" aria-hidden />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onFocus={() => query.length >= 2 && setIsOpen(true)}
            placeholder="Search restaurants and suggestions..."
            className="input-glass flex-1 min-w-0 border-0! p-3! focus:shadow-none!"
          />
          {query && (
            <button
              type="button"
              onClick={handleClear}
              className="shrink-0 p-2 mr-1 rounded-sm text-(--text-muted) hover:text-(--text) transition-colors"
              aria-label="Clear search"
            >
              <X className="w-5 h-5" />
            </button>
          )}
        </div>
        <button
          type="button"
          onClick={() => setShowFilters(!showFilters)}
          className={`relative p-2 rounded-md transition-all duration-200 border-2 ${
            showFilters
              ? 'border-(--accent) bg-(--accent-dim) text-(--accent)'
              : 'btn-glass'
          }`}
        >
          <Filter className="w-5 h-5" />
          {hasActiveFilters && (
            <span className="absolute top-0 right-0 w-2 h-2 rounded-full bg-(--accent) animate-pulse" />
          )}
        </button>
      </div>

      {showFilters && (
        <SearchFilters
          categories={categories}
          foodTypes={foodTypes}
          filters={filters}
          onFiltersChange={onFiltersChange}
        />
      )}

      {isOpen && (
        <div className="absolute z-50 w-full mt-3 card-glass rounded-lg border-2 border-(--border) max-h-96 overflow-y-auto">
          {isLoading ? (
            <div className="p-4 text-center text-(--text-muted)">
              Searching...
            </div>
          ) : results.length > 0 ? (
            <ul className="divide-y divide-(--border)">
              {results.map((restaurant, index) => (
                <li
                  key={`${restaurant.is_suggestion ? 's' : 'r'}-${restaurant.is_suggestion ? restaurant.suggestion_id : restaurant.id}-${index}`}
                  onClick={() => handleResultClick(restaurant)}
                  className="p-4 cursor-pointer transition-colors hover:bg-(--surface-hover)"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <h3 className="font-semibold text-(--text) truncate">
                          {restaurant.name}
                        </h3>
                        {restaurant.is_suggestion && (
                          <span className="badge-suggestion shrink-0">
                            <Lightbulb className="w-3 h-3" />
                            Suggestion
                          </span>
                        )}
                      </div>
                      {restaurant.address && (
                        <p className="text-sm text-(--text-muted) truncate mt-1">
                          {restaurant.address}
                        </p>
                      )}
                      {restaurant.category && (
                        <p className="text-xs text-(--text-muted) mt-1">
                          {restaurant.category.name}
                        </p>
                      )}
                    </div>
                    {!restaurant.is_suggestion && restaurant.avg_rating && (
                      <div className="text-right shrink-0">
                        <div className="text-sm font-semibold text-(--text)">
                          ★ {restaurant.avg_rating.overall.toFixed(1)}
                        </div>
                        <div className="text-xs text-(--text-muted)">
                          {restaurant.avg_rating.count} reviews
                        </div>
                      </div>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          ) : query.length >= 2 ? (
            <div className="p-4 text-center text-(--text-muted)">
              No results found for &quot;{query}&quot;
            </div>
          ) : null}
        </div>
      )}
    </div>
  );
}
