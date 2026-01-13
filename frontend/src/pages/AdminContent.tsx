import { useEffect, useState } from 'react';
import { api } from '../services/api';
import { MessageSquare, Star, Trash2, Eye, Camera } from 'lucide-react';
import { ConfirmDialog } from '../components/ConfirmDialog';

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
  photos?: Photo[];
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
  const [activeTab, setActiveTab] = useState<'content' | 'photos'>('content');

  // Content state (ratings with their photos)
  const [ratings, setRatings] = useState<Rating[]>([]);
  const [ratingsPage, setRatingsPage] = useState(1);
  const [ratingsTotalPages, setRatingsTotalPages] = useState(0);
  const [ratingsTotal, setRatingsTotal] = useState(0);
  const [ratingsLoading, setRatingsLoading] = useState(true);

  // Standalone photos state (menu photos only)
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [photosPage, setPhotosPage] = useState(1);
  const [photosTotalPages, setPhotosTotalPages] = useState(0);
  const [photosTotal, setPhotosTotal] = useState(0);
  const [photosLoading, setPhotosLoading] = useState(true);

  // Delete confirmation state
  const [deleteRatingConfirm, setDeleteRatingConfirm] = useState<{ id: number; name: string; photoCount: number } | null>(null);
  const [deletePhotoConfirm, setDeletePhotoConfirm] = useState<{ type: string; id: number } | null>(null);

  useEffect(() => {
    if (activeTab === 'content') {
      loadRatings();
    } else {
      loadPhotos();
    }
  }, [activeTab, ratingsPage, photosPage]);

  const loadRatings = async () => {
    try {
      setRatingsLoading(true);
      const params = new URLSearchParams({
        page: ratingsPage.toString(),
        limit: '20',
      });

      const response = await api.get<RatingsResponse>(`/admin/ratings?${params}`);

      // Fetch review photos for each rating
      const ratingsWithPhotos = await Promise.all(
        (response.ratings || []).map(async (rating) => {
          try {
            const photosResponse = await api.get<PhotosResponse>(`/admin/photos?type=review`);
            const ratingPhotos = photosResponse.photos.filter(p => p.rating_id === rating.id);
            return { ...rating, photos: ratingPhotos };
          } catch (error) {
            console.error(`Failed to load photos for rating ${rating.id}:`, error);
            return { ...rating, photos: [] };
          }
        })
      );

      setRatings(ratingsWithPhotos);
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
        type: 'menu', // Only load menu photos (standalone)
      });

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

  const handleDeleteRating = (ratingId: number, restaurantName: string, photoCount: number) => {
    setDeleteRatingConfirm({ id: ratingId, name: restaurantName, photoCount });
  };

  const confirmDeleteRating = async () => {
    if (!deleteRatingConfirm) return;

    try {
      await api.delete(`/admin/ratings/${deleteRatingConfirm.id}`);
      loadRatings();
    } catch (error) {
      console.error('Failed to delete rating:', error);
      alert('Failed to delete rating');
    }
  };

  const handleDeletePhoto = (photoType: string, photoId: number) => {
    setDeletePhotoConfirm({ type: photoType, id: photoId });
  };

  const confirmDeletePhoto = async () => {
    if (!deletePhotoConfirm) return;

    try {
      await api.delete(`/admin/photos/${deletePhotoConfirm.type}/${deletePhotoConfirm.id}`);
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
            className={`admin-btn admin-btn-sm ${activeTab === 'content' ? '' : 'admin-btn-secondary'}`}
            onClick={() => setActiveTab('content')}
          >
            <MessageSquare size={14} />
            Ratings & Reviews ({ratingsTotal})
          </button>
          <button
            className={`admin-btn admin-btn-sm ${activeTab === 'photos' ? '' : 'admin-btn-secondary'}`}
            onClick={() => setActiveTab('photos')}
          >
            <Camera size={14} />
            Menu Photos ({photosTotal})
          </button>
        </div>
      </div>

      {/* Content Tab */}
      {activeTab === 'content' && (
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
              <div style={{ padding: '20px', display: 'flex', flexDirection: 'column', gap: '16px' }}>
                {ratings.map(rating => (
                  <div key={rating.id} className="admin-card" style={{ padding: '20px', border: '1px solid var(--admin-border)' }}>
                    {/* Header */}
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: '16px' }}>
                      <div>
                        <h3 style={{ fontFamily: 'IBM Plex Mono, monospace', fontSize: '16px', fontWeight: 600, marginBottom: '4px' }}>
                          {rating.restaurant_name}
                        </h3>
                        <div style={{ fontSize: '12px', color: 'var(--admin-text-muted)', fontFamily: 'IBM Plex Mono, monospace' }}>
                          By {rating.username || 'Anonymous'} • {new Date(rating.created_at).toLocaleDateString()}
                        </div>
                      </div>
                      <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                        <span className="admin-badge admin-badge-primary" style={{ fontSize: '16px', padding: '6px 12px' }}>
                          ★ {getAverageRating(rating)}
                        </span>
                        <a
                          href={`/restaurants/${rating.restaurant_id}?rating=${rating.id}`}
                          className="admin-btn-icon"
                          title="View Rating"
                        >
                          <Eye size={16} />
                        </a>
                        <button
                          className="admin-btn-icon admin-btn-danger"
                          onClick={() => handleDeleteRating(rating.id, rating.restaurant_name, rating.photos?.length || 0)}
                          title="Delete Rating"
                        >
                          <Trash2 size={16} />
                        </button>
                      </div>
                    </div>

                    {/* Ratings */}
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px', marginBottom: '16px' }}>
                      <div>
                        <div style={{ fontSize: '11px', color: 'var(--admin-text-muted)', marginBottom: '4px', fontFamily: 'IBM Plex Mono, monospace' }}>
                          FOOD
                        </div>
                        <div style={{ display: 'flex', gap: '2px' }}>
                          {[...Array(5)].map((_, i) => (
                            <Star key={i} size={14} style={{ fill: i < rating.food_rating ? 'var(--admin-warning)' : 'none', color: i < rating.food_rating ? 'var(--admin-warning)' : 'var(--admin-text-muted)' }} />
                          ))}
                        </div>
                      </div>
                      <div>
                        <div style={{ fontSize: '11px', color: 'var(--admin-text-muted)', marginBottom: '4px', fontFamily: 'IBM Plex Mono, monospace' }}>
                          SERVICE
                        </div>
                        <div style={{ display: 'flex', gap: '2px' }}>
                          {[...Array(5)].map((_, i) => (
                            <Star key={i} size={14} style={{ fill: i < rating.service_rating ? 'var(--admin-warning)' : 'none', color: i < rating.service_rating ? 'var(--admin-warning)' : 'var(--admin-text-muted)' }} />
                          ))}
                        </div>
                      </div>
                      <div>
                        <div style={{ fontSize: '11px', color: 'var(--admin-text-muted)', marginBottom: '4px', fontFamily: 'IBM Plex Mono, monospace' }}>
                          AMBIANCE
                        </div>
                        <div style={{ display: 'flex', gap: '2px' }}>
                          {[...Array(5)].map((_, i) => (
                            <Star key={i} size={14} style={{ fill: i < rating.ambiance_rating ? 'var(--admin-warning)' : 'none', color: i < rating.ambiance_rating ? 'var(--admin-warning)' : 'var(--admin-text-muted)' }} />
                          ))}
                        </div>
                      </div>
                    </div>

                    {/* Comment */}
                    {rating.comment && (
                      <div style={{ marginBottom: '16px', padding: '12px', backgroundColor: 'var(--admin-surface-hover)', borderRadius: '4px', borderLeft: '3px solid var(--admin-primary)' }}>
                        <div style={{ fontSize: '13px', lineHeight: '1.6' }}>
                          {rating.comment}
                        </div>
                      </div>
                    )}

                    {/* Photos */}
                    {rating.photos && rating.photos.length > 0 && (
                      <div>
                        <div style={{ fontSize: '11px', color: 'var(--admin-text-muted)', marginBottom: '8px', fontFamily: 'IBM Plex Mono, monospace', display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <Camera size={12} />
                          ATTACHED PHOTOS ({rating.photos.length})
                        </div>
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(120px, 1fr))', gap: '8px' }}>
                          {rating.photos.map(photo => (
                            <div key={photo.id} style={{ position: 'relative', aspectRatio: '1', overflow: 'hidden', borderRadius: '4px', border: '1px solid var(--admin-border)' }}>
                              <img
                                src={photo.filename.startsWith('/') || photo.filename.startsWith('http') ? photo.filename : `/api/uploads/${photo.filename}`}
                                alt={photo.caption || 'Review photo'}
                                style={{ width: '100%', height: '100%', objectFit: 'cover' }}
                                onError={(e) => {
                                  // Fallback to placeholder on error
                                  e.currentTarget.src = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="100" height="100"%3E%3Crect fill="%23141414" width="100" height="100"/%3E%3Ctext fill="%23888" x="50%" y="50%" text-anchor="middle" dy=".3em"%3ENo Image%3C/text%3E%3C/svg%3E';
                                }}
                              />
                              <div style={{ position: 'absolute', top: '4px', right: '4px', display: 'flex', gap: '4px' }}>
                                <a
                                  href={`/restaurants/${rating.restaurant_id}?photo=${photo.id}`}
                                  className="admin-btn-icon"
                                  style={{ backgroundColor: 'rgba(0,0,0,0.7)', padding: '4px' }}
                                  title="View Photo"
                                >
                                  <Eye size={12} />
                                </a>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                ))}
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

      {/* Menu Photos Tab */}
      {activeTab === 'photos' && (
        <div className="admin-card">
          <div className="admin-card-header">
            <h2 className="admin-card-title">
              <Camera size={20} />
              Menu Photos
            </h2>
            <p style={{ fontSize: '12px', color: 'var(--admin-text-muted)', marginTop: '4px' }}>
              Standalone menu photos uploaded to restaurants (review photos are shown with their ratings in the Content tab)
            </p>
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
                                href={photo.restaurant_id ? `/restaurants/${photo.restaurant_id}?photo=${photo.id}` : `/api/uploads/${photo.filename}`}
                                className="admin-btn-icon"
                                title="View Photo"
                                {...(photo.restaurant_id ? {} : { target: '_blank', rel: 'noopener noreferrer' })}
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
      )}

      {/* Delete Rating Confirmation */}
      <ConfirmDialog
        isOpen={deleteRatingConfirm !== null}
        onClose={() => setDeleteRatingConfirm(null)}
        onConfirm={confirmDeleteRating}
        title="Delete Rating"
        message={
          deleteRatingConfirm?.photoCount && deleteRatingConfirm.photoCount > 0
            ? `Delete rating for "${deleteRatingConfirm?.name}" and ${deleteRatingConfirm.photoCount} attached photo(s)? This action cannot be undone.`
            : `Delete rating for "${deleteRatingConfirm?.name}"? This action cannot be undone.`
        }
        confirmText="Delete"
        cancelText="Cancel"
        isDangerous={true}
      />

      {/* Delete Photo Confirmation */}
      <ConfirmDialog
        isOpen={deletePhotoConfirm !== null}
        onClose={() => setDeletePhotoConfirm(null)}
        onConfirm={confirmDeletePhoto}
        title="Delete Photo"
        message={`Delete this ${deletePhotoConfirm?.type} photo? This action cannot be undone.`}
        confirmText="Delete"
        cancelText="Cancel"
        isDangerous={true}
      />
    </div>
  );
}
