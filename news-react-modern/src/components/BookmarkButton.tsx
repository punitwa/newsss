import { useState } from 'react';
import { Heart } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/contexts/AuthContext';
import { useBookmarkContext } from '@/components/BookmarkProvider';
import { toast } from 'sonner';

interface BookmarkButtonProps {
  articleId: string;
  articleTitle?: string;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'ghost' | 'outline' | 'default';
  showText?: boolean;
  onAuthRequired?: () => void; // Callback when user needs to login
  onBookmarkChange?: () => void; // Callback when bookmark state changes
  className?: string;
}

const BookmarkButton: React.FC<BookmarkButtonProps> = ({
  articleId,
  articleTitle = 'article',
  size = 'sm',
  variant = 'ghost',
  showText = false,
  onAuthRequired,
  onBookmarkChange,
  className = '',
}) => {
  const { authState, isBookmarked, addBookmark, removeBookmark } = useAuth();
  const { handleAuthRequired } = useBookmarkContext();
  const [isLoading, setIsLoading] = useState(false);

  const bookmarked = isBookmarked(articleId);

  const handleBookmarkClick = async (e: React.MouseEvent) => {
    e.stopPropagation();
    e.preventDefault();


    // Check if user is authenticated
    if (!authState.isAuthenticated) {
      const handleAuth = () => {
        // After login, attempt to bookmark the article
        handleBookmarkClick(e);
      };

      if (onAuthRequired) {
        onAuthRequired();
      } else {
        handleAuthRequired(handleAuth);
      }
      return;
    }

    setIsLoading(true);

    try {
      if (bookmarked) {
        // Remove bookmark
        const success = await removeBookmark(articleId);
        if (success) {
          toast.success('Bookmark Removed', {
            description: `Removed "${articleTitle}" from your bookmarks.`,
            duration: 2000,
          });
          // Trigger callback to refresh page/state
          onBookmarkChange?.();
        } else {
          toast.error('Failed to remove bookmark. Please try again.');
        }
      } else {
        // Add bookmark
        const success = await addBookmark(articleId);
        if (success) {
          toast.success('Bookmarked!', {
            description: `Added "${articleTitle}" to your bookmarks.`,
            duration: 2000,
          });
          // Trigger callback to refresh page/state
          onBookmarkChange?.();
        } else {
          toast.error('Failed to bookmark article. Please try again.');
        }
      }
    } catch (error) {
      console.error('Bookmark error:', error);
      toast.error('Something went wrong. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const getIconSize = () => {
    switch (size) {
      case 'lg': return 'h-5 w-5';
      case 'md': return 'h-4 w-4';
      case 'sm': return 'h-3 w-3';
      default: return 'h-4 w-4';
    }
  };

  const getButtonSize = () => {
    switch (size) {
      case 'lg': return 'default';
      case 'md': return 'sm';
      case 'sm': return 'sm';
      default: return 'sm';
    }
  };

  return (
    <Button
      variant={variant}
      size={getButtonSize() as any}
      onClick={handleBookmarkClick}
      disabled={isLoading}
      className={`
        transition-all duration-200 
        ${bookmarked 
          ? 'text-red-500 hover:text-red-600' 
          : 'text-gray-500 hover:text-red-500'
        }
        ${isLoading ? 'opacity-50 cursor-not-allowed' : ''}
        ${className}
      `}
      title={bookmarked ? 'Remove from bookmarks' : 'Add to bookmarks'}
    >
      <Heart 
        className={`
          ${getIconSize()} 
          transition-all duration-200
          ${bookmarked ? 'fill-current' : ''}
          ${isLoading ? 'animate-pulse' : ''}
        `} 
      />
      {showText && (
        <span className="ml-2">
          {isLoading 
            ? '...' 
            : bookmarked 
              ? 'Bookmarked' 
              : 'Bookmark'
          }
        </span>
      )}
    </Button>
  );
};

export default BookmarkButton;
