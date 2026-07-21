import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import { API_URL, setAccessToken } from '../services/api';

export function AuthCallbackPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { updateUser } = useAuth();

  useEffect(() => {
    // The backend redirects here with a one-time code; the tokens
    // themselves are fetched via POST so they never appear in a URL.
    const code = searchParams.get('code');

    if (code) {
      fetch(`${API_URL}/api/auth/oidc/exchange`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code }),
      })
        .then(res => {
          if (!res.ok) throw new Error('Code exchange failed');
          return res.json();
        })
        .then(data => {
          // Refresh token arrives as httpOnly cookie from the exchange;
          // the access token is held in memory only, so navigate within
          // the SPA (a hard reload would discard it and force an extra
          // silent-refresh roundtrip).
          setAccessToken(data.access_token);
          localStorage.removeItem('refresh_token');
          updateUser(data.user);
          navigate('/', { replace: true });
        })
        .catch(err => {
          console.error('Failed to complete OIDC login:', err);
          navigate('/');
        });
    } else {
      // No code, redirect to home
      navigate('/');
    }
  }, [searchParams, navigate, updateUser]);

  return (
    <div className="flex items-center justify-center h-screen">
      <div className="text-center">
        <Loader2 className="w-12 h-12 animate-spin text-(--accent) mx-auto mb-4" />
        <p className="text-(--text-muted)">Completing authentication...</p>
      </div>
    </div>
  );
}

export default AuthCallbackPage;
