import { useState } from 'react';
import { Search, Loader2 } from 'lucide-react';
import { searchPlaces, GooglePlaceResult } from '../services/api';

interface PlaceSearchProps {
  onSelect: (place: GooglePlaceResult) => void;
}

export function PlaceSearch({ onSelect }: PlaceSearchProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<GooglePlaceResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSearch = async () => {
    if (!query.trim()) return;

    setLoading(true);
    setError(null);
    try {
      const places = await searchPlaces(query);
      setResults(places);
      if (places.length === 0) {
        setError('No restaurants found. Try a different search.');
      }
    } catch (err) {
      setError('Failed to search. Please try again.');
      setResults([]);
    } finally {
      setLoading(false);
    }
  };

  const handleSelect = (place: GooglePlaceResult) => {
    onSelect(place);
    setResults([]);
    setQuery('');
  };

  return (
    <div className="relative">
      <div className="flex gap-2">
        <div className="relative flex-1">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            placeholder="Search restaurant name..."
            className="admin-input !pl-12"
          />
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
        </div>
        <button
          onClick={handleSearch}
          disabled={loading || !query.trim()}
          className="admin-btn-primary flex items-center gap-2"
        >
          {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Search'}
        </button>
      </div>

      {error && <p className="text-red-500 text-sm mt-2">{error}</p>}

      {results.length > 0 && (
        <div className="absolute z-10 w-full mt-2 rounded-lg shadow-lg border-2 max-h-80 overflow-y-auto" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
          {results.map((place) => (
            <button
              key={place.place_id}
              onClick={() => handleSelect(place)}
              className="w-full text-left px-4 py-3 transition-colors border-b-2 last:border-0"
              style={{
                borderColor: 'var(--border)',
                color: 'var(--text)'
              }}
              onMouseEnter={(e) => e.currentTarget.style.background = 'var(--surface-hover)'}
              onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
            >
              <p className="font-medium">{place.name}</p>
              <p className="text-sm" style={{ color: 'var(--text-muted)' }}>{place.address}</p>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
