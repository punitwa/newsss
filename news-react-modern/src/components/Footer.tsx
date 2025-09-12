import React from 'react';
import { 
  Github, 
  Twitter, 
  Linkedin, 
  Mail, 
  Heart,
  Globe,
  Rss,
  TrendingUp,
  Shield,
  Zap,
  Users
} from 'lucide-react';

const Footer: React.FC = () => {
  const currentYear = new Date().getFullYear();

  const footerSections = [
    {
      title: 'Product',
      links: [
        { name: 'Top Stories', href: '/', icon: TrendingUp },
        { name: 'All News', href: '/all-news', icon: Globe },
        { name: 'RSS Feeds', href: '#', icon: Rss },
        { name: 'Search', href: '/search', icon: Globe },
      ]
    },
    {
      title: 'Features',
      links: [
        { name: 'Real-time Updates', href: '#', icon: Zap },
        { name: 'Smart Filtering', href: '#', icon: Shield },
        { name: 'Multi-source', href: '#', icon: Users },
        { name: 'Fast Search', href: '#', icon: TrendingUp },
      ]
    },
    {
      title: 'Sources',
      links: [
        { name: 'BBC News', href: '#', icon: Globe },
        { name: 'CNN', href: '#', icon: Globe },
        { name: 'TechCrunch', href: '#', icon: Globe },
        { name: 'The Hindu', href: '#', icon: Globe },
      ]
    }
  ];

  const socialLinks = [
    { name: 'GitHub', href: 'https://github.com', icon: Github },
    { name: 'Twitter', href: 'https://twitter.com', icon: Twitter },
    { name: 'LinkedIn', href: 'https://linkedin.com', icon: Linkedin },
    { name: 'Email', href: 'mailto:contact@worldbrief.com', icon: Mail },
  ];

  const stats = [
    { label: 'News Sources', value: '10+' },
    { label: 'Articles Daily', value: '500+' },
    { label: 'Categories', value: '8' },
    { label: 'Languages', value: '3' },
  ];

  return (
    <footer className="relative bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-white mt-20">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-50">
        <div className="w-full h-full bg-[radial-gradient(circle_at_1px_1px,rgba(255,255,255,0.03)_1px,transparent_0)] bg-[length:20px_20px]"></div>
      </div>
      
      <div className="relative">
        {/* Stats Section */}
        <div className="border-b border-slate-700/50">
          <div className="max-w-7xl mx-auto px-4 py-8">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
              {stats.map((stat, index) => (
                <div key={index} className="text-center">
                  <div className="text-2xl md:text-3xl font-bold text-blue-400 mb-1">
                    {stat.value}
                  </div>
                  <div className="text-sm text-slate-400">
                    {stat.label}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Main Footer Content */}
        <div className="max-w-7xl mx-auto px-4 py-12">
          <div className="grid grid-cols-1 lg:grid-cols-4 gap-8 lg:gap-12">
            
            {/* Brand Section */}
            <div className="lg:col-span-1">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-blue-600 rounded-lg">
                  <Globe className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h3 className="text-xl font-bold">WorldBrief</h3>
                  <p className="text-xs text-slate-400">Stay Informed</p>
                </div>
              </div>
              
              <p className="text-slate-300 mb-6 leading-relaxed">
                Your trusted source for real-time news aggregation. 
                We bring you the most important stories from around the world, 
                all in one place.
              </p>

              {/* API Status */}
              <div className="bg-slate-800/50 rounded-lg p-4 mb-6">
                <div className="flex items-center gap-2 mb-2">
                  <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                  <span className="text-sm font-medium text-slate-300">API Status</span>
                </div>
                <a 
                  href="http://localhost:8082/health" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
                >
                  http://localhost:8082
                </a>
              </div>

              {/* Social Links */}
              <div className="flex gap-3">
                {socialLinks.map((social, index) => (
                  <a
                    key={index}
                    href={social.href}
                    className="p-2 bg-slate-800/50 hover:bg-slate-700/50 rounded-lg transition-colors group"
                    title={social.name}
                  >
                    <social.icon className="h-4 w-4 text-slate-400 group-hover:text-white transition-colors" />
                  </a>
                ))}
              </div>
            </div>

            {/* Footer Links */}
            {footerSections.map((section, index) => (
              <div key={index}>
                <h4 className="font-semibold text-white mb-4">{section.title}</h4>
                <ul className="space-y-3">
                  {section.links.map((link, linkIndex) => (
                    <li key={linkIndex}>
                      <a
                        href={link.href}
                        className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors group"
                      >
                        <link.icon className="h-4 w-4 opacity-50 group-hover:opacity-100 transition-opacity" />
                        {link.name}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>

        {/* Bottom Section */}
        <div className="border-t border-slate-700/50">
          <div className="max-w-7xl mx-auto px-4 py-6">
            <div className="flex flex-col md:flex-row justify-between items-center gap-4">
              
              {/* Copyright */}
              <div className="flex items-center gap-2 text-slate-400">
                <span>&copy; {currentYear} WorldBrief. All rights reserved.</span>
                <div className="hidden md:block">•</div>
                <div className="flex items-center gap-1">
                  <span className="text-xs">Made with</span>
                  <Heart className="h-3 w-3 text-red-400" />
                  <span className="text-xs">using React & Go</span>
                </div>
              </div>

              {/* Tech Stack */}
            </div>
          </div>
        </div>

        {/* Decorative Elements */}
        <div className="absolute top-0 left-0 w-full h-px bg-gradient-to-r from-transparent via-blue-500 to-transparent opacity-50"></div>
      </div>
    </footer>
  );
};

export default Footer;
