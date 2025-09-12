import React from 'react';
import { Card } from '@/components/ui/card';
import { BarChart3, Filter, Search } from 'lucide-react';

interface StatsBarProps {
  totalArticles: number;
  currentFilter?: string;
  searchQuery?: string;
  loading?: boolean;
}

const StatsBar: React.FC<StatsBarProps> = ({
  totalArticles,
  currentFilter,
  searchQuery,
  loading = false,
}) => {
  const getStatsMessage = () => {
    if (loading) {
      return "Loading articles...";
    }
    
    if (searchQuery) {
      return `Found ${totalArticles} articles matching "${searchQuery}"`;
    } else if (currentFilter) {
      return `Showing ${totalArticles} ${currentFilter} articles`;
    } else {
      return `Showing ${totalArticles} articles`;
    }
  };

  const getIcon = () => {
    if (searchQuery) return <Search className="h-4 w-4" />;
    if (currentFilter) return <Filter className="h-4 w-4" />;
    return <BarChart3 className="h-4 w-4" />;
  };

  return (
    <Card className="w-full max-w-4xl mx-auto">
      <div className="flex items-center justify-center gap-2 p-3">
        <div className="text-primary">
          {getIcon()}
        </div>
        <span className="text-sm font-medium text-primary">
          {getStatsMessage()}
        </span>
      </div>
    </Card>
  );
};

export default StatsBar;
