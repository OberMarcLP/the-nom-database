import { useEffect, useState } from 'react';
import { api } from '../services/api';
import { Utensils, Eye, Trash2, Search } from 'lucide-react';

interface Restaurant {
  id: number;
  name: string;
  address?: string;
  description?: string;
  category?: {
    id: number;
    name: string;
  };
  avg_rating?: {
    food: number;
    service: number;
    ambiance: number;
  };
  created_at: string;
}

interface Stats {
  total: number;
  with_ratings: number;
  without_ratings: number;
}

export function AdminRestaurants() {
  const [restaurants, setRestaurants] = useState<Restaurant[]>([]);
  const [stats, setStats] = useState<Stats>({ total: 0, with_ratings: 0, without_ratings: 0 });
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    loadRestaurants();
  }, []);

  const loadRestaurants = async () => {
    try {
      setLoading(true);
      const response = await api.get<Restaurant[]>('/restaurants');
      setRestaurants(response);

      // Calculate stats
      const withRatings = response.filter(r => r.avg_rating).length;
      setStats({
        total: response.length,
        with_ratings: withRatings,
        without_ratings: response.length - withRatings,
      });
    } catch (error) {
      console.error('Failed to load restaurants:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (restaurantId: number, name: string) => {
    if (!confirm(`Delete "${name}"? This will also delete all associated ratings and photos. This action cannot be undone.`)) return;

    try {
      await api.delete(`/admin/restaurants/${restaurantId}`);
      loadRestaurants();
    } catch (error) {
      console.error('Failed to delete restaurant:', error);
      alert('Failed to delete restaurant');
    }
  };

  const getAverageRating = (restaurant: Restaurant) => {
    if (!restaurant.avg_rating) return null;
    const avg = (restaurant.avg_rating.food + restaurant.avg_rating.service + restaurant.avg_rating.ambiance) / 3;
    return avg.toFixed(1);
  };

  const filteredRestaurants = restaurants.filter(restaurant =>
    restaurant.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    restaurant.address?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    restaurant.category?.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div>
      <div className="admin-page-header">
        <h1 className="admin-page-title">Restaurant Management</h1>
        <p className="admin-page-description">
          Manage existing restaurants and their information
        </p>
      </div>

      {/* Stats Cards */}
      <div className="admin-stats-grid">
        <div className="admin-stat-card">
          <div className="admin-stat-label">Total Restaurants</div>
          <div className="admin-stat-value">{stats.total}</div>
        </div>
        <div className="admin-stat-card">
          <div className="admin-stat-label">With Ratings</div>
          <div className="admin-stat-value">{stats.with_ratings}</div>
        </div>
        <div className="admin-stat-card">
          <div className="admin-stat-label">Without Ratings</div>
          <div className="admin-stat-value">{stats.without_ratings}</div>
        </div>
        <div className="admin-stat-card">
          <div className="admin-stat-label">Rating Coverage</div>
          <div className="admin-stat-value">
            {stats.total > 0 ? Math.round((stats.with_ratings / stats.total) * 100) : 0}%
          </div>
        </div>
      </div>

      {/* Search */}
      <div className="admin-card" style={{ marginBottom: '20px' }}>
        <div style={{ padding: '20px' }}>
          <div style={{ position: 'relative' }}>
            <Search
              size={18}
              style={{
                position: 'absolute',
                left: '12px',
                top: '50%',
                transform: 'translateY(-50%)',
                color: 'var(--admin-text-muted)',
              }}
            />
            <input
              type="text"
              placeholder="Search by name, address, or category..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="admin-input"
              style={{ paddingLeft: '40px' }}
            />
          </div>
        </div>
      </div>

      {/* Restaurants Table */}
      <div className="admin-card">
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <Utensils size={20} />
            Restaurants ({filteredRestaurants.length})
          </h2>
        </div>

        {loading ? (
          <div className="admin-loading">
            <div className="admin-spinner" />
          </div>
        ) : filteredRestaurants.length === 0 ? (
          <div className="admin-empty">
            <p className="admin-empty-text">No restaurants found</p>
          </div>
        ) : (
          <div className="admin-table-container">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Restaurant</th>
                  <th>Category</th>
                  <th>Average Rating</th>
                  <th>Added</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredRestaurants.map(restaurant => (
                  <tr key={restaurant.id}>
                    <td>
                      <div>
                        <div style={{ fontWeight: 600, marginBottom: '4px' }}>{restaurant.name}</div>
                        {restaurant.address && (
                          <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)' }}>
                            {restaurant.address}
                          </div>
                        )}
                      </div>
                    </td>
                    <td>
                      {restaurant.category ? (
                        <span className="admin-badge admin-badge-info">{restaurant.category.name}</span>
                      ) : (
                        <span style={{ color: 'var(--admin-text-muted)', fontSize: '12px' }}>Uncategorized</span>
                      )}
                    </td>
                    <td>
                      {getAverageRating(restaurant) ? (
                        <span className="admin-badge admin-badge-primary">
                          ⭐ {getAverageRating(restaurant)}
                        </span>
                      ) : (
                        <span style={{ color: 'var(--admin-text-muted)', fontSize: '12px' }}>No ratings</span>
                      )}
                    </td>
                    <td style={{ fontSize: '12px', color: 'var(--admin-text-muted)' }}>
                      {new Date(restaurant.created_at).toLocaleDateString()}
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '8px' }}>
                        <a
                          href={`/restaurants/${restaurant.id}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="admin-btn-icon"
                          title="View Restaurant"
                        >
                          <Eye size={16} />
                        </a>
                        <button
                          className="admin-btn-icon admin-btn-danger"
                          onClick={() => handleDelete(restaurant.id, restaurant.name)}
                          title="Delete Restaurant"
                        >
                          <Trash2 size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
