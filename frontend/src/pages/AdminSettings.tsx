import { useEffect, useState } from 'react';
import { api } from '../services/api';
import { Settings, Database, Key, Globe, Mail, Save } from 'lucide-react';

interface SystemSettings {
  app_name: string;
  auth_mode: string;
  jwt_expiration: string;
  jwt_refresh_expiration: string;
  allowed_origins: string;
  google_maps_enabled: boolean;
  s3_enabled: boolean;
  oidc_enabled: boolean;
  oidc_issuer: string;
}

export function AdminSettings() {
  const [settings, setSettings] = useState<SystemSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setLoading(true);
      const response = await api.get<SystemSettings>('/admin/settings');
      setSettings(response);
    } catch (error) {
      console.error('Failed to load settings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!settings) return;

    try {
      setSaving(true);
      await api.put('/admin/settings', settings);
      alert('Settings saved successfully! Note: Some changes may require server restart.');
    } catch (error) {
      console.error('Failed to save settings:', error);
      alert('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="admin-loading">
        <div className="admin-spinner" />
      </div>
    );
  }

  if (!settings) {
    return (
      <div className="admin-empty">
        <p className="admin-empty-text">Failed to load settings</p>
      </div>
    );
  }

  return (
    <div>
      <div className="admin-page-header">
        <h1 className="admin-page-title">System Settings</h1>
        <p className="admin-page-description">
          Configure system-wide application settings
        </p>
      </div>

      <form onSubmit={handleSave}>
        <div className="admin-card" style={{ marginBottom: '20px' }}>
          <div className="admin-card-header">
            <h2 className="admin-card-title">
              <Settings size={20} />
              General Settings
            </h2>
          </div>
          <div style={{ padding: '20px', display: 'grid', gap: '20px' }}>
            <div className="admin-form-group">
              <label className="admin-label">Application Name</label>
              <input
                type="text"
                className="admin-input"
                value={settings.app_name}
                onChange={(e) => setSettings({ ...settings, app_name: e.target.value })}
                placeholder="The Nom Database"
                readOnly
              />
              <p style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
                Display name for the application (read-only)
              </p>
            </div>
          </div>
        </div>

        <div className="admin-card" style={{ marginBottom: '20px' }}>
          <div className="admin-card-header">
            <h2 className="admin-card-title">
              <Key size={20} />
              Authentication Settings
            </h2>
          </div>
          <div style={{ padding: '20px', display: 'grid', gap: '20px' }}>
            <div className="admin-form-group">
              <label className="admin-label">Authentication Mode</label>
              <select
                className="admin-input"
                value={settings.auth_mode}
                onChange={(e) => setSettings({ ...settings, auth_mode: e.target.value })}
                disabled
              >
                <option value="none">None (No Authentication)</option>
                <option value="local">Local (JWT-based)</option>
                <option value="oauth">OAuth/OIDC Only</option>
                <option value="both">Both Local + OAuth/OIDC</option>
              </select>
              <p style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
                Current mode: <strong>{settings.auth_mode}</strong> (requires env variable change + restart)
              </p>
            </div>

            <div className="admin-form-group">
              <label className="admin-label">JWT Access Token Expiration</label>
              <input
                type="text"
                className="admin-input"
                value={settings.jwt_expiration}
                readOnly
                placeholder="e.g., 15m"
              />
              <p style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
                How long access tokens remain valid (read-only, set via JWT_EXPIRATION env)
              </p>
            </div>

            <div className="admin-form-group">
              <label className="admin-label">JWT Refresh Token Expiration</label>
              <input
                type="text"
                className="admin-input"
                value={settings.jwt_refresh_expiration}
                readOnly
                placeholder="e.g., 7d"
              />
              <p style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
                How long refresh tokens remain valid (read-only, set via JWT_REFRESH_EXPIRATION env)
              </p>
            </div>
          </div>
        </div>

        <div className="admin-card" style={{ marginBottom: '20px' }}>
          <div className="admin-card-header">
            <h2 className="admin-card-title">
              <Globe size={20} />
              CORS & Network Settings
            </h2>
          </div>
          <div style={{ padding: '20px', display: 'grid', gap: '20px' }}>
            <div className="admin-form-group">
              <label className="admin-label">Allowed Origins</label>
              <input
                type="text"
                className="admin-input"
                value={settings.allowed_origins}
                readOnly
                placeholder="http://localhost:3000"
              />
              <p style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
                Comma-separated list of allowed CORS origins (read-only, set via ALLOWED_ORIGINS env)
              </p>
            </div>
          </div>
        </div>

        <div className="admin-card" style={{ marginBottom: '20px' }}>
          <div className="admin-card-header">
            <h2 className="admin-card-title">
              <Database size={20} />
              External Services
            </h2>
          </div>
          <div style={{ padding: '20px', display: 'grid', gap: '20px' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px', background: 'var(--admin-bg-secondary)', borderRadius: '4px', border: '1px solid var(--admin-border)' }}>
              <div>
                <div style={{ fontWeight: 600, marginBottom: '4px' }}>Google Maps API</div>
                <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)' }}>
                  Used for restaurant location and mapping features
                </div>
              </div>
              <span className={`admin-badge ${settings.google_maps_enabled ? 'admin-badge-success' : 'admin-badge-danger'}`}>
                {settings.google_maps_enabled ? 'ENABLED' : 'DISABLED'}
              </span>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px', background: 'var(--admin-bg-secondary)', borderRadius: '4px', border: '1px solid var(--admin-border)' }}>
              <div>
                <div style={{ fontWeight: 600, marginBottom: '4px' }}>AWS S3 Storage</div>
                <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)' }}>
                  Cloud storage for photos and media uploads
                </div>
              </div>
              <span className={`admin-badge ${settings.s3_enabled ? 'admin-badge-success' : 'admin-badge-warning'}`}>
                {settings.s3_enabled ? 'ENABLED' : 'LOCAL STORAGE'}
              </span>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px', background: 'var(--admin-bg-secondary)', borderRadius: '4px', border: '1px solid var(--admin-border)' }}>
              <div>
                <div style={{ fontWeight: 600, marginBottom: '4px' }}>OIDC Authentication</div>
                <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)' }}>
                  {settings.oidc_enabled ? `Provider: ${settings.oidc_issuer}` : 'Single Sign-On via OpenID Connect'}
                </div>
              </div>
              <span className={`admin-badge ${settings.oidc_enabled ? 'admin-badge-success' : 'admin-badge-danger'}`}>
                {settings.oidc_enabled ? 'ENABLED' : 'DISABLED'}
              </span>
            </div>
          </div>
        </div>

        <div className="admin-card">
          <div style={{ padding: '20px', background: 'var(--admin-bg-secondary)', borderTop: '1px solid var(--admin-border)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '12px' }}>
              <Mail size={16} style={{ color: 'var(--admin-info)' }} />
              <p style={{ fontSize: '14px', color: 'var(--admin-text)' }}>
                Most settings are configured via environment variables and require a server restart to take effect.
              </p>
            </div>
            <p style={{ fontSize: '12px', color: 'var(--admin-text-muted)' }}>
              To modify these settings, update your <code style={{ background: 'var(--admin-bg)', padding: '2px 6px', borderRadius: '2px', fontFamily: 'IBM Plex Mono, monospace' }}>.env</code> file and restart the application.
            </p>
          </div>
        </div>

        {/* Save button (disabled since most settings are read-only) */}
        <div style={{ marginTop: '20px', display: 'flex', justifyContent: 'flex-end' }}>
          <button
            type="submit"
            className="admin-btn"
            disabled={saving}
            style={{ opacity: 0.5, cursor: 'not-allowed' }}
            title="Settings are configured via environment variables"
          >
            {saving ? (
              <>
                <div className="admin-spinner" style={{ width: '16px', height: '16px', marginRight: '8px' }} />
                Saving...
              </>
            ) : (
              <>
                <Save size={16} />
                Save Settings (Read-Only)
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
}
