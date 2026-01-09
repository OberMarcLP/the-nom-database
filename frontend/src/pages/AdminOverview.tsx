import { useEffect, useState } from 'react';
import { api } from '../services/api';
import { Users, UtensilsCrossed, Star, TrendingUp, Activity } from 'lucide-react';

interface SystemStats {
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

export function AdminOverview() {
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      const response = await api.get<SystemStats>('/admin/stats');
      setStats(response);
    } catch (error) {
      console.error('Failed to load stats:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="admin-loading">
        <div className="admin-spinner" />
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="admin-empty">
        <p className="admin-empty-text">Failed to load statistics</p>
      </div>
    );
  }

  const calculateChange = (current: number, total: number) => {
    if (total === 0) return 0;
    return ((current / total) * 100).toFixed(1);
  };

  return (
    <div>
      <div className="admin-page-header">
        <h1 className="admin-page-title">System Overview</h1>
        <p className="admin-page-description">
          Real-time metrics and system health monitoring
        </p>
      </div>

      <div className="admin-stats-grid">
        <div className="admin-stat-card">
          <div className="admin-stat-label">
            <Users size={14} style={{ display: 'inline', marginRight: '6px' }} />
            Total Users
          </div>
          <div className="admin-stat-value">{stats.users.total.toLocaleString()}</div>
          <div className="admin-stat-change positive">
            <TrendingUp size={14} />
            {stats.users.new_last_30_days} new (30d)
          </div>
        </div>

        <div className="admin-stat-card">
          <div className="admin-stat-label">
            <UtensilsCrossed size={14} style={{ display: 'inline', marginRight: '6px' }} />
            Restaurants
          </div>
          <div className="admin-stat-value">{stats.restaurants.total.toLocaleString()}</div>
          <div className="admin-stat-change">
            {stats.restaurants.pending_suggestions} pending
          </div>
        </div>

        <div className="admin-stat-card">
          <div className="admin-stat-label">
            <Star size={14} style={{ display: 'inline', marginRight: '6px' }} />
            Total Ratings
          </div>
          <div className="admin-stat-value">{stats.ratings.total.toLocaleString()}</div>
          <div className="admin-stat-change">
            Avg: {stats.ratings.avg_food.toFixed(1)} / 5.0
          </div>
        </div>

        <div className="admin-stat-card">
          <div className="admin-stat-label">
            <Activity size={14} style={{ display: 'inline', marginRight: '6px' }} />
            Activity (7d)
          </div>
          <div className="admin-stat-value">{stats.activity.last_7_days.toLocaleString()}</div>
          <div className="admin-stat-change positive">
            <TrendingUp size={14} />
            Active
          </div>
        </div>
      </div>

      <div className="admin-card">
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <Users size={20} />
            User Statistics
          </h2>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '20px' }}>
          <div>
            <div className="admin-label">Active Users</div>
            <div style={{ fontSize: '24px', fontWeight: 700, color: 'var(--admin-accent)', fontFamily: 'IBM Plex Mono, monospace' }}>
              {stats.users.active.toLocaleString()}
            </div>
            <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
              {calculateChange(stats.users.active, stats.users.total)}% of total
            </div>
          </div>
          <div>
            <div className="admin-label">Verified Emails</div>
            <div style={{ fontSize: '24px', fontWeight: 700, color: 'var(--admin-info)', fontFamily: 'IBM Plex Mono, monospace' }}>
              {stats.users.verified.toLocaleString()}
            </div>
            <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
              {calculateChange(stats.users.verified, stats.users.total)}% verified
            </div>
          </div>
          <div>
            <div className="admin-label">New (30 days)</div>
            <div style={{ fontSize: '24px', fontWeight: 700, color: 'var(--admin-warning)', fontFamily: 'IBM Plex Mono, monospace' }}>
              {stats.users.new_last_30_days.toLocaleString()}
            </div>
            <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
              Growth rate
            </div>
          </div>
        </div>
      </div>

      <div className="admin-card">
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <UtensilsCrossed size={20} />
            Restaurant & Content Metrics
          </h2>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '20px' }}>
          <div>
            <div className="admin-label">Menu Photos</div>
            <div style={{ fontSize: '28px', fontWeight: 700, color: 'var(--admin-text)', fontFamily: 'IBM Plex Mono, monospace' }}>
              {stats.content.menu_photos.toLocaleString()}
            </div>
          </div>
          <div>
            <div className="admin-label">Review Photos</div>
            <div style={{ fontSize: '28px', fontWeight: 700, color: 'var(--admin-text)', fontFamily: 'IBM Plex Mono, monospace' }}>
              {stats.content.review_photos.toLocaleString()}
            </div>
          </div>
          <div>
            <div className="admin-label">User Lists</div>
            <div style={{ fontSize: '28px', fontWeight: 700, color: 'var(--admin-text)', fontFamily: 'IBM Plex Mono, monospace' }}>
              {stats.content.lists.toLocaleString()}
            </div>
          </div>
        </div>
      </div>

      <div className="admin-card">
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <Star size={20} />
            Rating Breakdown
          </h2>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '20px' }}>
          <div>
            <div className="admin-label">Food Quality</div>
            <div style={{ fontSize: '32px', fontWeight: 700, color: 'var(--admin-accent)', fontFamily: 'IBM Plex Mono, monospace' }}>
              {stats.ratings.avg_food.toFixed(2)}
            </div>
            <div style={{ width: '100%', height: '6px', background: 'var(--admin-border)', borderRadius: '3px', marginTop: '8px', overflow: 'hidden' }}>
              <div style={{ width: `${(stats.ratings.avg_food / 5) * 100}%`, height: '100%', background: 'var(--admin-accent)', transition: 'width 1s ease-out' }} />
            </div>
          </div>
          <div>
            <div className="admin-label">Service</div>
            <div style={{ fontSize: '32px', fontWeight: 700, color: 'var(--admin-info)', fontFamily: 'IBM Plex Mono, monospace' }}>
              {stats.ratings.avg_service.toFixed(2)}
            </div>
            <div style={{ width: '100%', height: '6px', background: 'var(--admin-border)', borderRadius: '3px', marginTop: '8px', overflow: 'hidden' }}>
              <div style={{ width: `${(stats.ratings.avg_service / 5) * 100}%`, height: '100%', background: 'var(--admin-info)', transition: 'width 1s ease-out' }} />
            </div>
          </div>
          <div>
            <div className="admin-label">Ambiance</div>
            <div style={{ fontSize: '32px', fontWeight: 700, color: 'var(--admin-warning)', fontFamily: 'IBM Plex Mono, monospace' }}>
              {stats.ratings.avg_ambiance.toFixed(2)}
            </div>
            <div style={{ width: '100%', height: '6px', background: 'var(--admin-border)', borderRadius: '3px', marginTop: '8px', overflow: 'hidden' }}>
              <div style={{ width: `${(stats.ratings.avg_ambiance / 5) * 100}%`, height: '100%', background: 'var(--admin-warning)', transition: 'width 1s ease-out' }} />
            </div>
          </div>
        </div>
      </div>

      {stats.restaurants.pending_suggestions > 0 && (
        <div className="admin-card" style={{ borderColor: 'var(--admin-warning)' }}>
          <div className="admin-card-header">
            <h2 className="admin-card-title" style={{ color: 'var(--admin-warning)' }}>
              <TrendingUp size={20} />
              Pending Actions Required
            </h2>
          </div>
          <div style={{ fontSize: '14px', color: 'var(--admin-text-muted)' }}>
            <p style={{ margin: '0 0 12px 0' }}>
              You have <strong style={{ color: 'var(--admin-warning)' }}>{stats.restaurants.pending_suggestions} pending restaurant suggestions</strong> awaiting review.
            </p>
            <a href="/admin/restaurants" className="admin-btn admin-btn-sm" style={{ textDecoration: 'none' }}>
              Review Suggestions
            </a>
          </div>
        </div>
      )}
    </div>
  );
}
