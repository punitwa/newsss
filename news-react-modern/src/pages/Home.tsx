import React, { useState, useMemo } from 'react';
import { Link } from 'react-router-dom';
import HeroSection from '@/components/HeroSection';
import NewsCard from '@/components/NewsCard';
import TrendingTopics from '@/components/TrendingTopics';
import { useNews, useNewsSearch } from '@/hooks/useNews';
import { News } from '@/types/news';

interface HomeProps {
  onCategorySelect: (category: string) => void;
  selectedCategory: string;
  currentSearchQuery: string;
}

const Home: React.FC<HomeProps> = ({ onCategorySelect, selectedCategory, currentSearchQuery }) => {
  const { news, isLoading, error } = useNews();
  const { searchResults, isSearching, searchError } = useNewsSearch();
  const [showAllLatestNews, setShowAllLatestNews] = useState<boolean>(false);
  const [selectedTrendingTopic, setSelectedTrendingTopic] = useState<string>('');


  // Category mapping for broader classification
  const getCategoryMapping = (articleCategory: string): string => {
    const category = articleCategory.toLowerCase();
    
    // Direct matches first
    if (category === 'technology') return 'technology';
    if (category === 'business') return 'business'; 
    if (category === 'sports') return 'sports';
    if (category === 'entertainment') return 'entertainment';
    if (category === 'health') return 'health';
    if (category === 'science') return 'science';
    if (category === 'general') return 'general';
    
    // Technology keywords
    if (category.includes('tech') || category.includes('youtube') || category.includes('meta') || 
        category.includes('claude') || category.includes('ai') || category.includes('software') ||
        category.includes('app') || category.includes('digital') || category.includes('internet') ||
        category.includes('replit') || category.includes('playstation') || category.includes('iphone')) {
      return 'technology';
    }
    
    // Business keywords  
    if (category.includes('business') || category.includes('finance') || category.includes('economic') ||
        category.includes('market') || category.includes('startup') || category.includes('investment')) {
      return 'business';
    }
    
    // Sports keywords
    if (category.includes('sport') || category.includes('football') || category.includes('basketball') ||
        category.includes('soccer') || category.includes('baseball') || category.includes('tennis') ||
        category.includes('rsl')) {
      return 'sports';
    }
    
    // Entertainment keywords
    if (category.includes('entertainment') || category.includes('movie') || category.includes('music') ||
        category.includes('celebrity') || category.includes('tv') || category.includes('film') ||
        category.includes('vimeo') || category.includes('social media') || category.includes('bluesky')) {
      return 'entertainment';
    }
    
    // Health keywords
    if (category.includes('health') || category.includes('medical') || category.includes('medicine') ||
        category.includes('hospital') || category.includes('doctor') || category.includes('wellness')) {
      return 'health';
    }
    
    // Science keywords
    if (category.includes('science') || category.includes('research') || category.includes('study') ||
        category.includes('discovery') || category.includes('climate') || category.includes('space') ||
        category.includes('silicon') || category.includes('robotics')) {
      return 'science';
    }
    
    return 'general';
  };

  // Determine which data to display
  const displayedNews = useMemo(() => {
    if (currentSearchQuery) {
      return searchResults;
    }
    
    let filteredNews = news;
    
    // Filter by category
    if (selectedCategory && selectedCategory !== 'Top Stories') {
      const targetCategory = selectedCategory.toLowerCase();
      filteredNews = filteredNews.filter((article: News) => {
        const mappedCategory = getCategoryMapping(article.category);
        return mappedCategory === targetCategory;
      });
    }
    
    // Filter by trending topic
    if (selectedTrendingTopic) {
      filteredNews = filteredNews.filter((article: News) => {
        const searchTerm = selectedTrendingTopic.toLowerCase();
        return (
          article.title?.toLowerCase().includes(searchTerm) ||
          article.content?.toLowerCase().includes(searchTerm) ||
          article.category?.toLowerCase().includes(searchTerm) ||
          article.tags?.some((tag: string) => tag.toLowerCase().includes(searchTerm))
        );
      });
    }
    
    return filteredNews;
  }, [news, searchResults, selectedCategory, currentSearchQuery, selectedTrendingTopic]);

  const topStories = displayedNews.slice(0, 3);
  const latestNews = showAllLatestNews ? displayedNews.slice(3) : displayedNews.slice(3, 8);

  const handleViewAll = () => {
    setShowAllLatestNews(!showAllLatestNews);
  };

  const handleTrendingTopicClick = (topic: string) => {
    setSelectedTrendingTopic(topic);
    onCategorySelect('Top Stories'); // Reset to top stories
  };

  const clearFilters = () => {
    setSelectedTrendingTopic('');
    onCategorySelect('Top Stories'); // Reset to top stories
  };

  const handleSubscribe = () => {
    console.log('Subscribe clicked');
  };

  const currentLoading = isLoading || isSearching;

  return (
    <main>
      <HeroSection />
      
      {/* Main Content */}
      <div className="max-w-container mx-auto px-4 sm:px-6 py-8 sm:py-12">
        <div className="grid lg:grid-cols-4 gap-6 lg:gap-8">
          {/* News Articles */}
          <div className="lg:col-span-3 space-y-6 lg:space-y-8">
            {/* Top Stories */}
            <section>
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-2xl font-heading font-bold">
                  {selectedCategory ? selectedCategory : 'Top Stories'}
                </h2>
                <span className="text-sm text-muted-foreground">
                  {currentLoading ? 'Loading...' : 'Updated now'}
                </span>
              </div>
              
              {currentLoading ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
                  {[...Array(6)].map((_, i) => (
                    <div key={i} className="news-card p-4 sm:p-6 animate-pulse">
                      <div className="bg-muted h-40 sm:h-48 rounded-lg mb-4"></div>
                      <div className="space-y-2">
                        <div className="bg-muted h-4 rounded w-3/4"></div>
                        <div className="bg-muted h-4 rounded w-1/2"></div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : error || searchError ? (
                <div className="text-center py-8">
                  <p className="text-muted-foreground text-sm sm:text-base">Error loading news. Please try again.</p>
                </div>
              ) : topStories.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
                  {topStories.map((article: News) => (
                    <NewsCard key={article.id} news={article} />
                  ))}
                </div>
              ) : (
                <div className="text-center py-8">
                  <p className="text-muted-foreground">No news articles found.</p>
                </div>
              )}
            </section>

            {/* Latest News */}
            {!currentSearchQuery && latestNews.length > 0 && (
              <section>
                <div className="flex items-center justify-between mb-6">
                  <div>
                    <h2 className="text-2xl font-heading font-bold">Latest News</h2>
                    {selectedTrendingTopic && (
                      <div className="flex items-center space-x-2 mt-2">
                        <span className="text-sm text-muted-foreground">Filtered by:</span>
                        <span className="inline-flex items-center space-x-1 text-xs font-medium text-primary bg-primary/10 px-2 py-1 rounded-full">
                          <span>#{selectedTrendingTopic}</span>
                          <button 
                            onClick={clearFilters}
                            className="ml-1 hover:text-primary-hover"
                            title="Clear filter"
                          >
                            ×
                          </button>
                        </span>
                      </div>
                    )}
                  </div>
                  <div className="flex items-center space-x-4">
                    <button 
                      className="text-primary hover:text-primary-hover font-medium text-sm transition-colors"
                      onClick={handleViewAll}
                    >
                      {showAllLatestNews ? 'Show Less' : 'View All'}
                    </button>
                    <Link 
                      to="/all-news"
                      className="text-primary hover:text-primary-hover font-medium text-sm transition-colors"
                    >
                      Browse All Articles →
                    </Link>
                  </div>
                </div>
                
                <div className="space-y-6">
                  {latestNews.map((article: News) => (
                    <div key={article.id} className="news-card p-6">
                      <div className="grid md:grid-cols-4 gap-6">
                        <div className="md:col-span-3 space-y-3">
                          <div className="flex items-center space-x-3">
                            <span className="text-xs font-medium text-primary bg-primary/10 px-2 py-1 rounded-full">
                              {article.category}
                            </span>
                            <span className="text-xs text-muted-foreground">
                              {article.source} • {new Date(article.published_at).toLocaleDateString()}
                            </span>
                          </div>
                          
                          <h3 className="text-xl font-heading font-semibold leading-tight hover:text-primary cursor-pointer transition-colors">
                            {article.title}
                          </h3>
                          
                          <p className="text-muted-foreground leading-relaxed">
                            {article.summary || 'No description available'}
                          </p>
                        </div>
                        
                        {article.image_url && (
                          <div className="md:col-span-1">
                            <img 
                              src={article.image_url} 
                              alt={article.title}
                              className="w-full h-32 md:h-24 object-cover rounded-lg"
                            />
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>

                {/* Browse All Articles Link */}
                <div className="text-center mt-8">
                  <Link 
                    to="/all-news"
                    className="inline-flex items-center bg-primary text-primary-foreground px-6 py-3 rounded-lg font-medium hover:bg-primary-hover transition-colors"
                  >
                    Browse All {news.length} Articles →
                  </Link>
                </div>
              </section>
            )}
          </div>

          {/* Sidebar */}
          <aside className="lg:col-span-1 space-y-4 lg:space-y-6">
            <TrendingTopics onTopicClick={handleTrendingTopicClick} />
            
            {/* Newsletter Signup */}
            <div className="bg-gradient-primary rounded-lg p-4 sm:p-6 text-primary-foreground">
              <h3 className="font-heading font-semibold text-base sm:text-lg mb-2">Stay Updated</h3>
              <p className="text-xs sm:text-sm text-primary-foreground/80 mb-3 sm:mb-4">
                Get the latest news delivered to your inbox daily.
              </p>
              <div className="space-y-2 sm:space-y-3">
                <input 
                  type="email" 
                  placeholder="Enter your email"
                  className="w-full px-3 py-2 rounded-lg bg-primary-foreground/20 border border-primary-foreground/30 text-primary-foreground placeholder:text-primary-foreground/60 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-primary-foreground/50"
                />
                <button 
                  className="w-full bg-accent hover:bg-accent/90 text-accent-foreground py-2 px-4 rounded-lg font-medium text-xs sm:text-sm transition-colors"
                  onClick={handleSubscribe}
                >
                  Subscribe
                </button>
              </div>
            </div>
          </aside>
        </div>
      </div>
    </main>
  );
};

export default Home;