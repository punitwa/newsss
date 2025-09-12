import { Search, TrendingUp, Grid, Sparkles, Zap, Globe } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useState, useRef } from "react";
import { useAuth } from "@/contexts/AuthContext";
import UserMenu from "@/components/auth/UserMenu";
import AuthModal from "@/components/auth/AuthModal";

interface NavigationProps {
  onSearch?: (query: string) => void;
  onCategorySelect?: (category: string) => void;
  onTrendingClick?: () => void;
  selectedCategory?: string;
}

const Navigation = ({ onSearch, onCategorySelect, onTrendingClick, selectedCategory }: NavigationProps) => {
  const location = useLocation();
  const navigate = useNavigate();
  const isAllNewsPage = location.pathname === '/all-news';
  const [searchQuery, setSearchQuery] = useState('');
  const searchInputRef = useRef<HTMLInputElement>(null);
  const { } = useAuth();
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [authModalMode, setAuthModalMode] = useState<'login' | 'register'>('login');
  
  const categories = [
    { name: "Top Stories", icon: Sparkles, color: "from-blue-500 to-cyan-500" }, 
    { name: "Technology", icon: Zap, color: "from-purple-500 to-blue-500" }, 
    { name: "Business", icon: TrendingUp, color: "from-green-500 to-emerald-500" }, 
    { name: "Sports", icon: Grid, color: "from-orange-500 to-red-500" }, 
    { name: "Entertainment", icon: Sparkles, color: "from-pink-500 to-purple-500" }, 
    { name: "Health", icon: Zap, color: "from-teal-500 to-cyan-500" }, 
    { name: "Science", icon: Grid, color: "from-indigo-500 to-purple-500" }
  ];

  const handleSearchSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (searchQuery.trim() && onSearch) {
      onSearch(searchQuery.trim());
    }
  };

  const handleSearchInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
  };

  const handleSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      if (searchQuery.trim() && onSearch) {
        onSearch(searchQuery.trim());
      }
    }
  };

  const handleTrendingClick = () => {
    if (onTrendingClick) {
      onTrendingClick();
    }
    console.log('Trending clicked');
  };

  const handleCategoryClick = (category: string) => {
    if (onCategorySelect) {
      onCategorySelect(category);
    }
  };

  // Authentication handlers
  const openLoginModal = () => {
    setAuthModalMode('login');
    setAuthModalOpen(true);
  };

  const openRegisterModal = () => {
    setAuthModalMode('register');
    setAuthModalOpen(true);
  };

  const closeAuthModal = () => {
    setAuthModalOpen(false);
  };

  return (
    <nav className="sticky top-0 sm:top-4 z-[100] bg-white border-b border-gray-200 shadow-lg">
      <div className="max-w-7xl mx-auto px-3 sm:px-6">
        {/* Main Navigation Bar */}
        <div className="flex items-center justify-between h-16 sm:h-16">
          {/* Logo */}
          <div className="flex items-center space-x-2">
            <div className="bg-gradient-to-r from-blue-500 to-purple-600 p-1.5 sm:p-2.5 rounded-lg shadow-sm">
              <Globe className="h-4 w-4 sm:h-5 sm:w-5 text-white" />
            </div>
            <Link to="/" className="text-lg sm:text-xl font-bold text-gray-900 hover:text-blue-600 transition-colors duration-200">
              WorldBrief
            </Link>
          </div>

          {/* Search Bar - Hidden on mobile, shown on tablet+ */}
          <div className="hidden md:flex flex-1 max-w-md mx-8">
            <form onSubmit={handleSearchSubmit} className="relative w-full">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                ref={searchInputRef}
                type="search"
                placeholder="Search breaking news, trending stories..."
                className="pl-10 pr-4 py-2.5 bg-gray-50 border border-gray-200 rounded-lg focus:border-blue-500 focus:bg-white focus:outline-none focus:ring-1 focus:ring-blue-500/20 transition-all duration-200 text-gray-700 placeholder:text-gray-500"
                value={searchQuery}
                onChange={handleSearchInputChange}
                onKeyDown={handleSearchKeyDown}
              />
            </form>
          </div>

          {/* Enhanced Navigation Buttons */}
          <div className="flex items-center space-x-1">
            <Link to="/all-news">
              <Button
                variant="ghost"
                size="sm"
                className={`group relative min-w-[44px] min-h-[44px] px-3 sm:px-4 py-2 rounded-lg font-medium transition-all duration-200 ${
                  isAllNewsPage
                    ? "bg-blue-100 text-blue-700 border border-blue-200 shadow-sm"
                    : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                }`}
              >
                <Grid className="h-4 w-4 sm:mr-2" />
                <span className="hidden sm:inline ml-1">All News</span>
              </Button>
            </Link>

            <Button
              variant="ghost"
              size="sm"
              className="group relative min-w-[44px] min-h-[44px] px-3 sm:px-4 py-2 rounded-lg font-medium text-gray-600 hover:bg-gray-100 hover:text-gray-900 transition-all duration-200"
              onClick={handleTrendingClick}
            >
              <TrendingUp className="h-4 w-4 sm:mr-2" />
              <span className="hidden sm:inline ml-1">Trending</span>
            </Button>

            {/* Authentication Section */}
            <div className="flex items-center space-x-2 ml-2">
              {/* Always show UserMenu - it handles both authenticated and unauthenticated states */}
              <UserMenu 
                onProfileClick={() => navigate('/profile')}
                onSettingsClick={() => console.log('Settings clicked')}
                onBookmarksClick={() => navigate('/bookmarks')}
                onLoginClick={openLoginModal}
                onRegisterClick={openRegisterModal}
              />
            </div>
          </div>
        </div>

        {/* Enhanced Category Navigation */}
        <div className="flex items-center gap-2 py-2 sm:py-3 overflow-x-auto scrollbar-hide scroll-smooth category-scroll">
          {categories.map((category) => {
            const isActive = selectedCategory === category.name;
            const Icon = category.icon;
            return (
              <Button
                key={category.name}
                variant="ghost"
                size="sm"
                className={`group relative overflow-hidden whitespace-nowrap min-w-[44px] min-h-[40px] px-3 sm:px-4 py-2 sm:py-2.5 rounded-xl font-medium transition-all duration-300 text-xs sm:text-sm flex-shrink-0 ${
                  isActive
                    ? `bg-gradient-to-r ${category.color} text-white shadow-lg hover:shadow-xl transform hover:scale-105`
                    : "text-gray-600 hover:text-gray-900 hover:bg-gray-100/80 hover:shadow-md"
                }`}
                onClick={() => handleCategoryClick(category.name)}
              >
                <Icon className={`h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 transition-transform duration-300 ${isActive ? 'animate-pulse' : 'group-hover:rotate-12'}`} />
                <span className="text-xs sm:text-sm font-medium">{category.name}</span>
                {isActive && (
                  <div className="absolute inset-0 bg-gradient-to-r from-white/20 to-transparent rounded-xl"></div>
                )}
              </Button>
            );
          })}
        </div>

        {/* Mobile Search Bar - Shown below categories on mobile */}
        <div className="md:hidden pb-3">
          <form onSubmit={handleSearchSubmit} className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              type="search"
              placeholder="Search news..."
              className="pl-10 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-lg focus:border-blue-500 focus:bg-white focus:outline-none focus:ring-2 focus:ring-blue-500/20 transition-all duration-200 text-gray-700 placeholder:text-gray-500 text-base"
              value={searchQuery}
              onChange={handleSearchInputChange}
              onKeyDown={handleSearchKeyDown}
            />
          </form>
        </div>
      </div>

      {/* Authentication Modal */}
      <AuthModal
        isOpen={authModalOpen}
        onClose={closeAuthModal}
        initialMode={authModalMode}
        onSuccess={() => {
          console.log('Authentication successful!');
          // Add a small delay to ensure state updates before closing modal
          setTimeout(() => {
            closeAuthModal();
          }, 100);
          // Could add a success toast here
        }}
      />
    </nav>
  );
};

export default Navigation;
