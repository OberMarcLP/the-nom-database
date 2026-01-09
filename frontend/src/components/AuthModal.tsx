import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { X } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import { useEscapeKey } from '../hooks/useEscapeKey';

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  initialMode?: 'login' | 'register';
}

export function AuthModal({ isOpen, onClose, initialMode = 'login' }: AuthModalProps) {
  useEscapeKey(onClose, isOpen);
  const [mode, setMode] = useState<'login' | 'register'>(initialMode);
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [fullName, setFullName] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login, register } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const from = (location.state as any)?.from?.pathname || '/';

  const handleLoginSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await login({ email, password });

      // Check if user must change password
      if (response.user.password_must_change) {
        onClose();
        navigate('/change-password', { state: { required: true } });
      } else {
        onClose();
        navigate(from, { replace: true });
      }
    } catch (err: any) {
      setError(err.message || 'Failed to login');
    } finally {
      setLoading(false);
    }
  };

  const handleRegisterSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    setLoading(true);

    try {
      await register({
        email,
        username,
        password,
        full_name: fullName || undefined,
      });
      onClose();
      navigate('/');
    } catch (err: any) {
      setError(err.message || 'Failed to register');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div style={{ position: 'fixed', inset: 0, zIndex: 50, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '20px', background: 'rgba(0, 0, 0, 0.85)', backdropFilter: 'blur(8px)' }}>
      {/* Backdrop */}
      <div
        style={{ position: 'fixed', inset: 0 }}
        onClick={onClose}
      />

      {/* Modal */}
      <div style={{ position: 'relative', zIndex: 10, width: '100%', maxWidth: '480px', background: 'var(--brutalist-card)', border: '2px solid var(--brutalist-accent)', borderRadius: '8px', padding: '32px', boxShadow: '0 20px 60px rgba(0, 255, 136, 0.3)' }}>
        {/* Close button */}
        <button
          onClick={onClose}
          className="btn-glass"
          style={{ position: 'absolute', top: '16px', right: '16px', padding: '8px', minWidth: 'auto' }}
        >
          <X size={20} />
        </button>

        {/* Header */}
        <div style={{ marginBottom: '32px' }}>
          <h2 style={{ fontSize: '24px', fontWeight: '700', color: 'var(--brutalist-text)', marginBottom: '8px' }}>
            {mode === 'login' ? 'Sign in to your account' : 'Create your account'}
          </h2>
          <p style={{ fontSize: '14px', color: 'var(--brutalist-text-muted)' }}>
            {mode === 'login' ? (
              <>
                Or{' '}
                <button
                  onClick={() => setMode('register')}
                  className="btn-link"
                  style={{ color: 'var(--brutalist-accent)', textDecoration: 'none', background: 'none', border: 'none', padding: 0, cursor: 'pointer', fontWeight: 600 }}
                >
                  create a new account
                </button>
              </>
            ) : (
              <>
                Or{' '}
                <button
                  onClick={() => setMode('login')}
                  className="btn-link"
                  style={{ color: 'var(--brutalist-accent)', textDecoration: 'none', background: 'none', border: 'none', padding: 0, cursor: 'pointer', fontWeight: 600 }}
                >
                  sign in to existing account
                </button>
              </>
            )}
          </p>
        </div>

        {/* Error message */}
        {error && (
          <div style={{ marginBottom: '20px', padding: '16px', border: '2px solid var(--brutalist-danger)', backgroundColor: 'rgba(255, 51, 102, 0.1)', borderRadius: '6px' }}>
            <p style={{ fontSize: '14px', color: 'var(--brutalist-danger)', fontWeight: 600 }}>{error}</p>
          </div>
        )}

        {/* Login Form */}
        {mode === 'login' && (
          <form onSubmit={handleLoginSubmit} style={{ display: 'grid', gap: '20px' }}>
            <div>
              <label htmlFor="email" style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: 'var(--brutalist-text)', marginBottom: '8px' }}>
                Email address
              </label>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                className="input-glass"
                placeholder="Email address"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="password" style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: 'var(--brutalist-text)', marginBottom: '8px' }}>
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                className="input-glass"
                placeholder="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="btn-glass-primary"
              style={{ width: '100%', marginTop: '8px' }}
            >
              {loading ? 'Signing in...' : 'Sign in'}
            </button>
          </form>
        )}

        {/* Register Form */}
        {mode === 'register' && (
          <form onSubmit={handleRegisterSubmit} style={{ display: 'grid', gap: '20px' }}>
            <div>
              <label htmlFor="register-email" style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: 'var(--brutalist-text)', marginBottom: '8px' }}>
                Email address
              </label>
              <input
                id="register-email"
                name="email"
                type="email"
                autoComplete="email"
                required
                className="input-glass"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="register-username" style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: 'var(--brutalist-text)', marginBottom: '8px' }}>
                Username
              </label>
              <input
                id="register-username"
                name="username"
                type="text"
                autoComplete="username"
                required
                className="input-glass"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="register-fullname" style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: 'var(--brutalist-text)', marginBottom: '8px' }}>
                Full Name (optional)
              </label>
              <input
                id="register-fullname"
                name="fullName"
                type="text"
                autoComplete="name"
                className="input-glass"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="register-password" style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: 'var(--brutalist-text)', marginBottom: '8px' }}>
                Password
              </label>
              <input
                id="register-password"
                name="password"
                type="password"
                autoComplete="new-password"
                required
                className="input-glass"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
              <p style={{ marginTop: '4px', fontSize: '12px', color: 'var(--brutalist-text-muted)' }}>Must be at least 8 characters</p>
            </div>
            <div>
              <label htmlFor="register-confirm-password" style={{ display: 'block', fontSize: '14px', fontWeight: 600, color: 'var(--brutalist-text)', marginBottom: '8px' }}>
                Confirm Password
              </label>
              <input
                id="register-confirm-password"
                name="confirmPassword"
                type="password"
                autoComplete="new-password"
                required
                className="input-glass"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="btn-glass-primary"
              style={{ width: '100%', marginTop: '8px' }}
            >
              {loading ? 'Creating account...' : 'Create account'}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
