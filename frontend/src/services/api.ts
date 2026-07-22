export const API_URL = import.meta.env.VITE_API_URL || '';

// Allow only web schemes in user-provided links (blocks javascript:, data: etc.)
export function isSafeHttpUrl(url: string): boolean {
  try {
    const parsed = new URL(url, window.location.origin);
    return parsed.protocol === 'http:' || parsed.protocol === 'https:';
  } catch {
    return false;
  }
}

export interface Category {
  id: number;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface FoodType {
  id: number;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface UserSummary {
  id: number;
  username: string;
  email?: string;
  full_name: string | null;
  avatar_url: string | null;
}

export interface AvgRating {
  food: number;
  service: number;
  ambiance: number;
  overall: number;
  count: number;
}

export interface Restaurant {
  id: number;
  name: string;
  description: string | null;
  address: string | null;
  phone: string | null;
  website: string | null;
  latitude: number | null;
  longitude: number | null;
  google_place_id: string | null;
  category_id: number | null;
  price_range?: number | null; // 1 = $, 2 = $$, 3 = $$$, 4 = $$$$
  category?: Category;
  food_types?: FoodType[];
  avg_rating?: AvgRating;
  distance?: number; // Distance in km from search location
  is_suggestion: boolean; // Indicates if this is from suggestions table
  suggestion_id?: number;
  status?: 'pending' | 'approved' | 'tested' | 'rejected'; // For suggestions
  notes?: string | null; // For suggestions
  user_id?: number; // For suggestions
  user?: UserSummary; // For suggestions
  created_by?: number;
  updated_by?: number;
  created_by_user?: UserSummary;
  updated_by_user?: UserSummary;
  created_at: string;
  updated_at: string;
}

export interface RestaurantFilters {
  q?: string; // search query
  category_id?: number;
  food_type_ids?: number[];
  price_range?: number; // max price range 1-4
  min_rating?: number; // minimum average rating 1-5
  sort?: 'name' | 'rating' | 'date';
  lat?: number;
  lng?: number;
  radius?: number; // in km
  include_suggestions?: boolean;
}

export interface CreateRestaurantData {
  name: string;
  description?: string | null;
  address?: string | null;
  phone?: string | null;
  website?: string | null;
  latitude?: number | null;
  longitude?: number | null;
  google_place_id?: string | null;
  category_id?: number | null;
  price_range?: number; // 1 = $, 2 = $$, 3 = $$$, 4 = $$$$
  food_type_ids?: number[];
}

export interface ReviewPhoto {
  id: number;
  rating_id: number;
  photo_url: string;
  caption?: string;
  display_order: number;
  created_at: string;
}

export interface Rating {
  id: number;
  restaurant_id: number;
  restaurant?: Restaurant;
  food_rating: number;
  service_rating: number;
  ambiance_rating: number;
  comment: string | null;
  user_id?: number;
  user?: UserSummary;
  photos?: ReviewPhoto[];
  helpful_count: number;
  not_helpful_count: number;
  user_vote?: string | null;
  created_at: string;
  updated_at: string;
}

export interface GooglePlaceResult {
  place_id: string;
  name: string;
  address: string;
  phone?: string;
  website?: string;
  latitude: number;
  longitude: number;
}

// The access token lives only in this module-level variable (never in
// localStorage/sessionStorage) so an XSS payload cannot exfiltrate it from
// storage. Sessions survive page reloads exclusively via the httpOnly
// refresh cookie (see tryRefreshTokens/restoreSession).
let accessToken: string | null = null;

export const setAccessToken = (token: string) => {
  accessToken = token;
};

export const clearAccessToken = () => {
  accessToken = null;
};

// One-time migration: older app versions persisted the access token in
// localStorage; purge any leftover copy from previous sessions.
localStorage.removeItem('access_token');

let refreshInFlight: Promise<boolean> | null = null;

// Refresh the session once; concurrent 401s share the same attempt.
// The backend rotates the refresh token on every use, so both tokens
// must be replaced with the response values.
async function tryRefreshTokens(): Promise<boolean> {
  if (!refreshInFlight) {
    refreshInFlight = (async () => {
      try {
        // The refresh token normally arrives via httpOnly cookie; the body
        // fallback only serves clients that still hold a legacy token.
        const legacyToken = localStorage.getItem('refresh_token');
        const res = await fetch(`${API_URL}/api/auth/refresh`, {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(legacyToken ? { refresh_token: legacyToken } : {}),
        });
        if (!res.ok) return false;
        const data = await res.json();
        accessToken = data.access_token;
        localStorage.removeItem('refresh_token');
        return true;
      } catch {
        return false;
      } finally {
        // allow the next expiry to trigger a fresh attempt
        setTimeout(() => {
          refreshInFlight = null;
        }, 0);
      }
    })();
  }
  return refreshInFlight;
}

// Restore a session from the httpOnly refresh cookie (used on app start).
// Resolves true when a fresh in-memory access token was issued.
export const restoreSession = (): Promise<boolean> => tryRefreshTokens();

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const doFetch = () => {
    // Use the in-memory access token (never persisted to storage)
    const token = accessToken;

    const headers: Record<string, string> = {
      ...(options?.headers as Record<string, string> || {}),
    };

    // Only set Content-Type if body is not FormData (browser sets it automatically for FormData)
    if (!(options?.body instanceof FormData)) {
      headers['Content-Type'] = 'application/json';
    }

    // Add Authorization header if token exists
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    return fetch(`${API_URL}/api${endpoint}`, {
      ...options,
      credentials: 'include',
      headers,
    });
  };

  let response = await doFetch();

  // On an expired access token, refresh once and retry the request
  if (response.status === 401 && !endpoint.startsWith('/auth/')) {
    if (await tryRefreshTokens()) {
      response = await doFetch();
    }
  }

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error || 'An error occurred');
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}

// Categories
export const getCategories = () => fetchApi<Category[]>('/categories');
export const createCategory = (name: string) =>
  fetchApi<Category>('/categories', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
export const updateCategory = (id: number, name: string) =>
  fetchApi<Category>(`/categories/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name }),
  });
export const deleteCategory = (id: number) =>
  fetchApi<void>(`/categories/${id}`, { method: 'DELETE' });

// Food Types
export const getFoodTypes = () => fetchApi<FoodType[]>('/food-types');
export const createFoodType = (name: string) =>
  fetchApi<FoodType>('/food-types', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
export const updateFoodType = (id: number, name: string) =>
  fetchApi<FoodType>(`/food-types/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name }),
  });
export const deleteFoodType = (id: number) =>
  fetchApi<void>(`/food-types/${id}`, { method: 'DELETE' });

// Restaurants
export const getRestaurants = (filters?: RestaurantFilters) => {
  const params = new URLSearchParams();
  if (filters?.q) {
    params.set('q', filters.q);
  }
  if (filters?.category_id) {
    params.set('category_id', filters.category_id.toString());
  }
  if (filters?.food_type_ids && filters.food_type_ids.length > 0) {
    params.set('food_type_ids', filters.food_type_ids.join(','));
  }
  if (filters?.price_range) {
    params.set('price_range', filters.price_range.toString());
  }
  if (filters?.min_rating) {
    params.set('min_rating', filters.min_rating.toString());
  }
  if (filters?.sort) {
    params.set('sort', filters.sort);
  }
  if (filters?.lat !== undefined && filters?.lng !== undefined && filters?.radius !== undefined) {
    params.set('lat', filters.lat.toString());
    params.set('lng', filters.lng.toString());
    params.set('radius', filters.radius.toString());
  }
  if (filters?.include_suggestions) {
    params.set('include_suggestions', 'true');
  }
  const queryString = params.toString();
  return fetchApi<Restaurant[]>(`/restaurants${queryString ? `?${queryString}` : ''}`);
};

export const getRestaurant = (id: number) => fetchApi<Restaurant>(`/restaurants/${id}`);
export const createRestaurant = (data: CreateRestaurantData) =>
  fetchApi<Restaurant>('/restaurants', {
    method: 'POST',
    body: JSON.stringify(data),
  });
export const updateRestaurant = (id: number, data: CreateRestaurantData) =>
  fetchApi<Restaurant>(`/restaurants/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
export const deleteRestaurant = (id: number) =>
  fetchApi<void>(`/restaurants/${id}`, { method: 'DELETE' });

// Ratings
export const getRatings = (restaurantId: number) =>
  fetchApi<Rating[]>(`/restaurants/${restaurantId}/ratings`);
export const createRating = (data: {
  restaurant_id: number;
  food_rating: number;
  service_rating: number;
  ambiance_rating: number;
  comment?: string;
}) =>
  fetchApi<Rating>('/ratings', {
    method: 'POST',
    body: JSON.stringify(data),
  });
export const deleteRating = (id: number) =>
  fetchApi<void>(`/ratings/${id}`, { method: 'DELETE' });

export const voteOnReview = (ratingId: number, voteType: 'helpful' | 'not_helpful') =>
  fetchApi<Rating>(`/ratings/${ratingId}/vote`, {
    method: 'POST',
    body: JSON.stringify({ vote_type: voteType }),
  });

export const removeVote = (ratingId: number) =>
  fetchApi<Rating>(`/ratings/${ratingId}/vote`, { method: 'DELETE' });

// Review Photos
export const uploadReviewPhoto = (ratingId: number, photo: File, caption?: string) => {
  const formData = new FormData();
  formData.append('photo', photo);
  if (caption) {
    formData.append('caption', caption);
  }

  return fetchApi<ReviewPhoto>(`/ratings/${ratingId}/photos`, {
    method: 'POST',
    body: formData,
  });
};

export const deleteReviewPhoto = (photoId: number) =>
  fetchApi(`/review-photos/${photoId}`, { method: 'DELETE' });

// Google Maps
export const searchPlaces = (query: string) =>
  fetchApi<GooglePlaceResult[]>(`/places/search?q=${encodeURIComponent(query)}`);
export const getPlaceDetails = (placeId: string) =>
  fetchApi<GooglePlaceResult>(`/places/${placeId}`);
export const geocodeCities = (query: string) =>
  fetchApi<GooglePlaceResult[]>(`/geocode/cities?q=${encodeURIComponent(query)}`);

// Restaurant Suggestions
export interface RestaurantSuggestion {
  id: number;
  name: string;
  address: string | null;
  phone: string | null;
  website: string | null;
  latitude: number | null;
  longitude: number | null;
  google_place_id: string | null;
  suggested_category_id: number | null;
  category?: Category;
  food_types?: FoodType[];
  notes: string | null;
  status: 'pending' | 'approved' | 'tested' | 'rejected';
  user_id?: number;
  user?: UserSummary;
  created_at: string;
  updated_at: string;
}

export interface CreateSuggestionData {
  name: string;
  address?: string | null;
  phone?: string | null;
  website?: string | null;
  latitude?: number | null;
  longitude?: number | null;
  google_place_id?: string | null;
  suggested_category_id?: number | null;
  food_type_ids?: number[];
  notes?: string | null;
}

export const getSuggestions = (status?: string) => {
  const queryString = status ? `?status=${status}` : '';
  return fetchApi<RestaurantSuggestion[]>(`/suggestions${queryString}`);
};
export const getSuggestion = (id: number) =>
  fetchApi<RestaurantSuggestion>(`/suggestions/${id}`);
export const createSuggestion = (data: CreateSuggestionData) =>
  fetchApi<RestaurantSuggestion>('/suggestions', {
    method: 'POST',
    body: JSON.stringify(data),
  });
export const updateSuggestionStatus = (id: number, status: string) =>
  fetchApi<RestaurantSuggestion>(`/suggestions/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  });
export const convertSuggestion = (id: number, data: {
  description?: string;
  category_id?: number;
  food_rating: number;
  service_rating: number;
  ambiance_rating: number;
  comment?: string;
}) =>
  fetchApi<{ restaurant_id: number; message: string }>(`/suggestions/${id}/convert`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
export const deleteSuggestion = (id: number) =>
  fetchApi<void>(`/suggestions/${id}`, { method: 'DELETE' });

// Global Search
export const globalSearch = (query: string, signal?: AbortSignal) =>
  fetchApi<Restaurant[]>(`/search?q=${encodeURIComponent(query)}`, { signal });

// Menu Photos
export interface MenuPhoto {
  id: number;
  restaurant_id: number;
  filename: string;
  original_filename: string | null;
  caption: string;
  file_size: number | null;
  mime_type: string | null;
  url: string;
  created_at: string;
  updated_at: string;
}

export const getMenuPhotos = (restaurantId: number) =>
  fetchApi<MenuPhoto[]>(`/restaurants/${restaurantId}/photos`);

export const uploadMenuPhoto = async (restaurantId: number, photo: File, caption: string): Promise<{ photo: MenuPhoto }> => {
  const formData = new FormData();
  formData.append('photo', photo);
  formData.append('caption', caption);

  const response = await fetch(`${API_URL}/api/restaurants/${restaurantId}/photos`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error || 'Failed to upload photo');
  }

  return response.json();
};

export const updatePhotoCaption = (id: number, caption: string) =>
  fetchApi<MenuPhoto>(`/photos/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ caption }),
  });

export const updateReviewPhotoCaption = (id: number, caption: string) =>
  fetchApi<void>(`/review-photos/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ caption }),
  });

// Runtime configuration served by the backend so published Docker images
// don't need browser keys baked in at build time
export interface RuntimeConfig {
  google_maps_api_key: string;
  google_maps_map_id: string;
}

let runtimeConfigPromise: Promise<RuntimeConfig> | null = null;

export const getRuntimeConfig = (): Promise<RuntimeConfig> => {
  if (!runtimeConfigPromise) {
    runtimeConfigPromise = fetchApi<RuntimeConfig>('/config').catch(() => {
      runtimeConfigPromise = null;
      return { google_maps_api_key: '', google_maps_map_id: '' };
    });
  }
  return runtimeConfigPromise;
};

export const deleteMenuPhoto = (id: number) =>
  fetchApi<void>(`/photos/${id}`, { method: 'DELETE' });

// Authentication
export interface Permission {
  id: number;
  name: string;
  description?: string;
  resource: string;
  action: string;
  created_at: string;
}

export interface Role {
  id: number;
  name: string;
  description?: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
  permissions?: Permission[];
}

export interface User {
  id: number;
  email: string;
  username: string;
  provider: string;
  provider_id?: string;
  full_name?: string;
  avatar_url?: string;
  bio?: string;
  is_active: boolean;
  is_admin: boolean;
  email_verified: boolean;
  password_must_change: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
  roles?: Role[];
  permissions?: string[];
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user: User;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
  full_name?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

// Auth API functions
export const login = (credentials: LoginRequest) =>
  fetchApi<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(credentials),
  });

export const register = (data: RegisterRequest) =>
  fetchApi<LoginResponse>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(data),
  });

export const logout = () =>
  fetchApi<void>('/auth/logout', { method: 'POST', body: JSON.stringify({}) });

export const getCurrentUser = () =>
  fetchApi<User>('/auth/me');

export const changePassword = (data: ChangePasswordRequest) =>
  fetchApi<{ message: string }>('/auth/change-password', {
    method: 'POST',
    body: JSON.stringify(data),
  });

export const refreshToken = (refreshToken: string) =>
  fetchApi<LoginResponse>('/auth/refresh', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

// Restaurant Lists
export interface RestaurantList {
  id: number;
  user_id: number;
  name: string;
  description: string | null;
  is_public: boolean;
  restaurant_count?: number;
  created_at: string;
  updated_at: string;
}

export interface ListRestaurant {
  id: number;
  list_id: number;
  restaurant_id: number;
  restaurant?: Restaurant;
  notes: string | null;
  added_at: string;
}

export interface ListWithRestaurants {
  list: RestaurantList;
  restaurants: ListRestaurant[];
}

export interface ListWithStatus extends RestaurantList {
  contains_restaurant: boolean;
}

export const getUserLists = () =>
  fetchApi<RestaurantList[]>('/lists');

export const getList = (listId: number) =>
  fetchApi<ListWithRestaurants>(`/lists/${listId}`);

export const createList = (data: { name: string; description?: string; is_public: boolean }) =>
  fetchApi<RestaurantList>('/lists', {
    method: 'POST',
    body: JSON.stringify(data),
  });

export const updateList = (listId: number, data: { name?: string; description?: string; is_public?: boolean }) =>
  fetchApi<RestaurantList>(`/lists/${listId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });

export const deleteList = (listId: number) =>
  fetchApi<void>(`/lists/${listId}`, { method: 'DELETE' });

export const addRestaurantToList = (listId: number, restaurantId: number, notes?: string) =>
  fetchApi<ListRestaurant>(`/lists/${listId}/restaurants`, {
    method: 'POST',
    body: JSON.stringify({ restaurant_id: restaurantId, notes }),
  });

export const removeRestaurantFromList = (listId: number, restaurantId: number) =>
  fetchApi<void>(`/lists/${listId}/restaurants/${restaurantId}`, { method: 'DELETE' });

export const getRestaurantLists = (restaurantId: number) =>
  fetchApi<ListWithStatus[]>(`/restaurants/${restaurantId}/lists`);

// User Profile
export interface UserProfile {
  user: User;
  stats: {
    total_reviews: number;
    total_restaurants: number;
    total_lists: number;
    avg_food_rating: number;
    avg_service_rating: number;
    avg_ambiance_rating: number;
  };
}

export const getUserProfile = (userId: number) =>
  fetchApi<UserProfile>(`/users/${userId}`);

export const getUserReviews = (userId: number) =>
  fetchApi<Rating[]>(`/users/${userId}/reviews`);

export const updateUserProfile = (data: { username?: string; full_name?: string; email?: string }) =>
  fetchApi<User>('/user/profile', {
    method: 'PUT',
    body: JSON.stringify(data),
  });

export const uploadAvatar = async (file: File): Promise<{ avatar_url: string; message: string }> => {
  const formData = new FormData();
  formData.append('avatar', file);

  const token = accessToken;
  const response = await fetch(`${API_URL}/api/user/avatar`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    body: formData,
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error || 'Failed to upload avatar');
  }

  return response.json();
};


// Generic API client for admin and other features
export const api = {
  get: <T>(endpoint: string) => fetchApi<T>(endpoint),
  post: <T>(endpoint: string, data?: any) => fetchApi<T>(endpoint, {
    method: 'POST',
    body: data ? JSON.stringify(data) : undefined,
  }),
  put: <T>(endpoint: string, data?: any) => fetchApi<T>(endpoint, {
    method: 'PUT',
    body: data ? JSON.stringify(data) : undefined,
  }),
  delete: <T>(endpoint: string) => fetchApi<T>(endpoint, {
    method: 'DELETE',
  }),
};
