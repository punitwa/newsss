import React, { createContext, useContext, ReactNode } from 'react';
import { useAuthModal } from '@/hooks/useAuthModal';
import AuthModal from '@/components/auth/AuthModal';

interface BookmarkContextType {
  handleAuthRequired: (onSuccess?: () => void) => void;
}

const BookmarkContext = createContext<BookmarkContextType | undefined>(undefined);

interface BookmarkProviderProps {
  children: ReactNode;
}

export const BookmarkProvider: React.FC<BookmarkProviderProps> = ({ children }) => {
  const { 
    isOpen, 
    mode, 
    openLoginModal, 
    closeModal, 
    handleSuccess 
  } = useAuthModal();

  const handleAuthRequired = (onSuccess?: () => void) => {
    openLoginModal(onSuccess);
  };

  return (
    <BookmarkContext.Provider value={{ handleAuthRequired }}>
      {children}
      <AuthModal
        isOpen={isOpen}
        onClose={closeModal}
        initialMode={mode}
        onSuccess={handleSuccess}
      />
    </BookmarkContext.Provider>
  );
};

export const useBookmarkContext = (): BookmarkContextType => {
  const context = useContext(BookmarkContext);
  if (context === undefined) {
    throw new Error('useBookmarkContext must be used within a BookmarkProvider');
  }
  return context;
};
