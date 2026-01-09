import { useEffect, useState } from 'react';
import { api } from '../services/api';
import { Image, MessageSquare, Star, Trash2, Eye, Filter } from 'lucide-react';

interface Rating {
  id: number;
  restaurant_id: number;
  restaurant_name: string;
  user_id?: number;
  username?: string;
  food_rating: number;
  service_rating: number;
  ambiance_rating: number;
  comment?: string;
  created_at: string;
}

interface Photo {
  id: number;
  type: 'menu' | 'review';
  rating_id?: number;
  restaurant_id?: number;
  restaurant_name?: string;
  user_id?: number;
  username?: string;
  filename: string;
  caption?: string;
  file_size?: number;
  created_at: string;
}

interface RatingsResponse {
  ratings: Rating[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

interface PhotosResponse {
  photos: Photo[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export function AdminContent() {
  const [activeTab, setActiveTab] = useState<'ratings' | 'photos'>('ratings');

  // Ratings state
  const [ratings, setRatings] = useState<Rating[]>([]);
  const [ratingsPage, setRatingsPage] = useState(1);
  const [ratingsTotalPages, setRatingsTotalPages] = useState(0);
  const [ratingsTotal, setRatingsTotal] = useState(0);
  const [ratingsLoading, setRatingsLoading] = useState(true);

  // Photos state
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [photosPage, setPhotosPage] = useState(1);
  const [photosTotalPages, setPhotosTotalPages] = useState(0);
  const [photosTotal, setPhotosTotal] = useState(0);
  const [photosLoading, setPhotosLoading] = useState(true);
  const [photoTypeFilter, setPhotoTypeFilter] = useState<'all' | 'menu' | 'review'>('all');

  useEffect(() => {
    if (activeTab === 'ratings') {
      loadRatings();
    } else {
      loadPhotos();
    }
  }, [activeTab, ratingsPage, photosPage, photoTypeFilter]);

  const loadRatings = async () => {
    try {
      setRatingsLoading(true);
      const params = new URLSearchParams({
        page: ratingsPage.toString(),
        limit: '20',
      });

      const response = await api.get<RatingsResponse>(`/admin/ratings?${params}`);
      setRatings(response.ratings || []);
      setRatingsTotal(response.pagination.total);
      setRatingsTotalPages(response.pagination.totalPages);
    } catch (error) {
      console.error('Failed to load ratings:', error);
    } finally {
      setRatingsLoading(false);
    }
  };

  const loadPhotos = async () => {
    try {
      setPhotosLoading(true);
      const params = new URLSearchParams({
        page: photosPage.toString(),
        limit: '20',
      });

      if (photoTypeFilter !== 'all') {
        params.append('type', photoTypeFilter);
      }

      const response = await api.get<PhotosResponse>(`/admin/photos?${params}`);
      setPhotos(response.photos || []);
      setPhotosTotal(response.pagination.total);
      setPhotosTotalPages(response.pagination.totalPages);
    } catch (error) {
      console.error('Failed to load photos:', error);
    } finally {
      setPhotosLoading(false);
    }
  };

  const handleDeleteRating = async (ratingId: number, restaurantName: string) => {
    if (!confirm(`Delete rating for "${restaurantName}"? This action cannot be undone.`)) return;

    try {
      await api.delete(`/admin/ratings/${ratingId}`);
      loadRatings();
    } catch (error) {
      console.error('Failed to delete rating:', error);
      alert('Failed to delete rating');
    }
  };

  const handleDeletePhoto = async (photoType: string, photoId: number) => {
    if (!confirm(`Delete this ${photoType} photo? This action cannot be undone.`)) return;

    try {
      await api.delete(`/admin/photos/${photoType}/${photoId}`);
      loadPhotos();
    } catch (error) {
      console.error('Failed to delete photo:', error);
      alert('Failed to delete photo');
    }
  };

  const getAverageRating = (rating: Rating) => {
    return ((rating.food_rating + rating.service_rating + rating.ambiance_rating) / 3).toFixed(1);
  };

  const formatFileSize = (bytes?: number) => {
    if (!bytes) return 'N/A';
    const mb = bytes / (1024 * 1024);
    return `${mb.toFixed(2)} MB`;
  };

  return (
    <div>
      <div className="admin-page-header">
        <h1 className="admin-page-title">Content Moderation</h1>
        <p className="admin-page-description">
          Manage ratings, photos, and reviews
        </p>
      </div>

      {/* Tabs */}
      <div className="admin-card" style={{ marginBottom: '20px' }}>
        <div style={{ display: 'flex', gap: '8px', padding: '16px', borderBottom: '1px solid var(--admin-border)' }}>
          <button
            className={`admin-btn admin-btn-sm ${activeTab === 'ratings' ? '' : 'admin-btn-secondary'}`}
            onClick={() => setActiveTab('ratings')}
          >
            <MessageSquare size={14} />
            Ratings ({ratingsTotal})
          </button>
          <button
            className={`admin-btn admin-btn-sm ${activeTab === 'photos' ? '' : 'admin-btn-secondary'}`}
            onClick={() => setActiveTab('photos')}
          >
            <Image size={14} />
            Photos ({photosTotal})
          </button>
        </div>
      </div>

      {/* Ratings Tab */}
      {activeTab === 'ratings' && (
        <div className="admin-card">
          <div className="admin-card-header">
            <h2 className="admin-card-title">
              <MessageSquare size={20} />
              User Ratings
            </h2>
          </div>

          {ratingsLoading ? (
            <div className="admin-loading">
              <div className="admin-spinner" />
            </div>
          ) : ratings.length === 0 ? (
            <div className="admin-empty">
              <p className="admin-empty-text">No ratings found</p>
            </div>
          ) : (
            <>
              <div className="admin-table-container">
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>Restaurant</th>
                      <th>User</th>
                      <th>Ratings</th>
                      <th>Average</th>
                      <th>Comment</th>
                      <th>Date</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {ratings.map(rating => (
                      <tr key={rating.id}>
                        <td>
                          <div style={{ fontWeight: 600 }}>{rating.restaurant_name}</div>
                        </td>
                        <td>
                          <div>{rating.username || 'Anonymous'}</div>
                        </td>
                        <td>
                          <div style={{ fontSize: '12px', fontFamily: 'IBM Plex Mono, monospace' }}>
                            <div>Food: <Star size={10} style={{ display: 'inline', color: 'var(--admin-warning)' }} /> {rating.food_rating}/5</div>
                            <div>Service: <Star size={10} style={{ display: 'inline', color: 'var(--admin-warning)' }} /> {rating.service_rating}/5</div>
                            <div>Ambiance: <Star size={10} style={{ display: 'inline', color: 'var(--admin-warning)' }} /> {rating.ambiance_rating}/5</div>
                          </div>
                        </td>
                        <td>
                          <span className="admin-badge admin-badge-primary">
                            {getAverageRating(rating)}
                          </span>
                        </td>
                        <td style={{ maxWidth: '300px' }}>
                          {rating.comment ? (
                            <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)' }}>
                              {rating.comment.length > 100
                                ? `${rating.comment.substring(0, 100)}...`
                                : rating.comment}
                            </div>
                          ) : (
                            <span style={{ color: 'var(--admin-text-muted)', fontSize: '12px' }}>No comment</span>
                          )}
                        </td>
                        <td style={{ fontSize: '12px', color: 'var(--admin-text-muted)' }}>
                          {new Date(rating.created_at).toLocaleDateString()}
                        </td>
                        <td>
                          <div style={{ display: 'flex', gap: '8px' }}>
                            <a
                              href={`/restaurants/${rating.restaurant_id}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="admin-btn-icon"
                              title="View Restaurant"
                            >
                              <Eye size={16} />
                            </a>
                            <button
                              className="admin-btn-icon admin-btn-danger"
                              onClick={() => handleDeleteRating(rating.id, rating.restaurant_name)}
                              title="Delete Rating"
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

              {ratingsTotalPages > 1 && (
                <div className="admin-pagination">
                  <button
                    className="admin-btn admin-btn-sm admin-btn-secondary"
                    onClick={() => setRatingsPage(p => Math.max(1, p - 1))}
                    disabled={ratingsPage === 1}
                  >
                    Previous
                  </button>
                  <span className="admin-pagination-info">
                    Page {ratingsPage} of {ratingsTotalPages}
                  </span>
                  <button
                    className="admin-btn admin-btn-sm admin-btn-secondary"
                    onClick={() => setRatingsPage(p => Math.min(ratingsTotalPages, p + 1))}
                    disabled={ratingsPage === ratingsTotalPages}
                  >
                    Next
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      )}

      {/* Photos Tab */}
      {activeTab === 'photos' && (
        <>
          <div className="admin-card" style={{ marginBottom: '20px' }}>
            <div className="admin-card-header">
              <h2 className="admin-card-title">
                <Filter size={20} />
                Filter Photos
              </h2>
            </div>
            <div style={{ display: 'flex', gap: '8px', padding: '20px' }}>
              <button
                className={`admin-btn admin-btn-sm ${photoTypeFilter === 'all' ? '' : 'admin-btn-secondary'}`}
                onClick={() => {
                  setPhotoTypeFilter('all');
                  setPhotosPage(1);
                }}
              >
                All Photos
              </button>
              <button
                className={`admin-btn admin-btn-sm ${photoTypeFilter === 'menu' ? '' : 'admin-btn-secondary'}`}
                onClick={() => {
                  setPhotoTypeFilter('menu');
                  setPhotosPage(1);
                }}
              >
                <Image size={14} />
                Menu Photos
              </button>
              <button
                className={`admin-btn admin-btn-sm ${photoTypeFilter === 'review' ? '' : 'admin-btn-secondary'}`}
                onClick={() => {
                  setPhotoTypeFilter('review');
                  setPhotosPage(1);
                }}
              >
                <MessageSquare size={14} />
                Review Photos
              </button>
            </div>
          </div>

          <div className="admin-card">
            <div className="admin-card-header">
              <h2 className="admin-card-title">
                <Image size={20} />
                User Photos
              </h2>
            </div>

            {photosLoading ? (
              <div className="admin-loading">
                <div className="admin-spinner" />
              </div>
            ) : photos.length === 0 ? (
              <div className="admin-empty">
                <p className="admin-empty-text">No photos found</p>
              </div>
            ) : (
              <>
                <div className="admin-table-container">
                  <table className="admin-table">
                    <thead>
                      <tr>
                        <th>Type</th>
                        <th>Uploaded By</th>
                        <th>Restaurant</th>
                        <th>Filename</th>
                        <th>Caption</th>
                        <th>Size</th>
                        <th>Date</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {photos.map(photo => (
                        <tr key={`${photo.type}-${photo.id}`}>
                          <td>
                            <span className={`admin-badge ${photo.type === 'menu' ? 'admin-badge-info' : 'admin-badge-success'}`}>
                              {photo.type === 'menu' ? 'MENU' : 'REVIEW'}
                            </span>
                          </td>
                          <td>
                            <div>{photo.username || 'Anonymous'}</div>
                          </td>
                          <td>
                            <div>{photo.restaurant_name || '-'}</div>
                          </td>
                          <td>
                            <div style={{ fontSize: '12px', fontFamily: 'IBM Plex Mono, monospace', color: 'var(--admin-text-muted)' }}>
                              {photo.filename}
                            </div>
                          </td>
                          <td style={{ maxWidth: '200px' }}>
                            {photo.caption ? (
                              <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)' }}>
                                {photo.caption.length > 50
                                  ? `${photo.caption.substring(0, 50)}...`
                                  : photo.caption}
                              </div>
                            ) : (
                              <span style={{ color: 'var(--admin-text-muted)', fontSize: '12px' }}>-</span>
                            )}
                          </td>
                          <td style={{ fontSize: '12px', fontFamily: 'IBM Plex Mono, monospace', color: 'var(--admin-text-muted)' }}>
                            {formatFileSize(photo.file_size)}
                          </td>
                          <td style={{ fontSize: '12px', color: 'var(--admin-text-muted)' }}>
                            {new Date(photo.created_at).toLocaleDateString()}
                          </td>
                          <td>
                            <div style={{ display: 'flex', gap: '8px' }}>
                              <a
                                href={`/uploads/${photo.filename}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="admin-btn-icon"
                                title="View Photo"
                              >
                                <Eye size={16} />
                              </a>
                              <button
                                className="admin-btn-icon admin-btn-danger"
                                onClick={() => handleDeletePhoto(photo.type, photo.id)}
                                title="Delete Photo"
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

                {photosTotalPages > 1 && (
                  <div className="admin-pagination">
                    <button
                      className="admin-btn admin-btn-sm admin-btn-secondary"
                      onClick={() => setPhotosPage(p => Math.max(1, p - 1))}
                      disabled={photosPage === 1}
                    >
                      Previous
                    </button>
                    <span className="admin-pagination-info">
                      Page {photosPage} of {photosTotalPages}
                    </span>
                    <button
                      className="admin-btn admin-btn-sm admin-btn-secondary"
                      onClick={() => setPhotosPage(p => Math.min(photosTotalPages, p + 1))}
                      disabled={photosPage === photosTotalPages}
                    >
                      Next
                    </button>
                  </div>
                )}
              </>
            )}
          </div>
        </>
      )}
    </div>
  );
}
