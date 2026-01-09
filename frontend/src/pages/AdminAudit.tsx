import { useEffect, useState } from 'react';
import { api, User } from '../services/api';
import { FileText, Filter, Calendar, User as UserIcon } from 'lucide-react';

interface AuditLog {
  id: number;
  user_id: number | null;
  action: string;
  resource_type: string;
  resource_id: number | null;
  details: any;
  ip_address: string | null;
  user_agent: string | null;
  created_at: string;
  user?: User;
}

interface AuditResponse {
  logs: AuditLog[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export function AdminAudit() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [actionFilter, setActionFilter] = useState('');
  const [resourceFilter, setResourceFilter] = useState('');

  useEffect(() => {
    loadLogs();
  }, [page, actionFilter, resourceFilter]);

  const loadLogs = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams({
        page: page.toString(),
        limit: '50',
      });
      if (actionFilter) params.append('action', actionFilter);
      if (resourceFilter) params.append('resource_type', resourceFilter);

      const response = await api.get<AuditResponse>(`/admin/audit-logs?${params}`);
      setLogs(response.logs);
      setTotal(response.pagination.total);
      setTotalPages(response.pagination.totalPages);
    } catch (error) {
      console.error('Failed to load audit logs:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    }).format(date);
  };

  const getActionColor = (action: string) => {
    if (action.includes('create')) return 'var(--admin-success)';
    if (action.includes('update')) return 'var(--admin-info)';
    if (action.includes('delete')) return 'var(--admin-danger)';
    return 'var(--admin-text-muted)';
  };

  if (loading && logs.length === 0) {
    return (
      <div className="admin-loading">
        <div className="admin-spinner" />
      </div>
    );
  }

  return (
    <div>
      <div className="admin-page-header">
        <h1 className="admin-page-title">Audit Logs</h1>
        <p className="admin-page-description">
          Track all administrative actions and system changes
        </p>
      </div>

      <div className="admin-card" style={{ marginBottom: '20px' }}>
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <Filter size={20} />
            Filters
          </h2>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '16px', padding: '20px' }}>
          <div className="admin-form-group" style={{ marginBottom: 0 }}>
            <label className="admin-label">Action</label>
            <select
              className="admin-input"
              value={actionFilter}
              onChange={(e) => {
                setActionFilter(e.target.value);
                setPage(1);
              }}
            >
              <option value="">All Actions</option>
              <option value="create_user">Create User</option>
              <option value="update_user">Update User</option>
              <option value="delete_user">Delete User</option>
              <option value="assign_role">Assign Role</option>
              <option value="remove_role">Remove Role</option>
              <option value="reset_password">Reset Password</option>
            </select>
          </div>
          <div className="admin-form-group" style={{ marginBottom: 0 }}>
            <label className="admin-label">Resource Type</label>
            <select
              className="admin-input"
              value={resourceFilter}
              onChange={(e) => {
                setResourceFilter(e.target.value);
                setPage(1);
              }}
            >
              <option value="">All Resources</option>
              <option value="user">User</option>
              <option value="role">Role</option>
              <option value="permission">Permission</option>
              <option value="restaurant">Restaurant</option>
              <option value="rating">Rating</option>
            </select>
          </div>
        </div>
      </div>

      <div className="admin-card">
        <div className="admin-card-header">
          <h2 className="admin-card-title">
            <FileText size={20} />
            Activity Log ({total} total entries)
          </h2>
        </div>

        <div className="admin-table-container">
          <table className="admin-table">
            <thead>
              <tr>
                <th>Timestamp</th>
                <th>User</th>
                <th>Action</th>
                <th>Resource</th>
                <th>Details</th>
                <th>IP Address</th>
              </tr>
            </thead>
            <tbody>
              {logs.map(log => (
                <tr key={log.id}>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '12px', fontFamily: 'IBM Plex Mono, monospace' }}>
                      <Calendar size={14} style={{ color: 'var(--admin-text-muted)' }} />
                      {formatDate(log.created_at)}
                    </div>
                  </td>
                  <td>
                    {log.user ? (
                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <UserIcon size={14} style={{ color: 'var(--admin-text-muted)' }} />
                        <span>{log.user.username}</span>
                      </div>
                    ) : (
                      <span style={{ color: 'var(--admin-text-muted)', fontSize: '12px' }}>System</span>
                    )}
                  </td>
                  <td>
                    <span
                      className="admin-badge"
                      style={{
                        background: `${getActionColor(log.action)}20`,
                        color: getActionColor(log.action),
                        borderColor: getActionColor(log.action)
                      }}
                    >
                      {log.action.replace(/_/g, ' ').toUpperCase()}
                    </span>
                  </td>
                  <td>
                    <div style={{ fontFamily: 'IBM Plex Mono, monospace', fontSize: '13px' }}>
                      <div style={{ color: 'var(--admin-text)' }}>{log.resource_type}</div>
                      {log.resource_id && (
                        <div style={{ color: 'var(--admin-text-muted)', fontSize: '11px' }}>
                          ID: {log.resource_id}
                        </div>
                      )}
                    </div>
                  </td>
                  <td>
                    {log.details && (
                      <details style={{ cursor: 'pointer' }}>
                        <summary style={{ fontSize: '12px', color: 'var(--admin-accent)' }}>
                          View Details
                        </summary>
                        <pre style={{
                          marginTop: '8px',
                          padding: '8px',
                          background: 'var(--admin-bg)',
                          border: '1px solid var(--admin-border)',
                          borderRadius: '4px',
                          fontSize: '11px',
                          overflow: 'auto',
                          maxWidth: '300px'
                        }}>
                          {JSON.stringify(log.details, null, 2)}
                        </pre>
                      </details>
                    )}
                  </td>
                  <td style={{ fontFamily: 'IBM Plex Mono, monospace', fontSize: '12px', color: 'var(--admin-text-muted)' }}>
                    {log.ip_address || '-'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {totalPages > 1 && (
          <div className="admin-pagination">
            <button
              className="admin-btn admin-btn-sm admin-btn-secondary"
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              Previous
            </button>
            <span className="admin-pagination-info">
              Page {page} of {totalPages}
            </span>
            <button
              className="admin-btn admin-btn-sm admin-btn-secondary"
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
            >
              Next
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
