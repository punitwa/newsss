import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { authApi } from '@/services/authApi';

// Types
export interface User {
  id: string;
  email: string;
  username: string;
  first_name: string;
  last_name: string;
  avatar: string;
  is_active: boolean;
  is_admin: boolean;
  created_at: string;
  preferences: {
    categories: string[];
    sources: string[];
    notification_enabled: boolean;
    email_digest: boolean;
    digest_frequency: string;
    theme: string;
    language: string;
  };
}

export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  loading: boolean;
  error: string | null;
  bookmarks: string[]; // Array of article IDs
  bookmarksLoading: boolean;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  email: string;
  username: string;
  password: string;
  first_name: string;
  last_name: string;
}

interface AuthContextType {
  // State
  authState: AuthState;
  
  // Actions
  login: (credentials: LoginCredentials) => Promise<boolean>;
  register: (data: RegisterData) => Promise<boolean>;
  logout: () => void;
  refreshToken: () => Promise<boolean>;
  clearError: () => void;
  
  // Bookmark actions
  addBookmark: (articleId: string) => Promise<boolean>;
  removeBookmark: (articleId: string) => Promise<boolean>;
  isBookmarked: (articleId: string) => boolean;
  loadBookmarks: () => Promise<void>;
  
  // Utilities
  isLoading: boolean;
  hasError: boolean;
}

// Create context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Auth provider component
interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: false,
    user: null,
    token: null,
    loading: true,
    error: null,
    bookmarks: [],
    bookmarksLoading: false,
  });

  // Helper function to ensure error is always a string
  const normalizeError = (error: any): string => {
    if (typeof error === 'string') {
      return error;
    }
    if (error && typeof error === 'object') {
      if (error.message) {
        return error.message;
      }
      if (error.error) {
        return error.error;
      }
      return JSON.stringify(error);
    }
    return 'An unexpected error occurred';
  };

  // Debug: Log state changes (remove in production)
  // useEffect(() => {
  //   console.log('AuthContext: State changed:', authState);
  // }, [authState]);

  // Initialize auth state from localStorage
  useEffect(() => {
    console.log('AuthContext: useEffect initialization starting...');
    
    // Synchronous initialization - no async needed
    try {
      const token = localStorage.getItem('auth_token');
      const userData = localStorage.getItem('user_data');
      
      console.log('AuthContext: Retrieved from localStorage:', { 
        token: token ? token.substring(0, 20) + '...' : null, 
        userData: userData ? 'exists' : 'null' 
      });
      
      if (token && userData && token !== 'undefined' && userData !== 'undefined') {
        console.log('AuthContext: Both token and userData exist and are valid');
        try {
          const user = JSON.parse(userData);
          console.log('AuthContext: Parsed user data:', user);
        
          console.log('AuthContext: Setting authenticated state...');
          
          const newState = {
            isAuthenticated: true,
            user,
            token,
            loading: false,
            error: null,
            bookmarks: [],
            bookmarksLoading: false,
          };
          
          console.log('AuthContext: New state to set:', newState);
          setAuthState(newState);
          
          // Set token for API requests
          authApi.setAuthToken(token);
          console.log('AuthContext: State set successfully');
          
          // Load user bookmarks immediately after setting authentication state
          console.log('AuthContext: Loading bookmarks after initialization...');
          // Use a direct call since we know the user is authenticated
          loadBookmarksDirectly();
        } catch (parseError) {
          console.error('AuthContext: Failed to parse user data:', parseError);
          // Clear invalid data
          localStorage.removeItem('auth_token');
          localStorage.removeItem('user_data');
          setAuthState(prev => ({ ...prev, loading: false }));
        }
      } else {
        console.log('AuthContext: No valid token or userData found, clearing localStorage');
        // Clear any invalid data
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user_data');
        setAuthState(prev => ({ ...prev, loading: false }));
      }
    } catch (error) {
      console.error('Auth initialization error:', error);
      setAuthState(prev => ({ 
        ...prev, 
        loading: false,
        error: 'Failed to initialize authentication'
      }));
    }
  }, []);

  // Login function
  const login = async (credentials: LoginCredentials): Promise<boolean> => {
    setAuthState(prev => ({ ...prev, loading: true, error: null }));
    
    try {
      const response = await authApi.login(credentials);
      
      if (response.success && response.data) {
        // Handle the nested data structure from the API
        // Backend wraps response in: { data: { token, user, expires_in, token_type }, request_id, timestamp }
        const responseData = response.data.data || response.data;
        const { token, user } = responseData;
        
        if (!token || typeof token !== 'string') {
          const errorMsg = normalizeError('Invalid authentication token received');
          console.error('Login error: Invalid token', { token, response });
          setAuthState(prev => ({ ...prev, loading: false, error: errorMsg }));
          return false;
        }
        
        if (!user || typeof user !== 'object') {
          const errorMsg = normalizeError('Invalid user information received');
          console.error('Login error: Invalid user data', { user, response });
          setAuthState(prev => ({ ...prev, loading: false, error: errorMsg }));
          return false;
        }
        
        // Store in localStorage with error handling
        try {
          localStorage.setItem('auth_token', token);
          localStorage.setItem('user_data', JSON.stringify(user));
        } catch (storageError) {
          console.error('Failed to save authentication data:', storageError);
          setAuthState(prev => ({ ...prev, loading: false, error: normalizeError('Failed to save login session') }));
          return false;
        }
        
        // Set token for API requests
        authApi.setAuthToken(token);
        
        // Update state
        setAuthState({
          isAuthenticated: true,
          user,
          token,
          loading: false,
          error: null,
          bookmarks: [],
          bookmarksLoading: false,
        });
        
        // Load bookmarks after successful login
        loadBookmarksDirectly();
        
        console.log('Login successful');
        return true;
      } else {
        // Handle API error response
        const errorMsg = normalizeError(response.error || 'Invalid email or password');
        console.error('Login failed:', { error: response.error, response });
        
        setAuthState(prev => ({
          ...prev,
          loading: false,
          error: errorMsg,
        }));
        return false;
      }
    } catch (error: any) {
      // Handle network or other errors
      let errorMsg = 'Login failed. Please try again.';
      
      if (error.name === 'TypeError' && error.message.includes('fetch')) {
        errorMsg = 'Unable to connect to the server. Please check your internet connection.';
      } else if (error.message) {
        errorMsg = error.message;
      }
      
      console.error('Login error:', error);
      
      setAuthState(prev => ({
        ...prev,
        loading: false,
        error: normalizeError(errorMsg),
      }));
      return false;
    }
  };

  // Register function
  const register = async (data: RegisterData): Promise<boolean> => {
    setAuthState(prev => ({ ...prev, loading: true, error: null }));
    
    try {
      const response = await authApi.register(data);
      
      if (response.success && response.data) {
        // Handle the nested data structure from the API
        // Backend wraps response in: { data: { user, message }, request_id, timestamp }
        // Note: Registration doesn't return a token, just user info
        const responseData = response.data.data || response.data;
        const { user, token } = responseData;
        
        if (!user || typeof user !== 'object') {
          const errorMsg = normalizeError('Invalid user information received');
          console.error('Registration error: Invalid user data', { user, response });
          setAuthState(prev => ({ ...prev, loading: false, error: errorMsg }));
          return false;
        }
        
        // Check if token is provided (some backends auto-login after registration)
        if (token && typeof token === 'string') {
          // Auto-login after registration
          try {
            localStorage.setItem('auth_token', token);
            localStorage.setItem('user_data', JSON.stringify(user));
            authApi.setAuthToken(token);
            
            setAuthState({
              isAuthenticated: true,
              user,
              token,
              loading: false,
              error: null,
              bookmarks: [],
              bookmarksLoading: false,
            });
          } catch (storageError) {
            console.error('Failed to save registration data:', storageError);
            setAuthState(prev => ({ ...prev, loading: false, error: normalizeError('Failed to save registration session') }));
            return false;
          }
        } else {
          // Registration successful but no auto-login
          setAuthState(prev => ({ ...prev, loading: false, error: null }));
        }
        
        console.log('Registration successful');
        return true;
      } else {
        // Handle API error response
        const errorMsg = normalizeError(response.error || 'Registration failed. Please try again.');
        console.error('Registration failed:', { error: response.error, response });
        
        setAuthState(prev => ({
          ...prev,
          loading: false,
          error: errorMsg,
        }));
        return false;
      }
    } catch (error: any) {
      // Handle network or other errors
      let errorMsg = 'Registration failed. Please try again.';
      
      if (error.name === 'TypeError' && error.message.includes('fetch')) {
        errorMsg = 'Unable to connect to the server. Please check your internet connection.';
      } else if (error.message) {
        errorMsg = error.message;
      }
      
      console.error('Registration error:', error);
      
      setAuthState(prev => ({
        ...prev,
        loading: false,
        error: normalizeError(errorMsg),
      }));
      return false;
    }
  };

  // Logout function
  const logout = () => {
    // Clear localStorage
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user_data');
    
    // Clear API token
    authApi.clearAuthToken();
    
    // Update state
    setAuthState({
      isAuthenticated: false,
      user: null,
      token: null,
      loading: false,
      error: null,
      bookmarks: [],
      bookmarksLoading: false,
    });
  };

  // Refresh token function
  const refreshToken = async (): Promise<boolean> => {
    try {
      const currentToken = authState.token;
      if (!currentToken) return false;
      
      const response = await authApi.refreshToken(currentToken);
      
      if (response.success && response.data) {
        const responseData = response.data.data || response.data;
        const { token, user } = responseData;
        
        // Update localStorage
        localStorage.setItem('auth_token', token);
        if (user) {
          localStorage.setItem('user_data', JSON.stringify(user));
        }
        
        // Set token for API requests
        authApi.setAuthToken(token);
        
        // Update state
        setAuthState(prev => ({
          ...prev,
          token,
          user: user || prev.user,
        }));
        
        return true;
      }
      
      return false;
    } catch (error) {
      console.error('Token refresh failed:', error);
      logout(); // Force logout on refresh failure
      return false;
    }
  };

  // Clear error function
  const clearError = () => {
    setAuthState(prev => ({ ...prev, error: null }));
  };

  // Load user bookmarks (with authentication check)
  const loadBookmarks = async (): Promise<void> => {
    if (!authState.isAuthenticated) return;
    await loadBookmarksDirectly();
  };

  // Load user bookmarks directly (without authentication check)
  const loadBookmarksDirectly = async (): Promise<void> => {
    setAuthState(prev => ({ ...prev, bookmarksLoading: true }));
    
    try {
      const response = await authApi.getBookmarks();
      
      if (response.success && response.data) {
        // Extract article IDs from bookmarks - handle nested structure
        const responseData = response.data as any;
        const bookmarksData = responseData.data || response.data;
        // Need to handle the double nesting here too
        const bookmarksArray = bookmarksData.data || bookmarksData;
        const bookmarkIds = Array.isArray(bookmarksArray) 
          ? bookmarksArray.map((bookmark: any) => bookmark.article_id || bookmark.id)
          : [];
          
        setAuthState(prev => ({ 
          ...prev, 
          bookmarks: bookmarkIds,
          bookmarksLoading: false 
        }));
        console.log('AuthContext: Loaded bookmarks:', bookmarkIds);
      } else {
        console.error('Failed to load bookmarks:', response.error);
        setAuthState(prev => ({ ...prev, bookmarksLoading: false }));
      }
    } catch (error) {
      console.error('Error loading bookmarks:', error);
      setAuthState(prev => ({ ...prev, bookmarksLoading: false }));
    }
  };

  // Add bookmark
  const addBookmark = async (articleId: string): Promise<boolean> => {
    if (!authState.isAuthenticated) return false;
    
    try {
      const response = await authApi.addBookmark(articleId);
      
      if (response.success) {
        setAuthState(prev => ({ 
          ...prev, 
          bookmarks: [...prev.bookmarks, articleId]
        }));
        return true;
      } else {
        console.error('Failed to add bookmark:', response.error);
        return false;
      }
    } catch (error) {
      console.error('Error adding bookmark:', error);
      return false;
    }
  };

  // Remove bookmark
  const removeBookmark = async (articleId: string): Promise<boolean> => {
    if (!authState.isAuthenticated) return false;

    try {
      const response = await authApi.removeBookmark(articleId);
      
      if (response.success) {
        const newBookmarks = authState.bookmarks.filter(id => id !== articleId);
        
        setAuthState(prev => ({ 
          ...prev, 
          bookmarks: newBookmarks
        }));
        return true;
      } else {
        console.error('Failed to remove bookmark:', response.error);
        return false;
      }
    } catch (error) {
      console.error('Error removing bookmark:', error);
      return false;
    }
  };

  // Check if article is bookmarked
  const isBookmarked = (articleId: string): boolean => {
    return authState.bookmarks.includes(articleId);
  };

  // Context value
  const contextValue: AuthContextType = {
    authState,
    login,
    register,
    logout,
    refreshToken,
    clearError,
    addBookmark,
    removeBookmark,
    isBookmarked,
    loadBookmarks,
    isLoading: authState.loading,
    hasError: !!authState.error,
  };

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
};

// Custom hook to use auth context
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

// Higher-order component for protected routes
export const withAuth = <P extends object>(Component: React.ComponentType<P>) => {
  return (props: P) => {
    const { authState } = useAuth();
    
    if (authState.loading) {
      return <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-600"></div>
      </div>;
    }
    
    if (!authState.isAuthenticated) {
      return <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-4">Authentication Required</h2>
          <p className="text-gray-600">Please log in to access this page.</p>
        </div>
      </div>;
    }
    
    return <Component {...props} />;
  };
};
