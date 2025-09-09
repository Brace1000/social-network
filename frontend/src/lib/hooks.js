'use client';

import { useState, useEffect, useCallback } from 'react';
import { followAPI, notificationAPI } from './api';
import { useAuth } from '../store/AuthContext';
import { useRealTimeUpdates } from './websocket';

export const useFollowRequests = () => {
  const [followRequests, setFollowRequests] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const { isAuthenticated } = useAuth();

  const fetchFollowRequests = useCallback(async () => {
    // Only fetch follow requests if user is authenticated
    if (!isAuthenticated) {
      setFollowRequests([]);
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const response = await followAPI.getFollowRequests();
      // The backend returns the array directly, not wrapped in a data property
      const requests = Array.isArray(response) ? response : [];
      setFollowRequests(requests);
      if (requests.length === 0) {
        setError('No follow requests found');
      }
    } catch (err) {
      console.error('Error fetching follow requests:', err);
      setError('No follow requests found');
      setFollowRequests([]);
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated]);

  const handleFollowRequestAction = useCallback(async (requestId, action) => {
    try {
      if (action === 'accept') {
        await followAPI.acceptFollowRequest(requestId);
      } else if (action === 'decline') {
        await followAPI.declineFollowRequest(requestId);
      } else if (action === 'cancel') {
        await followAPI.cancelFollowRequest(requestId);
      }
      
      // Refresh the follow requests list
      await fetchFollowRequests();
    } catch (err) {
      console.error(`Error ${action}ing follow request:`, err);
      throw err;
    }
  }, [fetchFollowRequests]);

  // Register for real-time follow request updates
  useRealTimeUpdates('followRequestUpdate', useCallback(() => {
    console.log('Refreshing follow requests due to real-time update');
    fetchFollowRequests();
  }, [fetchFollowRequests]));

  // Fetch follow requests on mount and when authentication changes
  useEffect(() => {
    fetchFollowRequests();
  }, [fetchFollowRequests]);

  return {
    followRequests,
    loading,
    error,
    handleFollowRequestAction,
    refresh: fetchFollowRequests
  };
};

export const useMyFollowRequests = () => {
  const [myFollowRequests, setMyFollowRequests] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const { isAuthenticated } = useAuth();

  const fetchMyFollowRequests = useCallback(async () => {
    // Only fetch follow requests if user is authenticated
    if (!isAuthenticated) {
      setMyFollowRequests([]);
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const response = await followAPI.getMyFollowRequests();
      // The backend returns the array directly, not wrapped in a data property
      const requests = Array.isArray(response) ? response : [];
      setMyFollowRequests(requests);
      if (requests.length === 0) {
        setError('No follow requests found');
      }
    } catch (err) {
      console.error('Error fetching my follow requests:', err);
      setError('No follow requests found');
      setMyFollowRequests([]);
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated]);

  const cancelFollowRequest = useCallback(async (requestId) => {
    try {
      await followAPI.cancelFollowRequest(requestId);
      // Refresh the list
      await fetchMyFollowRequests();
    } catch (err) {
      console.error('Error canceling follow request:', err);
      throw err;
    }
  }, [fetchMyFollowRequests]);

  // Register for real-time follow request updates
  useRealTimeUpdates('followRequestUpdate', useCallback(() => {
    console.log('Refreshing my follow requests due to real-time update');
    fetchMyFollowRequests();
  }, [fetchMyFollowRequests]));

  // Fetch my follow requests on mount and when authentication changes
  useEffect(() => {
    fetchMyFollowRequests();
  }, [fetchMyFollowRequests]);

  return {
    myFollowRequests,
    loading,
    error,
    cancelFollowRequest,
    refresh: fetchMyFollowRequests
  };
};

export const useNotifications = () => {
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const { isAuthenticated } = useAuth();

  const fetchNotifications = useCallback(async () => {
    // Only fetch notifications if user is authenticated
    if (!isAuthenticated) {
      setNotifications([]);
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const response = await notificationAPI.getNotifications();
      const notifs = Array.isArray(response) ? response : [];
      setNotifications(notifs);
    } catch (err) {
      console.error('Error fetching notifications:', err);
      setError('Failed to load notifications');
      setNotifications([]);
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated]);

  const markAsRead = useCallback(async (notificationId) => {
    try {
      await notificationAPI.markAsRead(notificationId);
      // Update the notification locally
      setNotifications(prev => 
        prev.map(notif => 
          notif.id === notificationId ? { ...notif, read: true } : notif
        )
      );
    } catch (err) {
      console.error('Error marking notification as read:', err);
      throw err;
    }
  }, []);

  // Register for real-time notification updates
  useRealTimeUpdates('notification', useCallback((payload) => {
    console.log('Received real-time notification:', payload);
    // Add the new notification to the list
    setNotifications(prev => [payload, ...prev]);
  }, []));

  // Fetch notifications on mount and when authentication changes
  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  return {
    notifications,
    loading,
    error,
    markAsRead,
    refresh: fetchNotifications,
    unreadCount: notifications.filter(notif => !notif.read).length
  };
};