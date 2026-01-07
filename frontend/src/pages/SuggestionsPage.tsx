import { useState, useEffect, useCallback } from 'react';
import { Plus, Loader2, XCircle, Clock, ListChecks, Trash2 } from 'lucide-react';
import {
  RestaurantSuggestion,
  CreateSuggestionData,
  getSuggestions,
  createSuggestion,
  convertSuggestion,
  deleteSuggestion,
} from '../services/api';
import { SuggestionForm } from '../components/SuggestionForm';
import { Modal } from '../components/Modal';
import { ReviewConvertModal } from '../components/ReviewConvertModal';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { AlertDialog } from '../components/AlertDialog';
import { UserBadge } from '../components/UserBadge';
import { PermissionGuard } from '../components/PermissionGuard';

type StatusFilter = '' | 'pending' | 'approved' | 'tested' | 'rejected';

export function SuggestionsPage() {
  const [suggestions, setSuggestions] = useState<RestaurantSuggestion[]>([]);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('');
  const [loading, setLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [reviewingId, setReviewingId] = useState<number | null>(null);
  const [reviewingName, setReviewingName] = useState('');
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [alertMessage, setAlertMessage] = useState('');

  const fetchSuggestions = useCallback(async () => {
    try {
      const data = await getSuggestions(statusFilter);
      setSuggestions(data);
    } catch (error) {
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => {
    fetchSuggestions();
  }, [fetchSuggestions]);

  const handleCreate = async (data: CreateSuggestionData) => {
    try {
      await createSuggestion(data);
      setShowAddModal(false);
      fetchSuggestions();
    } catch (error) {
      setAlertMessage('Failed to create suggestion');
    }
  };

  const handleReviewAndConvert = async (data: {
    foodRating: number;
    serviceRating: number;
    ambianceRating: number;
    comment: string;
    description: string;
  }) => {
    if (!reviewingId) return;
    try {
      await convertSuggestion(reviewingId, {
        description: data.description || undefined,
        category_id: undefined,
        food_rating: data.foodRating,
        service_rating: data.serviceRating,
        ambiance_rating: data.ambianceRating,
        comment: data.comment || undefined,
      });
      setReviewingId(null);
      setReviewingName('');
      fetchSuggestions();
      setAlertMessage('Suggestion converted to restaurant successfully!');
    } catch (error) {
      setAlertMessage('Failed to convert suggestion');
    }
  };

  const handleDelete = async (id: number) => {
    setDeletingId(id);
  };

  const confirmDelete = async () => {
    if (!deletingId) return;
    try {
      await deleteSuggestion(deletingId);
      setDeletingId(null);
      fetchSuggestions();
    } catch (error) {
      setAlertMessage('Failed to delete suggestion');
      setDeletingId(null);
    }
  };

  const getStatusBadge = (status: string) => {
    const badges = {
      pending: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      rejected: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
    };
    return badges[status as keyof typeof badges] || '';
  };

  const getStatusIcon = (status: string) => {
    const icons = {
      pending: <Clock className="w-4 h-4" />,
      rejected: <XCircle className="w-4 h-4" />,
    };
    return icons[status as keyof typeof icons] || null;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Restaurant Suggestions</h1>
        <button onClick={() => setShowAddModal(true)} className="btn btn-primary flex items-center gap-2">
          <Plus className="w-5 h-5" />
          Suggest Restaurant
        </button>
      </div>

      {/* Status Filter Tabs */}
      <div className="flex gap-2 mb-6 overflow-x-auto pb-2">
        {[
          { value: '', label: 'All' },
          { value: 'pending', label: 'Pending' },
          { value: 'rejected', label: 'Rejected' },
        ].map((tab) => (
          <button
            key={tab.value}
            onClick={() => setStatusFilter(tab.value as StatusFilter)}
            className={`px-4 py-2 rounded-lg font-medium whitespace-nowrap transition-colors ${
              statusFilter === tab.value
                ? 'bg-blue-500 text-white'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Suggestions List */}
      {suggestions.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-500 dark:text-gray-400 mb-4">
            {statusFilter ? `No ${statusFilter} suggestions.` : 'No suggestions yet.'}
          </p>
          {!statusFilter && (
            <button onClick={() => setShowAddModal(true)} className="btn btn-primary">
              Suggest your first restaurant
            </button>
          )}
        </div>
      ) : (
        <div className="space-y-4">
          {suggestions.map((suggestion) => (
            <div
              key={suggestion.id}
              className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:shadow-lg transition-shadow bg-white dark:bg-gray-800"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <h3 className="text-lg font-semibold">{suggestion.name}</h3>
                    <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${getStatusBadge(suggestion.status)}`}>
                      {getStatusIcon(suggestion.status)}
                      {suggestion.status.toUpperCase()}
                    </span>
                  </div>

                  {suggestion.address && (
                    <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">{suggestion.address}</p>
                  )}

                  <div className="flex flex-wrap gap-2 mb-2">
                    {suggestion.category && (
                      <span className="px-2 py-1 bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 rounded text-xs">
                        {suggestion.category.name}
                      </span>
                    )}
                    {suggestion.food_types?.map((ft) => (
                      <span
                        key={ft.id}
                        className="px-2 py-1 bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded text-xs"
                      >
                        {ft.name}
                      </span>
                    ))}
                  </div>

                  {suggestion.notes && (
                    <p className="text-sm text-gray-600 dark:text-gray-400 italic">{suggestion.notes}</p>
                  )}

                  <div className="flex items-center justify-between mt-2">
                    <p className="text-xs text-gray-500 dark:text-gray-500">
                      Suggested {new Date(suggestion.created_at).toLocaleDateString()}
                    </p>
                    {suggestion.user && (
                      <UserBadge user={suggestion.user} size="sm" />
                    )}
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex flex-col gap-2 ml-4">
                  <PermissionGuard permission="suggestions.convert">
                    {suggestion.status === 'pending' && (
                      <button
                        onClick={() => {
                          setReviewingId(suggestion.id);
                          setReviewingName(suggestion.name);
                        }}
                        className="btn btn-sm bg-green-500 hover:bg-green-600 text-white flex items-center gap-1"
                      >
                        <ListChecks className="w-4 h-4" />
                        Test & Review
                      </button>
                    )}
                  </PermissionGuard>

                  <PermissionGuard permission="suggestions.delete">
                    <button
                      onClick={() => handleDelete(suggestion.id)}
                      className="btn btn-sm bg-gray-500 hover:bg-gray-600 text-white flex items-center gap-1"
                    >
                      <Trash2 className="w-4 h-4" />
                      Delete
                    </button>
                  </PermissionGuard>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Add Suggestion Modal */}
      <Modal isOpen={showAddModal} onClose={() => setShowAddModal(false)} title="Suggest Restaurant">
        <SuggestionForm onSubmit={handleCreate} onCancel={() => setShowAddModal(false)} />
      </Modal>

      {/* Review and Convert Modal */}
      <ReviewConvertModal
        isOpen={reviewingId !== null}
        onClose={() => {
          setReviewingId(null);
          setReviewingName('');
        }}
        onSubmit={handleReviewAndConvert}
        restaurantName={reviewingName}
      />

      <ConfirmDialog
        isOpen={deletingId !== null}
        onClose={() => setDeletingId(null)}
        onConfirm={confirmDelete}
        title="Delete Suggestion"
        message="Are you sure you want to delete this suggestion?"
        confirmText="Delete"
        cancelText="Cancel"
        confirmClassName="bg-red-600 hover:bg-red-700 text-white"
      />

      <AlertDialog
        isOpen={alertMessage !== ''}
        onClose={() => setAlertMessage('')}
        message={alertMessage}
      />
    </div>
  );
}
