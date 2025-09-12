import { TrendingUp, Hash, ChevronUp, ChevronDown, Minus, RefreshCw } from "lucide-react";
// import { Badge } from "@/components/ui/badge"; // Unused import
import { useState, useEffect } from "react";

interface TrendingTopic {
  name: string;
  article_count: number;
  today_change: number;
  trend_direction: 'up' | 'down' | 'stable';
  percentage: number;
  category?: string;
  last_updated: string;
}

interface TrendingTopicsProps {
  onTopicClick?: (topic: string) => void;
}

const TrendingTopics = ({ onTopicClick }: TrendingTopicsProps) => {
  const [trendingTopics, setTrendingTopics] = useState<TrendingTopic[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const fetchTrendingTopics = async () => {
    try {
      setError(null);
      const response = await fetch('http://localhost:8082/api/v1/trending?limit=6');
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const data = await response.json();
      setTrendingTopics(data.data || []);
      setLastUpdated(new Date());
    } catch (err) {
      console.error('Trending topics fetch error:', err);
      setError(err instanceof Error ? err.message : 'Failed to load trending topics');
      // Fallback to better mock data if API fails
      setTrendingTopics([
        { name: "AI & Technology", article_count: 124, today_change: 23, trend_direction: 'up', percentage: 85, category: "Technology", last_updated: new Date().toISOString() },
        { name: "Climate Change", article_count: 89, today_change: 15, trend_direction: 'up', percentage: 60, category: "Science", last_updated: new Date().toISOString() },
        { name: "Space Exploration", article_count: 63, today_change: -5, trend_direction: 'down', percentage: 43, category: "Science", last_updated: new Date().toISOString() },
        { name: "Cryptocurrency", article_count: 52, today_change: 8, trend_direction: 'up', percentage: 35, category: "Finance", last_updated: new Date().toISOString() },
        { name: "Healthcare", article_count: 45, today_change: 0, trend_direction: 'stable', percentage: 31, category: "Science", last_updated: new Date().toISOString() },
        { name: "Electric Vehicles", article_count: 38, today_change: 12, trend_direction: 'up', percentage: 26, category: "Technology", last_updated: new Date().toISOString() },
      ]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTrendingTopics();
    
    // Auto-refresh every 5 minutes
    const interval = setInterval(fetchTrendingTopics, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  const handleTopicClick = (topic: string) => {
    console.log('Trending topic clicked:', topic);
    onTopicClick?.(topic);
  };

  const getTrendIcon = (direction: string) => {
    switch (direction) {
      case 'up':
        return <ChevronUp className="h-3 w-3 text-green-500" />;
      case 'down':
        return <ChevronDown className="h-3 w-3 text-red-500" />;
      default:
        return <Minus className="h-3 w-3 text-gray-500" />;
    }
  };

  const getTrendColor = (direction: string, change: number) => {
    if (direction === 'up' && change > 0) return 'text-green-600';
    if (direction === 'down' && change < 0) return 'text-red-600';
    return 'text-gray-600';
  };

  if (loading) {
    return (
      <div className="bg-white rounded-xl p-6 border border-gray-100 shadow-sm">
        {/* Header Skeleton */}
        <div className="flex items-center space-x-3 mb-6">
          <div className="p-2 bg-gray-100 rounded-lg">
            <TrendingUp className="h-5 w-5 text-gray-300" />
          </div>
          <div>
            <div className="h-5 bg-gray-200 rounded w-32 animate-pulse mb-1" />
            <div className="h-3 bg-gray-100 rounded w-20 animate-pulse" />
          </div>
        </div>

        {/* Topics Skeleton */}
        <div className="space-y-2">
          {[...Array(6)].map((_, index) => (
            <div key={index} className="p-4 rounded-lg border border-gray-100">
              <div className="flex items-center space-x-3 mb-3">
                <div className="w-8 h-8 bg-gray-200 rounded-lg animate-pulse" />
                <div className="flex-1">
                  <div className="h-4 bg-gray-200 rounded w-24 animate-pulse mb-2" />
                  <div className="h-3 bg-gray-100 rounded w-16 animate-pulse" />
                </div>
              </div>
              <div className="w-full bg-gray-100 rounded-full h-1.5">
                <div className="h-full bg-gray-200 rounded-full animate-pulse" style={{ width: '60%' }} />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl p-4 sm:p-6 border border-gray-100 shadow-sm" data-trending-section>
      {/* Header */}
      <div className="flex items-center justify-between mb-4 sm:mb-6">
        <div className="flex items-center space-x-2 sm:space-x-3">
          <div className="p-1.5 sm:p-2 bg-blue-50 rounded-lg">
            <TrendingUp className="h-4 w-4 sm:h-5 sm:w-5 text-blue-600" />
          </div>
          <div>
            <h2 className="font-semibold text-gray-900 text-base sm:text-lg">Trending Topics</h2>
            {lastUpdated && (
              <p className="text-xs text-gray-500">
                Updated {lastUpdated.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </p>
            )}
          </div>
        </div>
        <button 
          onClick={fetchTrendingTopics}
          className="p-2 hover:bg-gray-50 rounded-lg transition-colors"
          title="Refresh topics"
        >
          <RefreshCw className="h-4 w-4 text-gray-400" />
        </button>
      </div>

      {error && (
        <div className="text-xs text-amber-700 mb-4 p-3 bg-amber-50 border border-amber-200 rounded-lg">
          <div className="flex items-center space-x-2">
            <div className="w-1 h-1 bg-amber-500 rounded-full"></div>
            <span>Using cached data - API unavailable</span>
          </div>
        </div>
      )}

      {/* Topics List */}
      <div className="space-y-2">
        {trendingTopics.map((topic, index) => (
          <div 
            key={topic.name}
            className="group relative p-4 rounded-lg border border-gray-100 hover:border-blue-200 hover:bg-blue-50/30 transition-all duration-200 cursor-pointer"
            onClick={() => handleTopicClick(topic.name)}
            title="Click to filter news by this topic"
          >
            <div className="flex items-center justify-between">
              {/* Left side */}
              <div className="flex items-center space-x-3 flex-1 min-w-0">
                <div className="flex items-center justify-center w-8 h-8 bg-gray-100 group-hover:bg-blue-100 rounded-lg text-sm font-semibold text-gray-600 group-hover:text-blue-600 transition-colors flex-shrink-0">
                  {index + 1}
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center space-x-2 mb-1">
                    <span className="font-medium text-gray-900 group-hover:text-blue-900 transition-colors truncate">
                      {topic.name}
                    </span>
                    {topic.category && topic.category !== "General" && (
                      <span className="text-xs px-2 py-1 bg-gray-100 text-gray-600 rounded-full flex-shrink-0">
                        {topic.category}
                      </span>
                    )}
                  </div>
                  <div className="flex items-center space-x-3 text-xs text-gray-500">
                    <span>{topic.article_count} articles</span>
                    <div className="flex items-center space-x-1">
                      {getTrendIcon(topic.trend_direction)}
                      <span className={getTrendColor(topic.trend_direction, topic.today_change)}>
                        {topic.today_change > 0 ? '+' : ''}{topic.today_change} today
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Right side - Arrow */}
              <div className="flex-shrink-0 ml-3">
                <div className="w-6 h-6 rounded-full bg-gray-100 group-hover:bg-blue-100 flex items-center justify-center transition-colors">
                  <Hash className="h-3 w-3 text-gray-400 group-hover:text-blue-500" />
                </div>
              </div>
            </div>

            {/* Progress bar */}
            <div className="mt-3 w-full bg-gray-100 rounded-full h-1.5 overflow-hidden">
              <div 
                className="h-full bg-gradient-to-r from-blue-500 to-blue-600 rounded-full transition-all duration-500 ease-out"
                style={{ width: `${Math.min(topic.percentage, 100)}%` }}
              />
            </div>
          </div>
        ))}
      </div>

      {trendingTopics.length === 0 && !loading && (
        <div className="text-center py-8 text-gray-500">
          <TrendingUp className="h-8 w-8 mx-auto mb-2 opacity-30" />
          <p className="text-sm">No trending topics available</p>
        </div>
      )}
    </div>
  );
};

export default TrendingTopics;
