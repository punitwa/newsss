import React, { useState, useEffect } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from 'next-themes';
import { BrowserRouter, Routes, Route, useNavigate, useLocation } from 'react-router-dom';
import { AuthProvider } from '@/contexts/AuthContext';
import { BookmarkProvider } from '@/components/BookmarkProvider';
import ErrorBoundary from '@/components/ErrorBoundary';
import Navigation from '@/components/Navigation';
import Footer from '@/components/Footer';
import Home from '@/pages/Home';
import AllNews from '@/pages/AllNews';
import SearchResults from '@/pages/SearchResults';
import Profile from '@/pages/Profile';
import Bookmarks from '@/pages/Bookmarks';
import { useNewsSearch } from '@/hooks/useNews';
import { Toaster } from 'sonner';

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

const NewsApp: React.FC = () => {
  const { clearSearch } = useNewsSearch();
  const [currentSearchQuery, setCurrentSearchQuery] = useState<string>('');
  const [selectedCategory, setSelectedCategory] = useState<string>('Top Stories');
  const navigate = useNavigate();
  const location = useLocation();

  // Clear search when navigating to home page
  useEffect(() => {
    if (location.pathname === '/' && currentSearchQuery) {
      setCurrentSearchQuery('');
      clearSearch();
    }
  }, [location.pathname, currentSearchQuery, clearSearch]);

  const handleSearch = async (query: string) => {
    if (!query.trim()) {
      setCurrentSearchQuery('');
      clearSearch();
      return;
    }
    setCurrentSearchQuery(query);
    setSelectedCategory('Top Stories'); // Reset category when searching
    // Navigate to search results page
    navigate(`/search?q=${encodeURIComponent(query.trim())}`);
  };

  const handleCategorySelect = (category: string) => {
    if (currentSearchQuery) {
      setCurrentSearchQuery('');
      clearSearch();
    }
    setSelectedCategory(category);
  };

  const handleTrendingClick = () => {
    // Scroll to trending topics section
    const trendingElement = document.querySelector('[data-trending-section]');
    if (trendingElement) {
      trendingElement.scrollIntoView({ 
        behavior: 'smooth',
        block: 'start'
      });
    } else {
      // If no trending section found, scroll to a reasonable position
      window.scrollTo({
        top: window.innerHeight * 0.8,
        behavior: 'smooth'
      });
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <Navigation 
        onSearch={handleSearch}
        onCategorySelect={handleCategorySelect}
        onTrendingClick={handleTrendingClick}
        selectedCategory={selectedCategory}
      />
      
      <Routes>
        <Route 
          path="/" 
          element={
            <Home 
              onCategorySelect={handleCategorySelect}
              selectedCategory={selectedCategory}
              currentSearchQuery={currentSearchQuery}
            />
          } 
        />
        <Route path="/all-news" element={<AllNews />} />
        <Route path="/search" element={<SearchResults />} />
        <Route path="/profile" element={<Profile />} />
        <Route path="/bookmarks" element={<Bookmarks />} />
      </Routes>

      <Footer />

      <Toaster 
        position="top-right"
        richColors
        closeButton
      />
    </div>
  );
};

const App: React.FC = () => {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <BookmarkProvider>
            <ThemeProvider attribute="class" defaultTheme="light" enableSystem>
              <BrowserRouter>
                <NewsApp />
              </BrowserRouter>
            </ThemeProvider>
          </BookmarkProvider>
        </AuthProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  );
};

export default App;
