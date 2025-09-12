import { useState } from 'react';

export const useAuthModal = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [pendingAction, setPendingAction] = useState<(() => void) | null>(null);

  const openLoginModal = (onSuccess?: () => void) => {
    setMode('login');
    setIsOpen(true);
    if (onSuccess) {
      setPendingAction(() => onSuccess);
    }
  };

  const openRegisterModal = (onSuccess?: () => void) => {
    setMode('register');
    setIsOpen(true);
    if (onSuccess) {
      setPendingAction(() => onSuccess);
    }
  };

  const closeModal = () => {
    setIsOpen(false);
    setPendingAction(null);
  };

  const handleSuccess = () => {
    if (pendingAction) {
      pendingAction();
    }
    closeModal();
  };

  return {
    isOpen,
    mode,
    openLoginModal,
    openRegisterModal,
    closeModal,
    handleSuccess,
  };
};
