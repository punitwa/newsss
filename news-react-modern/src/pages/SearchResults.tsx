import React from 'react';
import { ArrowLeft, Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import NewsCard from '@/components/NewsCard';
import { Link, useSearchParams } from 'react-router-dom';
import { useNewsSearch } from '@/hooks/useNews';
import { useEffect } from 'react';

const SearchResults: React.FC = () => {
  const [searchParams] = useSearchParams();
  const query = searchParams.get('q') || '';
  const { searchResults, isSearching, searchError, searchNews } = useNewsSearch();

  useEffect(() => {
    if (query) {
      searchNews(query);
    }
  }, [query, searchNews]);

  if (isSearching) {
    return (
      <div className="min-h-screen bg-background">
        <div className="max-w-container mx-auto px-4 py-12">
          {/* Header */}
          <div className="mb-8">
            <Link to="/" className="inline-flex items-center text-primary hover:text-primary-hover mb-4">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Home
            </Link>
            <h1 className="text-3xl font-heading font-bold">Search Results</h1>
            <p className="text-muted-foreground mt-1">Searching for "{query}"...</p>
          </div>

          {/* Loading State */}
          <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {[...Array(12)].map((_, i) => (
              <div key={i} className="news-card p-6 animate-pulse">
                <div className="bg-muted h-4 rounded w-3/4 mb-2"></div>
                <div className="bg-muted h-4 rounded w-1/2 mb-4"></div>
                <div className="space-y-2">
                  <div className="bg-muted h-3 rounded w-full"></div>
                  <div className="bg-muted h-3 rounded w-2/3"></div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (searchError) {
    return (
      <div className="min-h-screen bg-background">
        <div className="max-w-container mx-auto px-4 py-12">
          <Link to="/" className="inline-flex items-center text-primary hover:text-primary-hover mb-4">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Home
          </Link>
          <div className="text-center py-12">
            <Search className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
            <h1 className="text-2xl font-heading font-bold mb-4">Search Error</h1>
            <p className="text-muted-foreground mb-6">{searchError}</p>
            <Button onClick={() => window.location.reload()}>
              Try Again
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-container mx-auto px-4 py-12">
        {/* Header */}
        <div className="mb-8">
          <Link to="/" className="inline-flex items-center text-primary hover:text-primary-hover mb-4 transition-colors">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Home
          </Link>
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-heading font-bold">Search Results</h1>
              <p className="text-muted-foreground mt-1">
                {searchResults.length > 0 
                  ? `Found ${searchResults.length} results for "${query}"`
                  : `No results found for "${query}"`
                }
              </p>
            </div>
          </div>
        </div>

        {/* Search Results */}
        {searchResults.length > 0 ? (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 mb-12">
            {searchResults.map((article) => (
              <NewsCard key={article.id} news={article} />
            ))}
          </div>
        ) : (
          <div className="text-center py-12">
            <Search className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
            <h2 className="text-xl font-heading font-semibold mb-4">No Articles Found</h2>
            <p className="text-muted-foreground mb-6">
              We couldn't find any articles matching "{query}". Try different keywords or browse our latest news.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link to="/all-news">
                <Button variant="outline">
                  Browse All News
                </Button>
              </Link>
              <Link to="/">
                <Button>
                  Back to Home
                </Button>
              </Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default SearchResults;
