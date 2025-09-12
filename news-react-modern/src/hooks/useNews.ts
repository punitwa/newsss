import { useState, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { newsApi } from '@/services/api';
import { News } from '@/types/news';

export const useNews = () => {
  const queryClient = useQueryClient();

  const {
    data: newsResponse,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['news'],
    queryFn: () => newsApi.getAllNews(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 3,
  });

  const {
    data: categoriesResponse,
    isLoading: categoriesLoading,
  } = useQuery({
    queryKey: ['categories'],
    queryFn: () => newsApi.getCategories(),
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

  // Use all news articles from the API response (API handles filtering)
  const newsArticles = newsResponse?.data || [];

  return {
    news: newsArticles,
    categories: categoriesResponse?.data || [],
    isLoading: isLoading || categoriesLoading,
    error: error?.message || null,
    refetch,
    invalidateNews: () => queryClient.invalidateQueries({ queryKey: ['news'] }),
  };
};

export const useNewsSearch = () => {
  const [searchResults, setSearchResults] = useState<News[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [searchError, setSearchError] = useState<string | null>(null);

  const searchNews = useCallback(async (query: string) => {
    try {
      setIsSearching(true);
      setSearchError(null);
      
      const response = await newsApi.searchNews(query);
      setSearchResults(response.data || []);
    } catch (err) {
      console.error('Search error:', err);
      setSearchError('Failed to search news. Please try again.');
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  }, []);

  const clearSearch = useCallback(() => {
    setSearchResults([]);
    setSearchError(null);
  }, []);

  return {
    searchResults,
    isSearching,
    searchError,
    searchNews,
    clearSearch,
  };
};
