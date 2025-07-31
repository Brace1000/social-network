'use client';

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { chatAPI } from '../lib/api';
import { useAuth } from './AuthContext';

const ChatContext = createContext();

export const useChat = () => {
  const context = useContext(ChatContext);
  if (!context) {
    throw new Error('useChat must be used within a ChatProvider');
  }
  return context;
};

export const ChatProvider = ({ children }) => {
  const [conversations, setConversations] = useState([]);
  const [messages, setMessages] = useState({});
  const [activeConversation, setActiveConversation] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [websocket, setWebsocket] = useState(null);
  const { user, isAuthenticated } = useAuth();

  // Handle incoming WebSocket messages
  const handleIncomingMessage = useCallback((message) => {
    const conversationKey = message.groupId
      ? `group_${message.groupId}`
      : `private_${message.senderId === user.id ? message.recipientId : message.senderId}`;

    setMessages(prev => ({
      ...prev,
      [conversationKey]: [...(prev[conversationKey] || []), message]
    }));

    // Update conversation list with new message
    setConversations(prev =>
      prev.map(conv => {
        const convKey = conv.type === 'group' ? conv.groupId : conv.userId;
        const msgKey = message.groupId || (message.senderId === user.id ? message.recipientId : message.senderId);

        if (convKey === msgKey) {
          return {
            ...conv,
            lastMessage: message.content,
            lastMessageTime: message.createdAt,
            unreadCount: conv.unreadCount + 1
          };
        }
        return conv;
      })
    );
  }, [user]);

  // Initialize WebSocket connection
  useEffect(() => {
    console.log('WebSocket useEffect triggered', { isAuthenticated, user: user?.id });

    // Only run on client side
    if (typeof window === 'undefined') {
      console.log('WebSocket connection skipped: server-side rendering');
      return;
    }

    if (!isAuthenticated || !user) {
      console.log('WebSocket connection skipped: not authenticated or no user');
      return;
    }

    // Get session token from cookies
    const getSessionToken = () => {
      if (typeof document === 'undefined') {
        console.log('Document not available (SSR)');
        return null;
      }

      const cookies = document.cookie.split(';');
      console.log('All cookies:', document.cookie);

      for (let cookie of cookies) {
        const [name, value] = cookie.trim().split('=');
        console.log('Checking cookie:', { name, value });
        if (name === 'social_network_session') {
          console.log('Found session token:', value);
          return value;
        }
      }
      console.log('No session token found in cookies');
      return null;
    };

    const sessionToken = getSessionToken();
    if (!sessionToken) {
      console.error('No session token found for WebSocket connection');
      return;
    }

    console.log('Attempting WebSocket connection with token:', sessionToken);
    const ws = new WebSocket(`ws://localhost:8080/api/v1/ws?token=${sessionToken}`);

    ws.onopen = () => {
      console.log('WebSocket connected successfully');
      setWebsocket(ws);
      setError(''); // Clear any previous errors
    };

    ws.onmessage = (event) => {
      console.log('WebSocket message received:', event.data);
      try {
        const message = JSON.parse(event.data);
        handleIncomingMessage(message);
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    ws.onclose = (event) => {
      console.log('WebSocket disconnected', { code: event.code, reason: event.reason });
      setWebsocket(null);
      if (event.code !== 1000) { // 1000 is normal closure
        setError('Connection lost. Please refresh the page.');
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error details:', {
        error,
        readyState: ws.readyState,
        url: ws.url,
        protocol: ws.protocol
      });
      setError('Connection lost. Please refresh the page.');
    };

    return () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
    };
  }, [isAuthenticated, user, handleIncomingMessage]);



  // Load conversations
  const loadConversations = useCallback(async () => {
    if (!isAuthenticated) return;

    try {
      setLoading(true);
      setError('');
      const data = await chatAPI.getConversations();
      setConversations(data || []);
    } catch (err) {
      console.error('Error loading conversations:', err);
      setError('Failed to load conversations');
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated]);

  // Load messages for a conversation
  const loadMessages = useCallback(async (conversationId, type = 'private') => {
    if (!conversationId) return;

    const conversationKey = `${type}_${conversationId}`;

    try {
      setLoading(true);
      setError('');

      let data;
      if (type === 'private') {
        data = await chatAPI.getPrivateConversation(conversationId);
      } else {
        data = await chatAPI.getGroupConversation(conversationId);
      }

      setMessages(prev => ({
        ...prev,
        [conversationKey]: data || []
      }));
    } catch (err) {
      console.error('Error loading messages:', err);
      setError('Failed to load messages');
    } finally {
      setLoading(false);
    }
  }, []);

  // Send a message
  const sendMessage = useCallback((content, recipientId = null, groupId = null) => {
    if (!websocket || websocket.readyState !== WebSocket.OPEN) {
      setError('Connection lost. Please refresh the page.');
      return false;
    }

    const message = {
      type: 'message',
      content,
      recipientId,
      groupId,
      senderId: user.id,
      createdAt: new Date().toISOString()
    };

    try {
      websocket.send(JSON.stringify(message));
      return true;
    } catch (error) {
      console.error('Error sending message:', error);
      setError('Failed to send message');
      return false;
    }
  }, [websocket, user]);

  // Search users for new conversations
  const searchUsers = useCallback(async (query) => {
    if (!query.trim()) return [];

    try {
      const data = await chatAPI.searchUsers(query);
      return data || [];
    } catch (err) {
      console.error('Error searching users:', err);
      return [];
    }
  }, []);

  // Check if user can message another user
  const canMessage = useCallback(async (userId) => {
    try {
      const data = await chatAPI.canMessage(userId);
      return data?.canMessage || false;
    } catch (err) {
      console.error('Error checking message permissions:', err);
      return false;
    }
  }, []);

  // Start a new conversation
  const startConversation = useCallback((user) => {
    const newConversation = {
      type: 'private',
      userId: user.id,
      name: `${user.firstName} ${user.lastName}`,
      avatarPath: user.avatarPath,
      lastMessage: '',
      lastMessageTime: '',
      unreadCount: 0
    };

    // Check if conversation already exists
    const existingConv = conversations.find(conv =>
      conv.type === 'private' && conv.userId === user.id
    );

    if (!existingConv) {
      setConversations(prev => [newConversation, ...prev]);
    }

    setActiveConversation(newConversation);
    return newConversation;
  }, [conversations]);

  const value = {
    conversations,
    messages,
    activeConversation,
    loading,
    error,
    websocket,
    loadConversations,
    loadMessages,
    sendMessage,
    searchUsers,
    canMessage,
    startConversation,
    setActiveConversation,
    setError
  };

  return (
    <ChatContext.Provider value={value}>
      {children}
    </ChatContext.Provider>
  );
};