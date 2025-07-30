'use client';

import React, { createContext, useContext, useReducer, useCallback } from 'react';

// Notification reducer
const notificationReducer = (state, action) => {
  switch (action.type) {
    case 'ADD_NOTIFICATION':
      return {
        ...state,
        notifications: [action.payload, ...state.notifications]
      };
    
    case 'REMOVE_NOTIFICATION':
      return {
        ...state,
        notifications: state.notifications.filter(notif => notif.id !== action.payload)
      };
    
    case 'MARK_AS_READ':
      return {
        ...state,
        notifications: state.notifications.map(notif => 
          notif.id === action.payload ? { ...notif, read: true } : notif
        )
      };
    
    case 'CLEAR_ALL':
      return {
        ...state,
        notifications: []
      };
    
    default:
      return state;
  }
};

// Create context
const NotificationContext = createContext();

// Provider component
export const NotificationProvider = ({ children }) => {
  const [state, dispatch] = useReducer(notificationReducer, {
    notifications: []
  });

  const addNotification = useCallback((notification) => {
    dispatch({
      type: 'ADD_NOTIFICATION',
      payload: {
        id: notification.id || Date.now().toString(),
        message: notification.message,
        actorId: notification.actorId,
        notifType: notification.notifType,
        read: notification.read || false,
        timestamp: notification.timestamp || new Date()
      }
    });
  }, []);

  const removeNotification = useCallback((id) => {
    dispatch({
      type: 'REMOVE_NOTIFICATION',
      payload: id
    });
  }, []);

  const markAsRead = useCallback((id) => {
    dispatch({
      type: 'MARK_AS_READ',
      payload: id
    });
  }, []);

  const clearAll = useCallback(() => {
    dispatch({
      type: 'CLEAR_ALL'
    });
  }, []);

  const value = {
    notifications: state.notifications,
    addNotification,
    removeNotification,
    markAsRead,
    clearAll,
    unreadCount: state.notifications.filter(notif => !notif.read).length
  };

  return (
    <NotificationContext.Provider value={value}>
      {children}
    </NotificationContext.Provider>
  );
};

// Custom hook to use the notification context
export const useNotification = () => {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotification must be used within a NotificationProvider');
  }
  return context;
};