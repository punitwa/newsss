import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { authApi } from '@/services/authApi';
import { News } from '@/types/news';
import NewsCard from '@/components/NewsCard';
import LoadingSpinner from '@/components/LoadingSpinner';
import ErrorMessage from '@/components/ErrorMessage';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Heart, BookOpen, ArrowLeft, RefreshCw } from 'lucide-react';
import { Link } from 'react-router-dom';

const Bookmarks = () => {
  const { authState, loadBookmarks } = useAuth();
  const [bookmarkedArticles, setBookmarkedArticles] = useState<News[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Load bookmarked articles
  const fetchBookmarkedArticles = async () => {
    if (!authState.isAuthenticated) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await authApi.getBookmarks();
      
      if (response.success && response.data) {
        // The API returns nested structure: { data: { data: [...], total, page, limit } }
        const responseData = response.data as any;
        const bookmarksData = responseData.data || response.data;
        
        // The bookmarksData is an object with { data: [...], total, page, limit }
        // We need to access the actual array inside
        const bookmarksArray = bookmarksData.data || bookmarksData;
        const articles = Array.isArray(bookmarksArray) ? bookmarksArray : [];
        
        // Transform the bookmark data to match NewsCard expectations
        const transformedArticles = articles.map((bookmark: any) => {
          // If it's already a news object, return as-is
          if (bookmark.title) {
            return bookmark;
          }
          // If it's a bookmark object with nested news, extract the news
          return bookmark.news || bookmark;
        });
        
        setBookmarkedArticles(transformedArticles);
      } else {
        setError(response.error || 'Failed to load bookmarks');
      }
    } catch (err: any) {
      setError(err.message || 'An error occurred while loading bookmarks');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBookmarkedArticles();
  }, [authState.isAuthenticated]);

  // Refresh bookmarks
  const handleRefresh = async () => {
    await loadBookmarks();
    await fetchBookmarkedArticles();
  };

  // Handle bookmark changes (when user bookmarks/unbookmarks)
  const handleBookmarkChange = () => {
    // Refresh the bookmarks data to reflect changes
    handleRefresh();
  };


  // Not authenticated
  if (!authState.isAuthenticated) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-blue-50/30">
        <div className="container mx-auto px-4 py-8">
          <div className="max-w-2xl mx-auto">
            <Card className="border-0 shadow-xl bg-white/70 backdrop-blur-sm">
              <CardContent className="pt-6">
                <div className="text-center">
                  <div className="inline-flex items-center justify-center w-16 h-16 bg-red-100 rounded-full mb-4">
                    <Heart className="h-8 w-8 text-red-600" />
                  </div>
                  <h3 className="text-xl font-semibold text-gray-900 mb-2">Login Required</h3>
                  <p className="text-gray-600 mb-6">
                    Please log in to view your bookmarked articles.
                  </p>
                  <div className="flex gap-3 justify-center">
                    <Link to="/">
                      <Button variant="outline">
                        <ArrowLeft className="h-4 w-4 mr-2" />
                        Back to Home
                      </Button>
                    </Link>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-blue-50/30">
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-7xl mx-auto">
          {/* Header */}
          <div className="mb-8">
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-3 mb-2">
                  <div className="p-2 bg-red-100 rounded-lg">
                    <Heart className="h-6 w-6 text-red-600" />
                  </div>
                  <h1 className="text-3xl font-bold text-gray-900">My Bookmarks</h1>
                </div>
                <p className="text-gray-600">
                  Your saved articles for later reading
                </p>
              </div>
              <div className="flex gap-3">
                <Button
                  variant="outline"
                  onClick={handleRefresh}
                  disabled={loading}
                  className="hover:bg-gray-50"
                >
                  <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
                  Refresh
                </Button>
                <Link to="/">
                  <Button variant="outline">
                    <ArrowLeft className="h-4 w-4 mr-2" />
                    Back to Home
                  </Button>
                </Link>
              </div>
            </div>
          </div>

          {/* Content */}
          {loading ? (
            <div className="flex justify-center py-12">
              <LoadingSpinner message="Loading your bookmarks..." />
            </div>
          ) : error ? (
            <ErrorMessage 
              message={error} 
              onRetry={fetchBookmarkedArticles}
              title="Failed to load bookmarks"
            />
          ) : bookmarkedArticles.length === 0 ? (
            <Card className="border-0 shadow-lg bg-white/70 backdrop-blur-sm">
              <CardContent className="py-12">
                <div className="text-center">
                  <div className="inline-flex items-center justify-center w-16 h-16 bg-gray-100 rounded-full mb-4">
                    <BookOpen className="h-8 w-8 text-gray-400" />
                  </div>
                  <h3 className="text-xl font-semibold text-gray-900 mb-2">No Bookmarks Yet</h3>
                  <p className="text-gray-600 mb-6 max-w-md mx-auto">
                    Start exploring articles and bookmark the ones you want to read later. 
                    Look for the heart icon on any article card.
                  </p>
                  <Link to="/">
                    <Button className="bg-blue-600 hover:bg-blue-700">
                      <ArrowLeft className="h-4 w-4 mr-2" />
                      Explore Articles
                    </Button>
                  </Link>
                </div>
              </CardContent>
            </Card>
          ) : (
            <div>
              {/* Stats */}
              <Card className="border-0 shadow-lg bg-white/70 backdrop-blur-sm mb-8">
                <CardContent className="py-6">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className="text-center">
                        <div className="text-2xl font-bold text-gray-900">
                          {bookmarkedArticles.length}
                        </div>
                        <div className="text-sm text-gray-600">
                          {bookmarkedArticles.length === 1 ? 'Article' : 'Articles'}
                        </div>
                      </div>
                      <div className="h-8 w-px bg-gray-300"></div>
                      <div className="text-sm text-gray-600">
                        Keep track of articles you want to read later
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Articles Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {bookmarkedArticles.map((article) => (
                  <NewsCard 
                    key={article.id} 
                    news={article}
                    onBookmarkChange={handleBookmarkChange}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Bookmarks;
