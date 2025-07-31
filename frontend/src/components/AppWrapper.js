'use client';

import React from 'react';
import { AuthProvider } from '../store/AuthContext';
import { NotificationProvider } from '../store/NotificationContext';
import { ChatProvider } from '../store/ChatContext';

export default function AppWrapper({ children }) {
  return (
    <AuthProvider>
      <NotificationProvider>
        <ChatProvider>
          {children}
        </ChatProvider>
      </NotificationProvider>
    </AuthProvider>
  );
}