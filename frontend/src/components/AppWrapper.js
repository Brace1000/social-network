'use client';

import React from 'react';
import { AuthProvider } from '../store/AuthContext';
import { NotificationProvider } from '../store/NotificationContext';

export default function AppWrapper({ children }) {
  return (
    <AuthProvider>
      <NotificationProvider>
        {children}
      </NotificationProvider>
    </AuthProvider>
  );
} 