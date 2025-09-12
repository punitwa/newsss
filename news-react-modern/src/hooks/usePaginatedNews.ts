import { useQuery } from '@tanstack/react-query';
import { newsApi } from '@/services/api';

export const usePaginatedNews = (page: number = 1, limit: number = 12) => {
  const {
    data: newsResponse,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['paginatedNews', page, limit],
    queryFn: () => newsApi.getAllNews(page, limit),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 3,
  });

  // Use all news articles from the API response (API handles filtering)
  const newsArticles = newsResponse?.data || [];

  return {
    news: newsArticles,
    total: newsResponse?.meta?.pagination?.total || newsResponse?.total || 0,
    totalPages: newsResponse?.meta?.pagination?.pages || 0,
    currentPage: newsResponse?.meta?.pagination?.page || page,
    isLoading,
    error: error?.message || null,
    refetch,
  };
};
