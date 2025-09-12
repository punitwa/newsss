import axios from 'axios';
import { NewsResponse, CategoriesResponse, SearchResponse, News } from '@/types/news';

const API_BASE_URL = 'http://localhost:8080/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error);
    if (error.code === 'ECONNREFUSED') {
      throw new Error('Cannot connect to news server. Please make sure the Go server is running on http://localhost:8082');
    }
    throw error;
  }
);

export const newsApi = {
  // Get all news articles
  getAllNews: async (page?: number, limit?: number): Promise<NewsResponse> => {
    const params = new URLSearchParams();
    if (page) params.append('page', page.toString());
    if (limit) params.append('limit', limit.toString());

    const url = params.toString() ? `/news?${params.toString()}` : '/news';
    const response = await api.get<NewsResponse>(url);
    return response.data;
  },

  // Get news article by ID
  getNewsById: async (id: string): Promise<{ data: News; message: string }> => {
    const response = await api.get(`/news/${id}`);
    return response.data;
  },

  // Get all categories
  getCategories: async (): Promise<CategoriesResponse> => {
    const response = await api.get<CategoriesResponse>('/news/categories');
    return response.data;
  },

  // Search news articles
  searchNews: async (query: string): Promise<SearchResponse> => {
    const response = await api.get<SearchResponse>(`/news/search?q=${encodeURIComponent(query)}`);
    return response.data;
  },

  // Add news article (for testing)
  addNews: async (news: Partial<News>): Promise<{ data: News; message: string }> => {
    const response = await api.post('/news', news);
    return response.data;
  },

  // Check API health
  checkHealth: async (): Promise<{ status: string; services: any }> => {
    const response = await api.get('/../health'); // Go up one level to reach /health
    return response.data;
  },

  // Enhanced top stories with scoring
  getEnhancedTopStories: async (limit?: number): Promise<NewsResponse> => {
    const params = new URLSearchParams();
    if (limit) params.append('limit', limit.toString());

    const url = params.toString() ? `/news/top-stories?${params.toString()}` : '/news/top-stories';
    const response = await api.get<NewsResponse>(url);
    return response.data;
  },

  // Engagement tracking methods
  trackEngagement: async (articleId: string, type: 'view' | 'click' | 'share'): Promise<{ message: string }> => {
    const response = await api.post(`/news/${articleId}/track/${type}`);
    return response.data;
  },

  trackReadTime: async (articleId: string, readTime: number): Promise<{ message: string }> => {
    const response = await api.post(`/news/${articleId}/track/read-time`, { read_time: readTime });
    return response.data;
  },

  // Get article score
  getArticleScore: async (articleId: string): Promise<any> => {
    const response = await api.get(`/news/${articleId}/score`);
    return response.data;
  },

  // Get top scored articles
  getTopScoredArticles: async (limit?: number, minScore?: number): Promise<any> => {
    const params = new URLSearchParams();
    if (limit) params.append('limit', limit.toString());
    if (minScore) params.append('min_score', minScore.toString());

    const url = params.toString() ? `/news/scores/top?${params.toString()}` : '/news/scores/top';
    const response = await api.get(url);
    return response.data;
  },

  // Analytics endpoints
  getEngagementAnalytics: async (period?: string): Promise<any> => {
    const params = new URLSearchParams();
    if (period) params.append('period', period);

    const url = params.toString() ? `/news/analytics/engagement?${params.toString()}` : '/news/analytics/engagement';
    const response = await api.get(url);
    return response.data;
  },

  getSourceAnalytics: async (): Promise<any> => {
    const response = await api.get('/news/analytics/sources');
    return response.data;
  },
};

export default newsApi;
