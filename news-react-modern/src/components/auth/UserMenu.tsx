import React, { useState, useRef, useEffect } from 'react';
import { 
  User, 
  Settings, 
  LogOut, 
  ChevronDown,
  Bell,
  Heart,
  TrendingUp,
  LogIn,
  UserPlus,
  Sliders
} from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext';

interface UserMenuProps {
  onProfileClick?: () => void;
  onSettingsClick?: () => void;
  onBookmarksClick?: () => void;
  onLoginClick?: () => void;
  onRegisterClick?: () => void;
}

const UserMenu: React.FC<UserMenuProps> = ({
  onProfileClick,
  onSettingsClick,
  onBookmarksClick,
  onLoginClick,
  onRegisterClick
}) => {
  const { authState, logout } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const { user } = authState;

  // Get user initials for avatar
  const getInitials = (firstName: string, lastName: string) => {
    return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase();
  };

  // Handle menu item click
  const handleMenuClick = (action: () => void) => {
    action();
    setIsOpen(false);
  };

  return (
    <div className="relative" ref={dropdownRef}>
      {/* User Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center space-x-3 px-3 py-2 rounded-lg hover:bg-gray-100 transition-colors"
      >
        {authState.isAuthenticated && user ? (
          <>
            {/* Avatar */}
            <div className="relative">
              {user.avatar ? (
                <img
                  src={user.avatar}
                  alt={`${user.first_name} ${user.last_name}`}
                  className="h-8 w-8 rounded-full object-cover"
                />
              ) : (
                <div className="h-8 w-8 rounded-full bg-blue-600 flex items-center justify-center">
                  <span className="text-white text-sm font-medium">
                    {getInitials(user.first_name, user.last_name)}
                  </span>
                </div>
              )}
              
              {/* Online indicator */}
              <div className="absolute -bottom-0.5 -right-0.5 h-3 w-3 bg-green-400 rounded-full border-2 border-white"></div>
            </div>

            {/* User Info */}
            <div className="hidden md:block text-left">
              <p className="text-sm font-medium text-gray-900">
                {user.first_name} {user.last_name}
              </p>
              <p className="text-xs text-gray-500">@{user.username}</p>
            </div>
          </>
        ) : (
          <>
            {/* Generic User Icon for unauthenticated users */}
            <div className="h-8 w-8 rounded-full bg-gray-400 flex items-center justify-center">
              <User className="h-4 w-4 text-white" />
            </div>
            <div className="hidden md:block text-left">
              <p className="text-sm font-medium text-gray-900">Account</p>
            </div>
          </>
        )}

        {/* Dropdown Arrow */}
        <ChevronDown 
          className={`h-4 w-4 text-gray-400 transition-transform ${
            isOpen ? 'transform rotate-180' : ''
          }`} 
        />
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 py-2 z-50">
          {authState.isAuthenticated && user ? (
            <>
              {/* User Info Header */}
              <div className="px-4 py-3 border-b border-gray-100">
                <div className="flex items-center space-x-3">
                  {user.avatar ? (
                    <img
                      src={user.avatar}
                      alt={`${user.first_name} ${user.last_name}`}
                      className="h-10 w-10 rounded-full object-cover"
                    />
                  ) : (
                    <div className="h-10 w-10 rounded-full bg-blue-600 flex items-center justify-center">
                      <span className="text-white font-medium">
                        {getInitials(user.first_name, user.last_name)}
                      </span>
                    </div>
                  )}
                  <div>
                    <p className="font-medium text-gray-900">
                      {user.first_name} {user.last_name}
                    </p>
                    <p className="text-sm text-gray-500">{user.email}</p>
                  </div>
                </div>
              </div>

              {/* Authenticated Menu Items */}
              <div className="py-2">
                {/* Profile */}
                <button
                  onClick={() => handleMenuClick(() => onProfileClick?.())}
                  className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <User className="h-4 w-4 mr-3 text-gray-400" />
                  View Profile
                </button>

                {/* Bookmarks */}
                <button
                  onClick={() => handleMenuClick(() => onBookmarksClick?.())}
                  className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <Heart className="h-4 w-4 mr-3 text-red-500" />
                  My Bookmarks
                </button>

                {/* Reading Activity */}
                <button
                  onClick={() => handleMenuClick(() => console.log('Reading activity'))}
                  className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <TrendingUp className="h-4 w-4 mr-3 text-gray-400" />
                  Reading Activity
                </button>

                {/* Notifications */}
                <button
                  onClick={() => handleMenuClick(() => console.log('Notifications'))}
                  className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <Bell className="h-4 w-4 mr-3 text-gray-400" />
                  Notifications
                </button>

                <div className="border-t border-gray-100 my-2"></div>

                {/* Settings */}
                <button
                  onClick={() => handleMenuClick(() => onSettingsClick?.())}
                  className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <Settings className="h-4 w-4 mr-3 text-gray-400" />
                  Settings
                </button>

                {/* Preferences */}
                <button
                  onClick={() => handleMenuClick(() => console.log('Preferences'))}
                  className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <Sliders className="h-4 w-4 mr-3 text-gray-400" />
                  Preferences
                </button>

                <div className="border-t border-gray-100 my-2"></div>

                {/* Logout */}
                <button
                  onClick={() => handleMenuClick(logout)}
                  className="flex items-center w-full px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors"
                >
                  <LogOut className="h-4 w-4 mr-3 text-red-500" />
                  Sign Out
                </button>
              </div>

              {/* Admin Badge */}
              {user.is_admin && (
                <div className="px-4 py-2 border-t border-gray-100">
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                    Administrator
                  </span>
                </div>
              )}
            </>
          ) : (
            <>
              {/* Unauthenticated Menu Items */}
              <div className="px-4 py-3 border-b border-gray-100">
                <p className="text-sm text-gray-600">Sign in to access your account</p>
              </div>
              
              <div className="py-2">
                {/* Login */}
                <button
                  onClick={() => handleMenuClick(() => onLoginClick?.())}
                  className="flex items-center w-full px-4 py-2 text-sm text-blue-600 hover:bg-blue-50 transition-colors"
                >
                  <LogIn className="h-4 w-4 mr-3 text-blue-500" />
                  Sign In
                </button>

                {/* Register */}
                <button
                  onClick={() => handleMenuClick(() => onRegisterClick?.())}
                  className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <UserPlus className="h-4 w-4 mr-3 text-gray-400" />
                  Create Account
                </button>
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default UserMenu;
