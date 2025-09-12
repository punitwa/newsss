import React from 'react';
import { Button } from '@/components/ui/button';
import { getCategoryIcon } from '@/lib/utils';

interface CategoryFilterProps {
  categories: string[];
  selectedCategory: string;
  onCategorySelect: (category: string) => void;
  loading?: boolean;
}

const CategoryFilter: React.FC<CategoryFilterProps> = ({
  categories,
  selectedCategory,
  onCategorySelect,
  loading = false,
}) => {
  const formatCategoryName = (category: string) => {
    return category.charAt(0).toUpperCase() + category.slice(1);
  };

  return (
    <div className="w-full max-w-6xl mx-auto bg-card rounded-lg p-4 shadow-sm border">
      <div className="flex flex-wrap gap-2 justify-center">
        <Button
          variant={selectedCategory === '' ? 'default' : 'outline'}
          size="sm"
          onClick={() => onCategorySelect('')}
          disabled={loading}
          className="transition-all duration-200 hover:scale-105"
        >
          📰 All
        </Button>
        
        {categories.map((category) => (
          <Button
            key={category}
            variant={selectedCategory === category ? 'default' : 'outline'}
            size="sm"
            onClick={() => onCategorySelect(category)}
            disabled={loading}
            className="transition-all duration-200 hover:scale-105"
          >
            {getCategoryIcon(category)} {formatCategoryName(category)}
          </Button>
        ))}
      </div>
    </div>
  );
};

export default CategoryFilter;
