'use client';

import { useEffect, useRef, useCallback } from 'react';

let ws = null;
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
const reconnectDelay = 1000;

// WebSocket event listeners
const eventListeners = new Map();

export const useRealTimeUpdates = (eventType, callback, isAuthenticated = true) => {
  const callbackRef = useRef(callback);
  
  // Update callback ref when callback changes
  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  useEffect(() => {
    // Only connect if authenticated
    if (!isAuthenticated) {
      return;
    }
    
    // Initialize WebSocket connection if not already connected
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      connectWebSocket();
    }

    // Add event listener
    if (!eventListeners.has(eventType)) {
      eventListeners.set(eventType, new Set());
    }
    eventListeners.get(eventType).add(callbackRef);

    // Cleanup function
    return () => {
      const listeners = eventListeners.get(eventType);
      if (listeners) {
        listeners.delete(callbackRef);
        if (listeners.size === 0) {
          eventListeners.delete(eventType);
        }
      }
    };
  }, [eventType]);

  // Cleanup WebSocket on unmount if no more listeners
  useEffect(() => {
    return () => {
      if (eventListeners.size === 0 && ws) {
        ws.close();
        ws = null;
      }
    };
  }, []);
};

const connectWebSocket = () => {
  try {
    console.log('Attempting to connect to WebSocket...');
    
    // Check if we have a session cookie (basic auth check)
    const cookies = document.cookie.split(';');
    const sessionCookie = cookies.find(cookie => cookie.trim().startsWith('session='));
    
    if (!sessionCookie) {
      console.log('No session cookie found, skipping WebSocket connection');
      return;
    }
    
    ws = new WebSocket('ws://localhost:8080/api/v1/ws');
    
    ws.onopen = () => {
      console.log('WebSocket connected successfully');
      reconnectAttempts = 0;
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        const { type, payload } = data;
        
        // Notify all listeners for this event type
        const listeners = eventListeners.get(type);
        if (listeners) {
          listeners.forEach(callbackRef => {
            if (callbackRef.current) {
              callbackRef.current(payload);
            }
          });
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    ws.onclose = (event) => {
      console.log('WebSocket disconnected:', event.code, event.reason);
      if (reconnectAttempts < maxReconnectAttempts) {
        setTimeout(() => {
          reconnectAttempts++;
          connectWebSocket();
        }, reconnectDelay * reconnectAttempts);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      console.error('WebSocket readyState:', ws.readyState);
      console.error('WebSocket URL:', ws.url);
    };
  } catch (error) {
    console.error('Failed to connect WebSocket:', error);
  }
};

// Export the hook with the correct name
export const useWebSocket = (eventType, callback, isAuthenticated) => {
  return useRealTimeUpdates(eventType, callback, isAuthenticated);
};

// Simple hook for just connecting without listening to specific events
export const useWebSocketConnection = (isAuthenticated = true) => {
  useEffect(() => {
    // Only connect if authenticated
    if (!isAuthenticated) {
      // Close existing connection if user is not authenticated
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.close();
        ws = null;
      }
      return;
    }
    
    // Initialize WebSocket connection if not already connected
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      connectWebSocket();
    }
  }, [isAuthenticated]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.close();
        ws = null;
      }
    };
  }, []);
};
