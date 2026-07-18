import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import { API_URL } from '../services/api';

export function AuthCallbackPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { updateUser } = useAuth();

  useEffect(() => {
    const accessToken = searchParams.get('access_token');
    const refreshToken = searchParams.get('refresh_token');

    if (accessToken && refreshToken) {
      // Store tokens
      localStorage.setItem('access_token', accessToken);
      localStorage.setItem('refresh_token', refreshToken);

      // Fetch full user details
      fetch(`${API_URL}/api/auth/me`, {
        headers: {
          'Authorization': `Bearer ${accessToken}`
        }
      })
        .then(res => res.json())
        .then(data => {
          updateUser(data);
          // Redirect to home
          window.location.href = '/';
        })
        .catch(err => {
          console.error('Failed to fetch user:', err);
          navigate('/');
        });
    } else {
      // No tokens, redirect to home
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
