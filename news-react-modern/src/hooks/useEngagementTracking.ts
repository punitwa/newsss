import { useCallback, useEffect, useRef } from 'react';
import { newsApi } from '@/services/api';

interface EngagementTrackingOptions {
  trackViews?: boolean;
  trackClicks?: boolean;
  trackShares?: boolean;
  trackReadTime?: boolean;
  readTimeThreshold?: number; // minimum seconds before tracking read time
}

interface UseEngagementTrackingReturn {
  trackView: (articleId: string) => void;
  trackClick: (articleId: string) => void;
  trackShare: (articleId: string) => void;
  startReadTimeTracking: (articleId: string) => void;
  stopReadTimeTracking: () => void;
}

export const useEngagementTracking = (
  options: EngagementTrackingOptions = {}
): UseEngagementTrackingReturn => {
  const {
    trackViews = true,
    trackClicks = true,
    trackShares = true,
    trackReadTime = true,
    readTimeThreshold = 10, // 10 seconds minimum
  } = options;

  const readStartTime = useRef<number | null>(null);
  const currentArticleId = useRef<string | null>(null);
  const hasTrackedView = useRef<Set<string>>(new Set());
  const readTimeInterval = useRef<NodeJS.Timeout | null>(null);

  // Track article view
  const trackView = useCallback(async (articleId: string) => {
    if (!trackViews || hasTrackedView.current.has(articleId)) {
      return;
    }

    try {
      await newsApi.trackEngagement(articleId, 'view');
      hasTrackedView.current.add(articleId);
    } catch (error) {
      console.warn('Failed to track view:', error);
    }
  }, [trackViews]);

  // Track article click
  const trackClick = useCallback(async (articleId: string) => {
    if (!trackClicks) return;

    try {
      await newsApi.trackEngagement(articleId, 'click');
    } catch (error) {
      console.warn('Failed to track click:', error);
    }
  }, [trackClicks]);

  // Track article share
  const trackShare = useCallback(async (articleId: string) => {
    if (!trackShares) return;

    try {
      await newsApi.trackEngagement(articleId, 'share');
    } catch (error) {
      console.warn('Failed to track share:', error);
    }
  }, [trackShares]);

  // Start tracking read time
  const startReadTimeTracking = useCallback((articleId: string) => {
    if (!trackReadTime) return;

    // Stop any existing tracking
    stopReadTimeTracking();

    readStartTime.current = Date.now();
    currentArticleId.current = articleId;

    // Track read time every 30 seconds while user is reading
    readTimeInterval.current = setInterval(() => {
      if (readStartTime.current && currentArticleId.current) {
        const readTime = Math.floor((Date.now() - readStartTime.current) / 1000);
        
        if (readTime >= readTimeThreshold) {
          trackReadTime(currentArticleId.current, readTime);
          readStartTime.current = Date.now(); // Reset timer
        }
      }
    }, 30000); // Check every 30 seconds

  }, [trackReadTime, readTimeThreshold]);

  // Stop tracking read time
  const stopReadTimeTracking = useCallback(() => {
    if (readTimeInterval.current) {
      clearInterval(readTimeInterval.current);
      readTimeInterval.current = null;
    }

    // Track final read time if threshold is met
    if (readStartTime.current && currentArticleId.current && trackReadTime) {
      const totalReadTime = Math.floor((Date.now() - readStartTime.current) / 1000);
      
      if (totalReadTime >= readTimeThreshold) {
        trackReadTime(currentArticleId.current, totalReadTime);
      }
    }

    readStartTime.current = null;
    currentArticleId.current = null;
  }, [trackReadTime, readTimeThreshold]);

  // Internal method to track read time
  const trackReadTime = async (articleId: string, readTime: number) => {
    try {
      await newsApi.trackReadTime(articleId, readTime);
    } catch (error) {
      console.warn('Failed to track read time:', error);
    }
  };

  // Handle page visibility changes to pause/resume read time tracking
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden) {
        // Page is hidden, pause read time tracking
        if (readTimeInterval.current) {
          clearInterval(readTimeInterval.current);
          readTimeInterval.current = null;
        }
      } else {
        // Page is visible again, resume tracking if we were tracking
        if (currentArticleId.current && trackReadTime && !readTimeInterval.current) {
          readStartTime.current = Date.now(); // Reset start time
          
          readTimeInterval.current = setInterval(() => {
            if (readStartTime.current && currentArticleId.current) {
              const readTime = Math.floor((Date.now() - readStartTime.current) / 1000);
              
              if (readTime >= readTimeThreshold) {
                trackReadTime(currentArticleId.current, readTime);
                readStartTime.current = Date.now();
              }
            }
          }, 30000);
        }
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [trackReadTime, readTimeThreshold]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      stopReadTimeTracking();
    };
  }, [stopReadTimeTracking]);

  return {
    trackView,
    trackClick,
    trackShare,
    startReadTimeTracking,
    stopReadTimeTracking,
  };
};

// Hook for automatic engagement tracking on article components
export const useAutoEngagementTracking = (
  articleId: string,
  options: EngagementTrackingOptions = {}
) => {
  const {
    trackView,
    trackClick,
    trackShare,
    startReadTimeTracking,
    stopReadTimeTracking,
  } = useEngagementTracking(options);

  const elementRef = useRef<HTMLElement>(null);

  // Auto-track view when component mounts and becomes visible
  useEffect(() => {
    if (!articleId) return;

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            // Article is visible, track view and start read time tracking
            trackView(articleId);
            startReadTimeTracking(articleId);
          } else {
            // Article is no longer visible, stop read time tracking
            stopReadTimeTracking();
          }
        });
      },
      {
        threshold: 0.5, // Track when 50% of the article is visible
        rootMargin: '0px 0px -100px 0px', // Account for header/footer
      }
    );

    if (elementRef.current) {
      observer.observe(elementRef.current);
    }

    return () => {
      observer.disconnect();
      stopReadTimeTracking();
    };
  }, [articleId, trackView, startReadTimeTracking, stopReadTimeTracking]);

  // Handle click tracking
  const handleClick = useCallback((event: React.MouseEvent) => {
    trackClick(articleId);
    
    // If it's a link click, track it as a click-through
    const target = event.target as HTMLElement;
    if (target.tagName === 'A' || target.closest('a')) {
      // This is a click to read the full article
      trackClick(articleId);
    }
  }, [articleId, trackClick]);

  // Handle share tracking
  const handleShare = useCallback((platform?: string) => {
    trackShare(articleId);
    
    // Could be extended to track which platform was used for sharing
    console.log(`Article ${articleId} shared${platform ? ` on ${platform}` : ''}`);
  }, [articleId, trackShare]);

  return {
    elementRef,
    handleClick,
    handleShare,
    trackView: () => trackView(articleId),
    trackClick: () => trackClick(articleId),
    trackShare: () => trackShare(articleId),
  };
};
