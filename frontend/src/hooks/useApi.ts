import { useQuery, useMutation, useQueryClient, useInfiniteQuery, UseQueryOptions, UseMutationOptions, UseInfiniteQueryOptions } from '@tanstack/react-query';
import * as api from '../services/api';
import type {
  Restaurant,
  Category,
  FoodType,
  Rating,
  RestaurantSuggestion,
  MenuPhoto,
  RestaurantFilters,
  CreateRestaurantData,
  CreateSuggestionData,
  GooglePlaceResult,
  PaginatedResponse,
} from '../services/api';

// Query Keys - centralized for easy cache management
export const queryKeys = {
  restaurants: (filters?: RestaurantFilters) => ['restaurants', filters] as const,
  restaurantsPaginated: (filters?: RestaurantFilters) => ['restaurants', 'paginated', filters] as const,
  restaurant: (id: number) => ['restaurant', id] as const,
  categories: () => ['categories'] as const,
  foodTypes: () => ['foodTypes'] as const,
  ratings: (restaurantId: number) => ['ratings', restaurantId] as const,
  suggestions: (status?: string) => ['suggestions', status] as const,
  suggestion: (id: number) => ['suggestion', id] as const,
  menuPhotos: (restaurantId: number) => ['menuPhotos', restaurantId] as const,
  globalSearch: (query: string) => ['globalSearch', query] as const,
  placesSearch: (query: string) => ['placesSearch', query] as const,
};

// ============= RESTAURANTS =============

export const useRestaurants = (
  filters?: RestaurantFilters,
  options?: Omit<UseQueryOptions<Restaurant[], Error>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: queryKeys.restaurants(filters),
    queryFn: () => api.getRestaurants(filters),
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
    ...options,
  });
};

export const useRestaurantsPaginated = (
  filters?: RestaurantFilters,
  options?: Omit<UseInfiniteQueryOptions<PaginatedResponse<Restaurant>, Error>, 'queryKey' | 'queryFn' | 'getNextPageParam' | 'initialPageParam'>
) => {
  return useInfiniteQuery({
    queryKey: queryKeys.restaurantsPaginated(filters),
    queryFn: ({ pageParam }) => api.getRestaurantsPaginated(filters, { cursor: pageParam as string | undefined, limit: 20 }),
    initialPageParam: undefined,
    getNextPageParam: (lastPage) => {
      return lastPage.has_more ? lastPage.next_cursor : undefined;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    ...options,
  });
};

export const useRestaurant = (
  id: number,
  options?: Omit<UseQueryOptions<Restaurant, Error>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: queryKeys.restaurant(id),
    queryFn: () => api.getRestaurant(id),
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    enabled: id > 0,
    ...options,
  });
};

export const useCreateRestaurant = (
  options?: UseMutationOptions<Restaurant, Error, CreateRestaurantData>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: api.createRestaurant,
    onSuccess: (newRestaurant) => {
      // Invalidate all restaurant queries to refetch
      queryClient.invalidateQueries({ queryKey: ['restaurants'] });
      // Optionally set the new restaurant in cache
      queryClient.setQueryData(queryKeys.restaurant(newRestaurant.id), newRestaurant);
    },
    ...options,
  });
};

export const useUpdateRestaurant = (
  options?: UseMutationOptions<Restaurant, Error, { id: number; data: CreateRestaurantData }, { previousRestaurant: Restaurant | undefined }>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }) => api.updateRestaurant(id, data),
    onMutate: async ({ id, data }): Promise<{ previousRestaurant: Restaurant | undefined }> => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: queryKeys.restaurant(id) });

      // Snapshot previous value
      const previousRestaurant = queryClient.getQueryData<Restaurant>(queryKeys.restaurant(id));

      // Optimistically update
      if (previousRestaurant) {
        queryClient.setQueryData<Restaurant>(queryKeys.restaurant(id), {
          ...previousRestaurant,
          ...data,
        });
      }

      return { previousRestaurant };
    },
    onError: (_err, { id }, context) => {
      // Rollback on error
      if (context?.previousRestaurant) {
        queryClient.setQueryData(queryKeys.restaurant(id), context.previousRestaurant);
      }
    },
    onSuccess: (updatedRestaurant, { id }) => {
      // Update cache with server response
      queryClient.setQueryData(queryKeys.restaurant(id), updatedRestaurant);
      // Invalidate restaurant lists
      queryClient.invalidateQueries({ queryKey: ['restaurants'] });
    },
    ...options,
  });
};

export const useDeleteRestaurant = (
  options?: UseMutationOptions<void, Error, number, { previousRestaurants: [readonly unknown[], Restaurant[] | undefined][] }>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: api.deleteRestaurant,
    onMutate: async (id): Promise<{ previousRestaurants: [readonly unknown[], Restaurant[] | undefined][] }> => {
      // Cancel related queries
      await queryClient.cancelQueries({ queryKey: ['restaurants'] });

      // Snapshot
      const previousRestaurants = queryClient.getQueriesData<Restaurant[]>({ queryKey: ['restaurants'] });

      // Optimistically remove from all restaurant lists
      queryClient.setQueriesData<Restaurant[]>({ queryKey: ['restaurants'] }, (old) =>
        old ? old.filter((r) => r.id !== id) : []
      );

      return { previousRestaurants };
    },
    onError: (_err, _id, context) => {
      // Rollback
      if (context?.previousRestaurants) {
        context.previousRestaurants.forEach(([key, data]) => {
          queryClient.setQueryData(key, data);
        });
      }
    },
    onSuccess: (_, id) => {
      // Remove from cache and invalidate
      queryClient.removeQueries({ queryKey: queryKeys.restaurant(id) });
      queryClient.invalidateQueries({ queryKey: ['restaurants'] });
    },
    ...options,
  });
};

// ============= CATEGORIES =============

export const useCategories = (
  options?: Omit<UseQueryOptions<Category[], Error>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: queryKeys.categories(),
    queryFn: api.getCategories,
    staleTime: 10 * 60 * 1000, // 10 minutes - categories change rarely
    gcTime: 30 * 60 * 1000, // 30 minutes
    ...options,
  });
};

export const useCreateCategory = (
  options?: UseMutationOptions<Category, Error, string>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: api.createCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories() });
    },
    ...options,
  });
};

export const useUpdateCategory = (
  options?: UseMutationOptions<Category, Error, { id: number; name: string }>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, name }) => api.updateCategory(id, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories() });
    },
    ...options,
  });
};

export const useDeleteCategory = (
  options?: UseMutationOptions<void, Error, number>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: api.deleteCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.categories() });
    },
    ...options,
  });
};

// ============= FOOD TYPES =============

export const useFoodTypes = (
  options?: Omit<UseQueryOptions<FoodType[], Error>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: queryKeys.foodTypes(),
    queryFn: api.getFoodTypes,
    staleTime: 10 * 60 * 1000, // 10 minutes
    gcTime: 30 * 60 * 1000,
    ...options,
  });
};

export const useCreateFoodType = (
  options?: UseMutationOptions<FoodType, Error, string>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: api.createFoodType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.foodTypes() });
    },
    ...options,
  });
};

export const useUpdateFoodType = (
  options?: UseMutationOptions<FoodType, Error, { id: number; name: string }>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, name }) => api.updateFoodType(id, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.foodTypes() });
    },
    ...options,
  });
};

export const useDeleteFoodType = (
  options?: UseMutationOptions<void, Error, number>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: api.deleteFoodType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.foodTypes() });
    },
    ...options,
  });
};

// ============= RATINGS =============

export const useRatings = (
  restaurantId: number,
  options?: Omit<UseQueryOptions<Rating[], Error>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: queryKeys.ratings(restaurantId),
    queryFn: () => api.getRatings(restaurantId),
    staleTime: 2 * 60 * 1000, // 2 minutes
    gcTime: 10 * 60 * 1000,
    enabled: restaurantId > 0,
    ...options,
  });
};

export const useCreateRating = (
  options?: UseMutationOptions<
    Rating,
    Error,
    {
      restaurant_id: number;
      food_rating: number;
      service_rating: number;
      ambiance_rating: number;
      comment?: string;
    }
  >
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: api.createRating,
    onSuccess: (_, variables) => {
      // Invalidate ratings for this restaurant
      queryClient.invalidateQueries({ queryKey: queryKeys.ratings(variables.restaurant_id) });
      // Invalidate restaurant to update avg_rating
      queryClient.invalidateQueries({ queryKey: queryKeys.restaurant(variables.restaurant_id) });
      queryClient.invalidateQueries({ queryKey: ['restaurants'] });
    },
    ...options,
  });
};

export const useDeleteRating = (
  options?: UseMutationOptions<void, Error, { id: number; restaurantId: number }>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id }) => api.deleteRating(id),
    onSuccess: (_, { restaurantId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ratings(restaurantId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.restaurant(restaurantId) });
      queryClient.invalidateQueries({ queryKey: ['restaurants'] });
    },
    ...options,
  });
};

// ============= SUGGESTIONS =============

export const useSuggestions = (
  status?: string,
  options?: Omit<UseQueryOptions<RestaurantSuggestion[], Error>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: queryKeys.suggestions(status),
    queryFn: () => api.getSuggestions(status),
    staleTime: 3 * 60 * 1000, // 3 minutes
    gcTime: 10 * 60 * 1000,
    ...options,
  });
};

export const useSuggestion = (
  id: number,
  options?: Omit<UseQueryOptions<RestaurantSuggestion, Error>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: queryKeys.suggestion(id),
    queryFn: () => api.getSuggestion(id),
    staleTime: 3 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    enabled: id > 0,
    ...options,
  });
};

export const useCreateSuggestion = (
  options?: UseMutationOptions<RestaurantSuggestion, Error, CreateSuggestionData>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: api.createSuggestion,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['suggestions'], refetchType: 'active' });
      queryClient.invalidateQueries({ queryKey: ['restaurants'], refetchType: 'active' });
    },
    ...options,
  });
};

export const useUpdateSuggestionStatus = (
  options?: UseMutationOptions<RestaurantSuggestion, Error, { id: number; status: string }>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, status }) => api.updateSuggestionStatus(id, status),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['suggestions'] });
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestion(id) });
    },
    ...options,
  });
};

export const useConvertSuggestion = (
  options?: UseMutationOptions<
    { restaurant_id: number; message: string },
    Error,
    {
      id: number;
      data: {
        description?: string;
        category_id?: number;
        food_rating: number;
        service_rating: number;
        ambiance_rating: number;
        comment?: string;
      };
    }
  >
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }) => api.convertSuggestion(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['suggestions'] });
      queryClient.invalidateQueries({ queryKey: ['restaurants'] });
    },
    ...options,
  });
};

export const useDeleteSuggestion = (
  options?: UseMutationOptions<void, Error, number>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: api.deleteSuggestion,
    onSuccess: (_, id) => {
      queryClient.removeQueries({ queryKey: queryKeys.suggestion(id) });
      queryClient.invalidateQueries({ queryKey: ['suggestions'] });
    },
    ...options,
  });
};

// ============= MENU PHOTOS =============

export const useMenuPhotos = (
  restaurantId: number,
  options?: Omit<UseQueryOptions<MenuPhoto[], Error>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: queryKeys.menuPhotos(restaurantId),
    queryFn: () => api.getMenuPhotos(restaurantId),
    staleTime: 5 * 60 * 1000,
    gcTime: 15 * 60 * 1000,
    enabled: restaurantId > 0,
    ...options,
  });
};

export const useUploadMenuPhoto = (
  options?: UseMutationOptions<
    { photo: MenuPhoto },
    Error,
    { restaurantId: number; photo: File; caption: string }
  >
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ restaurantId, photo, caption }) =>
      api.uploadMenuPhoto(restaurantId, photo, caption),
    onSuccess: (_, { restaurantId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.menuPhotos(restaurantId) });
    },
    ...options,
  });
};

export const useUpdatePhotoCaption = (
  options?: UseMutationOptions<MenuPhoto, Error, { id: number; caption: string; restaurantId: number }>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, caption }) => api.updatePhotoCaption(id, caption),
    onSuccess: (_, { restaurantId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.menuPhotos(restaurantId) });
    },
    ...options,
  });
};

export const useDeleteMenuPhoto = (
  options?: UseMutationOptions<void, Error, { id: number; restaurantId: number }>
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id }) => api.deleteMenuPhoto(id),
    onSuccess: (_, { restaurantId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.menuPhotos(restaurantId) });
    },
    ...options,
  });
};

// ============= SEARCH =============

export const useGlobalSearch = (
  query: string,
  options?: Omit<UseQueryOptions<Restaurant[], Error>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: queryKeys.globalSearch(query),
    queryFn: () => api.globalSearch(query),
    enabled: query.length >= 2, // Only search if query has at least 2 characters
    staleTime: 1 * 60 * 1000, // 1 minute
    gcTime: 5 * 60 * 1000,
    ...options,
  });
};

export const useSearchPlaces = (
  query: string,
  options?: Omit<UseQueryOptions<GooglePlaceResult[], Error>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: queryKeys.placesSearch(query),
    queryFn: () => api.searchPlaces(query),
    enabled: query.length >= 2,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    ...options,
  });
};
