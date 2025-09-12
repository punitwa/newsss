import { LoginCredentials, RegisterData, User } from '@/contexts/AuthContext';

const API_BASE_URL = '/api/v1';

interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

interface LoginResponseData {
  token: string;
  user: User;
  expires_in: number;
  token_type: string;
}

interface LoginResponse {
  data: LoginResponseData;
  request_id: string;
  timestamp: string;
}

interface RegisterResponseData {
  user: User;
  message: string;
  token?: string; // Optional since backend might not auto-login
}

interface RegisterResponse {
  data: RegisterResponseData;
  request_id: string;
  timestamp: string;
}

class AuthApiService {
  private authToken: string | null = null;

  // Set authentication token for requests
  setAuthToken(token: string) {
    this.authToken = token;
  }

  // Clear authentication token
  clearAuthToken() {
    this.authToken = null;
  }

  // Get headers with authentication
  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    };

    if (this.authToken) {
      headers['Authorization'] = `Bearer ${this.authToken}`;
    }

    return headers;
  }

  // Handle API response
  private async handleResponse<T>(response: Response): Promise<ApiResponse<T>> {
    try {
      // Handle empty responses
      const contentType = response.headers.get('content-type');
      let data;

      if (contentType && contentType.includes('application/json')) {
        data = await response.json();
      } else {
        data = { message: await response.text() };
      }

      if (response.ok) {
        return {
          success: true,
          data,
        };
      } else {
        // Handle different HTTP status codes
        let errorMessage = data.error || data.message || 'An error occurred';

        switch (response.status) {
          case 400:
            errorMessage = data.error || 'Invalid request. Please check your input.';
            break;
          case 401:
            errorMessage = data.error || 'Invalid email or password.';
            break;
          case 403:
            errorMessage = data.error || 'Access denied.';
            break;
          case 404:
            errorMessage = data.error || 'Service not found.';
            break;
          case 409:
            errorMessage = data.error || 'Email address is already registered.';
            break;
          case 429:
            errorMessage = data.error || 'Too many requests. Please try again later.';
            break;
          case 500:
            errorMessage = 'Server error. Please try again later.';
            break;
          case 503:
            errorMessage = 'Service temporarily unavailable. Please try again later.';
            break;
          default:
            errorMessage = data.error || `Request failed (${response.status})`;
        }

        return {
          success: false,
          error: errorMessage,
        };
      }
    } catch (error) {
      console.error('Response parsing error:', error);
      return {
        success: false,
        error: 'Unable to process server response. Please try again.',
      };
    }
  }

  // Login user
  async login(credentials: LoginCredentials): Promise<ApiResponse<LoginResponse>> {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/login`, {
        method: 'POST',
        headers: this.getHeaders(),
        body: JSON.stringify(credentials),
      });

      return this.handleResponse<LoginResponse>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error during login',
      };
    }
  }

  // Register user
  async register(data: RegisterData): Promise<ApiResponse<RegisterResponse>> {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/register`, {
        method: 'POST',
        headers: this.getHeaders(),
        body: JSON.stringify(data),
      });

      return this.handleResponse<RegisterResponse>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error during registration',
      };
    }
  }

  // Refresh token
  async refreshToken(token: string): Promise<ApiResponse<LoginResponse>> {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: {
          ...this.getHeaders(),
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ token }),
      });

      return this.handleResponse<LoginResponse>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error during token refresh',
      };
    }
  }

  // Verify token validity
  async verifyToken(token: string): Promise<boolean> {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/status`, {
        method: 'GET',
        headers: {
          ...this.getHeaders(),
          'Authorization': `Bearer ${token}`,
        },
      });

      return response.ok;
    } catch (error) {
      return false;
    }
  }

  // Logout user
  async logout(): Promise<ApiResponse<void>> {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/logout`, {
        method: 'POST',
        headers: this.getHeaders(),
      });

      return this.handleResponse<void>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error during logout',
      };
    }
  }

  // Forgot password
  async forgotPassword(email: string): Promise<ApiResponse<{ message: string }>> {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/forgot-password`, {
        method: 'POST',
        headers: this.getHeaders(),
        body: JSON.stringify({ email }),
      });

      return this.handleResponse<{ message: string }>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error during password reset request',
      };
    }
  }

  // Reset password
  async resetPassword(token: string, newPassword: string): Promise<ApiResponse<{ message: string }>> {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/reset-password`, {
        method: 'POST',
        headers: this.getHeaders(),
        body: JSON.stringify({
          token,
          new_password: newPassword
        }),
      });

      return this.handleResponse<{ message: string }>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error during password reset',
      };
    }
  }

  // Get user profile
  async getUserProfile(): Promise<ApiResponse<User>> {
    try {
      const response = await fetch(`${API_BASE_URL}/user/profile`, {
        method: 'GET',
        headers: this.getHeaders(),
      });

      return this.handleResponse<User>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error fetching profile',
      };
    }
  }

  // Update user profile
  async updateProfile(data: Partial<User>): Promise<ApiResponse<User>> {
    try {
      const response = await fetch(`${API_BASE_URL}/user/profile`, {
        method: 'PUT',
        headers: this.getHeaders(),
        body: JSON.stringify(data),
      });

      return this.handleResponse<User>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error updating profile',
      };
    }
  }

  // Change password
  async changePassword(currentPassword: string, newPassword: string): Promise<ApiResponse<{ message: string }>> {
    try {
      const response = await fetch(`${API_BASE_URL}/user/change-password`, {
        method: 'POST',
        headers: this.getHeaders(),
        body: JSON.stringify({
          current_password: currentPassword,
          new_password: newPassword,
        }),
      });

      return this.handleResponse<{ message: string }>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error changing password',
      };
    }
  }

  // Get user bookmarks
  async getBookmarks(): Promise<ApiResponse<any>> {
    try {
      const response = await fetch(`${API_BASE_URL}/user/bookmarks`, {
        method: 'GET',
        headers: this.getHeaders(),
      });

      return this.handleResponse<any>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error fetching bookmarks',
      };
    }
  }

  // Add bookmark
  async addBookmark(articleId: string): Promise<ApiResponse<{ message: string }>> {
    try {
      const response = await fetch(`${API_BASE_URL}/user/bookmarks`, {
        method: 'POST',
        headers: this.getHeaders(),
        body: JSON.stringify({ article_id: articleId }),
      });

      return this.handleResponse<{ message: string }>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error adding bookmark',
      };
    }
  }

  // Remove bookmark
  async removeBookmark(articleId: string): Promise<ApiResponse<{ message: string }>> {
    try {
      const response = await fetch(`${API_BASE_URL}/user/bookmarks/${articleId}`, {
        method: 'DELETE',
        headers: this.getHeaders(),
      });

      return this.handleResponse<{ message: string }>(response);
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Network error removing bookmark',
      };
    }
  }
}

// Export singleton instance
export const authApi = new AuthApiService();
