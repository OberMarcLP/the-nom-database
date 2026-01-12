import { useState } from 'react';
import { NavLink, Outlet, Navigate } from 'react-router-dom';
import { usePermissions } from '../hooks/usePermissions';
import {
  Users,
  Shield,
  Settings,
  BarChart3,
  AlertTriangle,
  FileText,
  UtensilsCrossed,
  Image,
  ChevronLeft,
  ChevronRight,
  Tag,
} from 'lucide-react';

export function AdminDashboard() {
  const { isAdmin } = usePermissions();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  if (!isAdmin()) {
    return <Navigate to="/" replace />;
  }

  const navItems = [
    { to: '/admin', icon: BarChart3, label: 'Overview', end: true },
    { to: '/admin/users', icon: Users, label: 'Users' },
    { to: '/admin/roles', icon: Shield, label: 'Roles' },
    { to: '/admin/restaurants', icon: UtensilsCrossed, label: 'Restaurants' },
    { to: '/admin/categories', icon: Tag, label: 'Categories & Types' },
    { to: '/admin/content', icon: Image, label: 'Content' },
    { to: '/admin/analytics', icon: BarChart3, label: 'Analytics' },
    { to: '/admin/audit', icon: FileText, label: 'Audit Logs' },
    { to: '/admin/settings', icon: Settings, label: 'Settings' },
  ];

  return (
    <div className="admin-layout">
      <aside className={`admin-sidebar ${sidebarCollapsed ? 'collapsed' : ''}`}>
        <div className="admin-sidebar-header">
          {!sidebarCollapsed && (
            <div className="admin-logo">
              <AlertTriangle size={28} strokeWidth={2.5} />
              <span>ADMIN</span>
            </div>
          )}
          <button
            className="sidebar-toggle"
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            aria-label={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {sidebarCollapsed ? <ChevronRight size={20} /> : <ChevronLeft size={20} />}
          </button>
        </div>

        <nav className="admin-nav">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `admin-nav-item ${isActive ? 'active' : ''}`
              }
              title={sidebarCollapsed ? item.label : undefined}
            >
              <item.icon size={20} strokeWidth={2} />
              {!sidebarCollapsed && <span>{item.label}</span>}
            </NavLink>
          ))}
        </nav>

        <div className="admin-sidebar-footer">
          {!sidebarCollapsed && (
            <div className="admin-build-info">
              <div className="build-version">v1.0.0</div>
              <div className="build-status">Production</div>
            </div>
          )}
        </div>
      </aside>

      <main className="admin-content">
        <div className="admin-content-inner">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
