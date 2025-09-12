import React from 'react';
import { Newspaper, Github, Heart } from 'lucide-react';

const Header: React.FC = () => {
  return (
    <header className="news-gradient text-white py-12 mb-8">
      <div className="container mx-auto px-4 text-center">
        <div className="flex items-center justify-center gap-3 mb-4">
          <Newspaper className="h-8 w-8" />
          <h1 className="text-4xl md:text-5xl font-bold">
            News Aggregator
          </h1>
        </div>
        
        <p className="text-lg md:text-xl opacity-90 mb-6 max-w-2xl mx-auto">
          Stay updated with the latest news from around the world. 
          Discover, search, and explore stories that matter to you.
        </p>
        
        <div className="flex items-center justify-center gap-4 text-sm opacity-80">
          <div className="flex items-center gap-1">
            <Heart className="h-4 w-4" />
            <span>Built with React & Go</span>
          </div>
          <div className="hidden md:block">•</div>
          <div className="flex items-center gap-1">
            <Github className="h-4 w-4" />
            <span>Open Source</span>
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
