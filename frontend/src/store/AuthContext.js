'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';
import { userAPI } from '../lib/api';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Check if user is authenticated on mount
  const checkAuth = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Only check auth in browser environment
      if (typeof window === 'undefined') {
        setLoading(false);
        return;
      }

      const userData = await userAPI.getCurrentUser();
      setUser(userData);
    } catch (err) {
      // Handle unauthorized gracefully - this is expected when user is not logged in
      if (err.message.trim() === 'Unauthorized' || err.message.includes('401')) {
        console.log('User not authenticated - this is normal for non-logged in users');
        setUser(null);
        setError(null); // Don't set error for expected unauthorized state
      } else {
        console.error('Authentication check failed:', err.message);
        setUser(null);
        setError(err.message);
      }
    } finally {
      setLoading(false);
    }
  };

  // Login function
  const login = async (email, password) => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch('http://localhost:8080/api/v1/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Login failed');
      }

      const userData = await response.json();
      setUser(userData);
      return userData;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Logout function
  const logout = async () => {
    try {
      await fetch('http://localhost:8080/api/v1/logout', {
        method: 'POST',
        credentials: 'include',
      });
    } catch (err) {
      console.error('Logout error:', err);
    } finally {
      setUser(null);
      setError(null);
    }
  };

  // Check authentication on mount
  useEffect(() => {
    checkAuth();
  }, []);

  const value = {
    user,
    loading,
    error,
    login,
    logout,
    checkAuth,
    isAuthenticated: !!user,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};