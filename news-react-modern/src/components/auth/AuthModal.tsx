import React, { useState } from 'react';
import { X } from 'lucide-react';
import LoginForm from './LoginForm';
import RegisterForm from './RegisterForm';

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  initialMode?: 'login' | 'register';
  onSuccess?: () => void;
}

const AuthModal: React.FC<AuthModalProps> = ({ 
  isOpen, 
  onClose, 
  initialMode = 'login',
  onSuccess 
}) => {
  const [currentMode, setCurrentMode] = useState<'login' | 'register'>(initialMode);

  // Handle successful authentication
  const handleSuccess = () => {
    onSuccess?.();
    onClose();
  };

  // Handle switching between modes
  const switchToLogin = () => setCurrentMode('login');
  const switchToRegister = () => setCurrentMode('register');

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div 
        className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
        onClick={onClose}
      />
      
      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative bg-white rounded-xl shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
          {/* Close Button */}
          <button
            onClick={onClose}
            className="absolute top-4 right-4 z-10 p-2 text-gray-400 hover:text-gray-600 transition-colors"
          >
            <X className="h-5 w-5" />
          </button>

          {/* Content */}
          <div className="p-6">
            {currentMode === 'login' ? (
              <LoginForm
                onSuccess={handleSuccess}
                onSwitchToRegister={switchToRegister}
                onForgotPassword={() => {
                  // Handle forgot password - could open another modal or navigate
                  console.log('Forgot password clicked');
                }}
              />
            ) : (
              <RegisterForm
                onSuccess={handleSuccess}
                onSwitchToLogin={switchToLogin}
              />
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default AuthModal;
