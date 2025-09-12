export interface News {
  id: string;
  title: string;
  content: string;
  summary: string;
  url: string;
  image_url: string;
  author: string;
  source: string;
  category: string;
  tags: string[];
  published_at: string;
  created_at: string;
}

export interface NewsResponse {
  data: News[];
  total?: number;
  message?: string;
  meta?: {
    pagination: {
      page: number;
      limit: number;
      total: number;
      pages: number;
      has_next: boolean;
      has_prev: boolean;
    };
  };
  request_id?: string;
  timestamp?: string;
}

export interface CategoriesResponse {
  data: string[];
  total: number;
  message: string;
}

export interface SearchResponse {
  data: News[] | null;
  meta?: {
    pagination?: {
      total: number;
      pages: number;
      page: number;
      limit: number;
      has_next: boolean;
      has_prev: boolean;
    };
  };
  request_id?: string;
  timestamp?: string;
}

export interface ApiError {
  error: string;
  message?: string;
}
