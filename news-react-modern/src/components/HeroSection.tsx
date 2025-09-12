import { ArrowRight, Globe, TrendingUp, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useState, useEffect } from "react";
// import heroImage from "@/assets/news-hero.jpg"; // Placeholder for hero image

const HeroSection = () => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    setIsVisible(true);
  }, []);

  const handleExploreNews = () => {
    console.log('Explore News clicked');
    // Scroll to news section
    document.querySelector('main > div')?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleLearnMore = () => {
    console.log('Learn More clicked');
  };


  return (
    <section className="relative overflow-hidden py-16 lg:py-24">
      {/* Animated Background */}
      <div className="absolute inset-0 bg-gradient-to-br from-blue-900 via-blue-800 to-indigo-900" />
      <div 
        className="absolute inset-0 opacity-50"
        style={{
          backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.03'%3E%3Ccircle cx='30' cy='30' r='1'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
          backgroundRepeat: 'repeat'
        }}
      />
      
      {/* Floating Elements */}
      <div className="absolute top-10 left-10 w-24 h-24 bg-blue-400/10 rounded-full blur-xl animate-pulse" />
      <div className="absolute bottom-10 right-10 w-32 h-32 bg-indigo-400/10 rounded-full blur-xl animate-pulse delay-1000" />
      <div className="absolute top-1/2 left-1/4 w-16 h-16 bg-cyan-400/10 rounded-full blur-xl animate-pulse delay-2000" />
      
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6">
        <div className="grid lg:grid-cols-2 gap-8 lg:gap-12 items-center">
          {/* Content */}
          <div className={`space-y-6 lg:space-y-8 transition-all duration-1000 ${isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'}`}>
            {/* Badge */}
            <div className="inline-flex items-center space-x-2 bg-white/10 backdrop-blur-sm border border-white/20 rounded-full px-4 py-2">
              <Sparkles className="h-4 w-4 text-yellow-400 animate-pulse" />
              <span className="text-sm font-medium text-white uppercase tracking-wider">
                Real-time News Intelligence
              </span>
            </div>
            
            {/* Main Headline */}
            <div className="space-y-3 sm:space-y-4">
              <h1 className="text-3xl sm:text-4xl lg:text-6xl font-black text-white leading-tight">
                Stay
                <span className="block bg-gradient-to-r from-orange-400 via-red-500 to-pink-500 bg-clip-text text-transparent animate-pulse">
                  Informed
                </span>
                <span className="block text-2xl sm:text-3xl lg:text-4xl font-medium text-blue-100">
                  with Global News
                </span>
              </h1>
              
              <p className="text-base sm:text-lg text-blue-100 leading-relaxed max-w-lg">
                Discover breaking news, trending stories, and in-depth analysis from 
                <span className="text-orange-300 font-semibold"> trusted sources worldwide</span>. 
                All powered by AI, all in one place.
              </p>
            </div>

            {/* Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-4">
              <Button 
                size="lg" 
                className="group bg-gradient-to-r from-orange-500 to-red-600 hover:from-orange-600 hover:to-red-700 text-white font-bold shadow-2xl hover:shadow-orange-500/25 transition-all duration-300 transform hover:scale-105 px-8 py-4 text-lg"
                onClick={handleExploreNews}
              >
                <Globe className="h-5 w-5 mr-3 group-hover:animate-spin" />
                Explore News
                <ArrowRight className="h-5 w-5 ml-3 group-hover:translate-x-1 transition-transform" />
              </Button>
              
              <Button 
                variant="outline" 
                size="lg"
                className="border-2 border-white/60 text-white bg-white/10 hover:bg-white/20 backdrop-blur-sm hover:border-white/80 transition-all duration-300 px-8 py-4 text-lg font-semibold"
                onClick={handleLearnMore}
              >
                <TrendingUp className="h-5 w-5 mr-2" />
                Learn More
              </Button>
            </div>

          </div>

          {/* Enhanced Hero Visual */}
          <div className={`relative transition-all duration-1000 delay-300 ${isVisible ? 'opacity-100 translate-x-0' : 'opacity-0 translate-x-8'}`}>
            {/* Glowing Background */}
            <div className="absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-600/20 rounded-3xl blur-3xl" />
            
            {/* Main Image Container */}
            <div className="relative rounded-3xl overflow-hidden shadow-2xl shadow-blue-900/50 border border-white/10 backdrop-blur-sm">
              <img 
                src="https://images.unsplash.com/photo-1504711434969-e33886168f5c?w=800&h=600&fit=crop&crop=top" 
                alt="Advanced News Dashboard"
                className="w-full h-64 lg:h-80 object-cover"
              />
              
              {/* Overlay Effects */}
              <div className="absolute inset-0 bg-gradient-to-t from-blue-900/60 via-transparent to-transparent" />
              
              {/* Floating UI Elements */}
              <div className="absolute top-6 left-6 bg-white/90 backdrop-blur-sm rounded-lg px-4 py-2 shadow-lg animate-pulse">
                <div className="flex items-center space-x-2">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                  <span className="text-sm font-semibold text-gray-800">Live Updates</span>
                </div>
              </div>
              
              <div className="absolute bottom-6 right-6 bg-gradient-to-r from-orange-500 to-red-600 text-white rounded-lg px-4 py-2 shadow-lg">
                <div className="flex items-center space-x-2">
                  <TrendingUp className="h-4 w-4" />
                  <span className="text-sm font-bold">Trending Now</span>
                </div>
              </div>
              
              {/* Animated Particles */}
              <div className="absolute top-1/4 left-1/4 w-2 h-2 bg-yellow-400 rounded-full animate-ping" />
              <div className="absolute top-3/4 left-3/4 w-1 h-1 bg-blue-400 rounded-full animate-ping delay-1000" />
              <div className="absolute top-1/2 right-1/4 w-1.5 h-1.5 bg-green-400 rounded-full animate-ping delay-2000" />
            </div>
            
            {/* Decorative Elements */}
            <div className="absolute -top-4 -right-4 w-24 h-24 bg-gradient-to-r from-yellow-400 to-orange-500 rounded-full opacity-20 blur-xl animate-pulse" />
            <div className="absolute -bottom-6 -left-6 w-32 h-32 bg-gradient-to-r from-purple-400 to-blue-500 rounded-full opacity-20 blur-xl animate-pulse delay-1000" />
          </div>
        </div>
      </div>
    </section>
  );
};

export default HeroSection;
