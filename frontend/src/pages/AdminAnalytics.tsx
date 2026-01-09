import { useEffect, useState } from 'react';
import { api } from '../services/api';
import { BarChart3, Users, TrendingUp, Star, Image as ImageIcon, Activity } from 'lucide-react';

interface Statistics {
  users: {
    total: number;
    active: number;
    verified: number;
    new_last_30_days: number;
  };
  restaurants: {
    total: number;
    pending_suggestions: number;
    approved_suggestions: number;
    rejected_suggestions: number;
  };
  ratings: {
    total: number;
    avg_food: number;
    avg_service: number;
    avg_ambiance: number;
  };
  content: {
    menu_photos: number;
    review_photos: number;
    lists: number;
  };
  activity: {
    last_7_days: number;
  };
}

interface UserGrowth {
  date: string;
  count: number;
}

interface ActiveUser {
  id: number;
  username: string;
  email: string;
  avatar_url?: string;
  rating_count: number;
  restaurant_count: number;
  list_count: number;
  total_activity: number;
}

interface PopularRestaurant {
  id: number;
  name: string;
  address?: string;
  category?: string;
  rating_count: number;
  avg_rating: number;
}

export function AdminAnalytics() {
  const [stats, setStats] = useState<Statistics | null>(null);
  const [userGrowth, setUserGrowth] = useState<UserGrowth[]>([]);
  const [activeUsers, setActiveUsers] = useState<ActiveUser[]>([]);
  const [popularRestaurants, setPopularRestaurants] = useState<PopularRestaurant[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadAnalytics();
  }, []);

  const loadAnalytics = async () => {
    try {
      setLoading(true);
      const [statsData, growthData, usersData, restaurantsData] = await Promise.all([
        api.get<Statistics>('/admin/stats'),
        api.get<UserGrowth[]>('/admin/analytics/user-growth?days=30'),
        api.get<ActiveUser[]>('/admin/analytics/active-users?limit=10'),
        api.get<PopularRestaurant[]>('/admin/analytics/popular-restaurants?limit=10'),
      ]);

      setStats(statsData);
      setUserGrowth(Array.isArray(growthData) ? growthData : []);
      setActiveUsers(Array.isArray(usersData) ? usersData : []);
      setPopularRestaurants(Array.isArray(restaurantsData) ? restaurantsData : []);
    } catch (error) {
      console.error('Failed to load analytics:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading || !stats) {
    return (
      <div className="admin-loading">
        <div className="admin-spinner" />
      </div>
    );
  }

  const maxGrowthCount = Math.max(...userGrowth.map(d => d.count), 1);
  const maxActivity = Math.max(...activeUsers.map(u => u.total_activity), 1);

  return (
    <div>
      <div className="admin-page-header">
        <h1 className="admin-page-title">Analytics Dashboard</h1>
        <p className="admin-page-description">
          View detailed analytics and performance metrics
        </p>
      </div>

      {/* Key Metrics */}
      <div className="admin-stats-grid">
        <div className="admin-stat-card">
          <div className="admin-stat-label">
            <Users size={14} style={{ display: 'inline', marginRight: '6px' }} />
            Total Users
          </div>
          <div className="admin-stat-value">{stats.users.total}</div>
          <div className="admin-stat-change positive">
            +{stats.users.new_last_30_days} last 30 days
          </div>
        </div>

        <div className="admin-stat-card">
          <div className="admin-stat-label">
            <TrendingUp size={14} style={{ display: 'inline', marginRight: '6px' }} />
            Total Restaurants
          </div>
          <div className="admin-stat-value">{stats.restaurants.total}</div>
          <div className="admin-stat-change">
            {stats.restaurants.pending_suggestions} pending suggestions
          </div>
        </div>

        <div className="admin-stat-card">
          <div className="admin-stat-label">
            <Star size={14} style={{ display: 'inline', marginRight: '6px' }} />
            Total Ratings
          </div>
          <div className="admin-stat-value">{stats.ratings.total}</div>
          <div className="admin-stat-change">
            Avg: {((stats.ratings.avg_food + stats.ratings.avg_service + stats.ratings.avg_ambiance) / 3).toFixed(1)}/5
          </div>
        </div>

        <div className="admin-stat-card">
          <div className="admin-stat-label">
            <Activity size={14} style={{ display: 'inline', marginRight: '6px' }} />
            Recent Activity
          </div>
          <div className="admin-stat-value">{stats.activity.last_7_days}</div>
          <div className="admin-stat-change">
            Last 7 days
          </div>
        </div>
      </div>

      {/* User Growth Chart */}
      <div className="admin-card" style={{ marginBottom: '20px' }}>
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <BarChart3 size={20} />
            User Growth (Last 30 Days)
          </h2>
        </div>
        <div style={{ padding: '20px' }}>
          {userGrowth.length === 0 ? (
            <p style={{ color: 'var(--admin-text-muted)', textAlign: 'center', padding: '40px 0' }}>
              No user growth data available
            </p>
          ) : (
            <div style={{ display: 'flex', alignItems: 'flex-end', gap: '4px', height: '200px' }}>
              {userGrowth.map((item, index) => (
                <div
                  key={index}
                  style={{
                    flex: 1,
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'flex-end',
                    gap: '8px',
                  }}
                >
                  <div
                    style={{
                      width: '100%',
                      height: `${(item.count / maxGrowthCount) * 160}px`,
                      minHeight: item.count > 0 ? '4px' : '0',
                      background: 'var(--admin-accent)',
                      borderRadius: '4px 4px 0 0',
                      transition: 'all 0.3s ease',
                      position: 'relative',
                    }}
                    title={`${item.date}: ${item.count} users`}
                  >
                    {item.count > 0 && (
                      <div
                        style={{
                          position: 'absolute',
                          top: '-20px',
                          left: '50%',
                          transform: 'translateX(-50%)',
                          fontSize: '10px',
                          fontFamily: 'IBM Plex Mono, monospace',
                          color: 'var(--admin-text-muted)',
                        }}
                      >
                        {item.count}
                      </div>
                    )}
                  </div>
                  <div
                    style={{
                      fontSize: '9px',
                      fontFamily: 'IBM Plex Mono, monospace',
                      color: 'var(--admin-text-muted)',
                      transform: 'rotate(-45deg)',
                      whiteSpace: 'nowrap',
                      marginTop: '12px',
                    }}
                  >
                    {new Date(item.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Rating Breakdown */}
      <div className="admin-card" style={{ marginBottom: '20px' }}>
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <Star size={20} />
            Average Rating Breakdown
          </h2>
        </div>
        <div style={{ padding: '20px' }}>
          <div style={{ display: 'grid', gap: '16px' }}>
            {[
              { label: 'Food Quality', value: stats.ratings.avg_food, color: 'var(--admin-success)' },
              { label: 'Service', value: stats.ratings.avg_service, color: 'var(--admin-info)' },
              { label: 'Ambiance', value: stats.ratings.avg_ambiance, color: 'var(--admin-warning)' },
            ].map(rating => (
              <div key={rating.label}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                  <span style={{ fontSize: '13px', fontWeight: 600 }}>{rating.label}</span>
                  <span style={{ fontSize: '13px', fontFamily: 'IBM Plex Mono, monospace', color: 'var(--admin-text-muted)' }}>
                    {rating.value.toFixed(2)} / 5.00
                  </span>
                </div>
                <div
                  style={{
                    width: '100%',
                    height: '8px',
                    background: 'var(--admin-bg-secondary)',
                    borderRadius: '4px',
                    overflow: 'hidden',
                  }}
                >
                  <div
                    style={{
                      width: `${(rating.value / 5) * 100}%`,
                      height: '100%',
                      background: rating.color,
                      transition: 'width 0.3s ease',
                    }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Content Statistics */}
      <div className="admin-card" style={{ marginBottom: '20px' }}>
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <ImageIcon size={20} />
            Content Statistics
          </h2>
        </div>
        <div className="admin-stats-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr)', padding: '20px' }}>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '32px', fontWeight: 700, color: 'var(--admin-accent)' }}>
              {stats.content.menu_photos}
            </div>
            <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
              Menu Photos
            </div>
          </div>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '32px', fontWeight: 700, color: 'var(--admin-info)' }}>
              {stats.content.review_photos}
            </div>
            <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
              Review Photos
            </div>
          </div>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '32px', fontWeight: 700, color: 'var(--admin-success)' }}>
              {stats.content.lists}
            </div>
            <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
              User Lists
            </div>
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', gap: '20px' }}>
        {/* Most Active Users */}
        <div className="admin-card">
          <div className="admin-card-header">
            <h2 className="admin-card-title">
              <Users size={20} />
              Most Active Users
            </h2>
          </div>
          <div style={{ padding: '20px' }}>
            {activeUsers.length === 0 ? (
              <p style={{ color: 'var(--admin-text-muted)', textAlign: 'center' }}>No active users</p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {activeUsers.map((user, index) => (
                  <div
                    key={user.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '12px',
                      padding: '12px',
                      background: 'var(--admin-bg-secondary)',
                      border: '1px solid var(--admin-border)',
                      borderRadius: '4px',
                    }}
                  >
                    <div
                      style={{
                        width: '24px',
                        height: '24px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        background: index < 3 ? 'var(--admin-accent)' : 'var(--admin-border)',
                        color: index < 3 ? '#000' : 'var(--admin-text-muted)',
                        borderRadius: '50%',
                        fontSize: '12px',
                        fontWeight: 700,
                        fontFamily: 'IBM Plex Mono, monospace',
                      }}
                    >
                      {index + 1}
                    </div>
                    <div style={{ flex: 1 }}>
                      <div style={{ fontWeight: 600, fontSize: '13px' }}>{user.username}</div>
                      <div style={{ fontSize: '11px', color: 'var(--admin-text-muted)', fontFamily: 'IBM Plex Mono, monospace' }}>
                        {user.rating_count} ratings • {user.restaurant_count} restaurants • {user.list_count} lists
                      </div>
                    </div>
                    <div
                      style={{
                        width: '60px',
                        height: '6px',
                        background: 'var(--admin-bg)',
                        borderRadius: '3px',
                        overflow: 'hidden',
                      }}
                    >
                      <div
                        style={{
                          width: `${(user.total_activity / maxActivity) * 100}%`,
                          height: '100%',
                          background: 'var(--admin-accent)',
                        }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Popular Restaurants */}
        <div className="admin-card">
          <div className="admin-card-header">
            <h2 className="admin-card-title">
              <TrendingUp size={20} />
              Popular Restaurants
            </h2>
          </div>
          <div style={{ padding: '20px' }}>
            {popularRestaurants.length === 0 ? (
              <p style={{ color: 'var(--admin-text-muted)', textAlign: 'center' }}>No restaurants yet</p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {popularRestaurants.map((restaurant, index) => (
                  <div
                    key={restaurant.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '12px',
                      padding: '12px',
                      background: 'var(--admin-bg-secondary)',
                      border: '1px solid var(--admin-border)',
                      borderRadius: '4px',
                    }}
                  >
                    <div
                      style={{
                        width: '24px',
                        height: '24px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        background: index < 3 ? 'var(--admin-warning)' : 'var(--admin-border)',
                        color: index < 3 ? '#000' : 'var(--admin-text-muted)',
                        borderRadius: '50%',
                        fontSize: '12px',
                        fontWeight: 700,
                        fontFamily: 'IBM Plex Mono, monospace',
                      }}
                    >
                      {index + 1}
                    </div>
                    <div style={{ flex: 1 }}>
                      <div style={{ fontWeight: 600, fontSize: '13px' }}>{restaurant.name}</div>
                      <div style={{ fontSize: '11px', color: 'var(--admin-text-muted)' }}>
                        {restaurant.category || 'Uncategorized'}
                      </div>
                    </div>
                    <div style={{ textAlign: 'right' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '4px', justifyContent: 'flex-end' }}>
                        <Star size={12} style={{ color: 'var(--admin-warning)' }} />
                        <span style={{ fontSize: '13px', fontWeight: 600 }}>
                          {restaurant.avg_rating.toFixed(1)}
                        </span>
                      </div>
                      <div style={{ fontSize: '11px', color: 'var(--admin-text-muted)', fontFamily: 'IBM Plex Mono, monospace' }}>
                        {restaurant.rating_count} ratings
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
