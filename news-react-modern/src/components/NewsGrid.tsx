import React from 'react';
import NewsCard from './NewsCard';
import LoadingSpinner from './LoadingSpinner';
import ErrorMessage from './ErrorMessage';
import { News } from '@/types/news';
import { FileText } from 'lucide-react';

interface NewsGridProps {
  news: News[];
  loading: boolean;
  error: string | null;
  onRetry?: () => void;
  searchQuery?: string;
  category?: string;
}

const NewsGrid: React.FC<NewsGridProps> = ({
  news,
  loading,
  error,
  onRetry,
  searchQuery,
  category,
}) => {
  if (loading && news.length === 0) {
    return <LoadingSpinner />;
  }

  if (error && news.length === 0) {
    return <ErrorMessage message={error} onRetry={onRetry} />;
  }

  if (!loading && news.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center p-12 text-center">
        <FileText className="h-16 w-16 text-muted-foreground mb-4" />
        <h3 className="text-xl font-semibold text-foreground mb-2">
          No articles found
        </h3>
        <p className="text-muted-foreground max-w-md">
          {searchQuery 
            ? `No articles match your search for "${searchQuery}"`
            : category 
              ? `No articles found in the ${category} category`
              : 'No articles available at the moment'
          }
        </p>
        {onRetry && (
          <button
            onClick={onRetry}
            className="mt-4 text-primary hover:text-primary/80 transition-colors"
          >
            Try refreshing the page
          </button>
        )}
      </div>
    );
  }

  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {news.map((article) => (
          <NewsCard key={article.id} news={article} />
        ))}
      </div>
      
      {loading && news.length > 0 && (
        <div className="flex justify-center mt-8">
          <LoadingSpinner message="Loading more articles..." size="sm" />
        </div>
      )}
    </>
  );
};

export default NewsGrid;
