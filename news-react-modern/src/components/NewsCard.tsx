import React from 'react';
import { Clock, ExternalLink, Share2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { News } from '@/types/news';
import BookmarkButton from '@/components/BookmarkButton';
// import { useAutoEngagementTracking } from '@/hooks/useEngagementTracking'; // Temporarily disabled

interface NewsCardProps {
  news: News;
  onClick?: () => void;
  onBookmarkChange?: () => void;
}

const NewsCard: React.FC<NewsCardProps> = ({ news, onClick, onBookmarkChange }) => {
  // Engagement tracking (temporarily disabled)
  // const { elementRef, handleClick, handleShare } = useAutoEngagementTracking(
  //   news.id,
  //   {
  //     trackViews: false, // Temporarily disabled until backend endpoints are ready
  //     trackClicks: false,
  //     trackShares: false,
  //     trackReadTime: false,
  //   }
  // );
  
  // Temporary placeholder refs and handlers
  const elementRef = React.useRef<HTMLDivElement>(null);
  const timeAgo = (date: string) => {
    const now = new Date().getTime();
    const published = new Date(date).getTime();
    const diff = Math.floor((now - published) / (1000 * 60)); // minutes
    
    if (diff < 60) return `${diff}m ago`;
    if (diff < 1440) return `${Math.floor(diff / 60)}h ago`;
    return `${Math.floor(diff / 1440)}d ago`;
  };

  // Bookmark functionality now handled by BookmarkButton component

  const handleReadMore = (e: React.MouseEvent) => {
    e.stopPropagation();
    // handleClick(e); // Track engagement - temporarily disabled
    console.log('Read more clicked for:', news.title);
    window.open(news.url, '_blank');
  };

  const handleShareClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    
    // Simple share functionality
    if (navigator.share) {
      navigator.share({
        title: news.title,
        text: news.summary || news.title,
        url: news.url,
      });
    } else {
      // Fallback: copy to clipboard
      navigator.clipboard.writeText(news.url);
      console.log('URL copied to clipboard');
    }
  };

  const handleCardClick = () => {
    if (onClick) {
      onClick();
    } else {
      window.open(news.url, '_blank', 'noopener,noreferrer');
    }
  };

  const truncateToWords = (text: string, wordLimit: number) => {
    if (!text) return '';
    const words = text.split(' ');
    if (words.length <= wordLimit) return text;
    return words.slice(0, wordLimit).join(' ') + '...';
  };

  return (
    <article 
      ref={elementRef}
      className="news-card p-0 overflow-hidden group cursor-pointer h-full flex flex-col" 
      onClick={handleCardClick}
    >
      {/* Image Section */}
      <div className="relative overflow-hidden">
          <img 
            src={news.image_url || "https://groundwater.org/wp-content/uploads/2022/07/news-placeholder.png"} 
            alt={news.title}
            className="w-full h-48 object-cover transition-transform duration-300 group-hover:scale-105"
            onError={(e) => {
              const target = e.target as HTMLImageElement;
              target.src = "https://groundwater.org/wp-content/uploads/2022/07/news-placeholder.png";
            }}
          />
          <Badge 
            variant="secondary" 
            className="absolute top-3 left-3 bg-background/90 backdrop-blur-sm text-xs"
          >
            {news.category}
          </Badge>
        </div>

      {/* Content Section */}
      <div className="p-6 space-y-4 flex-1 flex flex-col">

        {/* Header */}
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <h3 className="font-heading font-semibold text-lg leading-tight line-clamp-2 group-hover:text-primary transition-colors">
              {news.title}
            </h3>
          </div>
        </div>

        {/* Description */}
        <div className="flex-1">
          <p className="text-muted-foreground text-sm leading-relaxed line-clamp-4">
            {truncateToWords(news.summary || news.content || 'No description available', 80)}
          </p>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between pt-2 border-t border-border/50 mt-auto">
          <div className="flex items-center gap-2 text-xs text-muted-foreground flex-wrap">
            <span className="font-medium text-foreground whitespace-nowrap">{news.source}</span>
            <div className="flex items-center gap-1">
              <Clock className="h-3 w-3 flex-shrink-0" />
              <span className="whitespace-nowrap">{timeAgo(news.published_at)}</span>
            </div>
          </div>
          
          <div className="flex gap-2">
            <Button 
              variant="ghost" 
              size="sm" 
              className="text-xs text-primary hover:text-primary-hover transition-colors flex-1"
              onClick={handleReadMore}
            >
              <ExternalLink className="h-3 w-3 mr-1" />
              Read More
            </Button>
            
            <BookmarkButton 
              articleId={news.id}
              articleTitle={news.title}
              size="sm"
              variant="ghost"
              className="text-xs px-2"
              onBookmarkChange={onBookmarkChange}
            />
            
            <Button 
              variant="ghost" 
              size="sm" 
              className="text-xs text-muted-foreground hover:text-primary transition-colors px-2"
              onClick={handleShareClick}
              title="Share article"
            >
              <Share2 className="h-3 w-3" />
            </Button>
          </div>
        </div>
      </div>

      <style>{`
        .line-clamp-2 {
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
        }
        
        .line-clamp-3 {
          display: -webkit-box;
          -webkit-line-clamp: 3;
          -webkit-box-orient: vertical;
          overflow: hidden;
        }
        
        .line-clamp-4 {
          display: -webkit-box;
          -webkit-line-clamp: 4;
          -webkit-box-orient: vertical;
          overflow: hidden;
        }
      `}</style>
    </article>
  );
};

export default NewsCard;
