'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '../../store/AuthContext';
import { useChat } from '../../store/ChatContext';
import ConversationList from '../../components/chat/ConversationList';
import ChatWindow from '../../components/chat/ChatWindow';
import UserSearch from '../../components/chat/UserSearch';

export default function MessagesPage() {
  const [showUserSearch, setShowUserSearch] = useState(false);
  const [selectedConversation, setSelectedConversation] = useState(null);
  const { user, isAuthenticated, authLoading } = useAuth();
  const {
    conversations,
    activeConversation,
    loadConversations,
    setActiveConversation,
    startConversation,
    loading,
    error
  } = useChat();
  const router = useRouter();

  // Redirect if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/auth/login');
    }
  }, [isAuthenticated, authLoading, router]);

  // Load conversations on mount
  useEffect(() => {
    if (isAuthenticated) {
      loadConversations();
    }
  }, [isAuthenticated, loadConversations]); 

  // Handle conversation selection
  const handleConversationSelect = (conversation) => {
    setSelectedConversation(conversation);
    setActiveConversation(conversation);
  };

  // Handle starting new conversation
  const handleStartConversation = (user) => {
    const conversation = startConversation(user);
    setSelectedConversation(conversation);
    setShowUserSearch(false);
  };

  if (authLoading) {
    return (
      <div style={styles.loadingContainer}>
        <div style={styles.loadingText}>Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
          <Link
            href="/home"
            style={{
              color: '#1c1e21',
              textDecoration: 'none',
              padding: '8px 16px',
              borderRadius: '6px',
              backgroundColor: '#f0f2f5',
              transition: 'background-color 0.2s',
              fontSize: '14px',
              fontWeight: '500',
              display: 'flex',
              alignItems: 'center',
              gap: '8px'
            }}
            onMouseEnter={(e) => e.target.style.backgroundColor = '#e4e6ea'}
            onMouseLeave={(e) => e.target.style.backgroundColor = '#f0f2f5'}
          >
            <span style={{ fontSize: '16px' }}>←</span>
            Home
          </Link>
          <h1 style={styles.title}>Messages</h1>
        </div>
        <button
          style={styles.newChatButton}
          onClick={() => setShowUserSearch(true)}
        >
          + New Chat
        </button>
      </div>

      <div style={styles.content}>
        {/* Sidebar with conversations */}
        <div style={styles.sidebar}>
          {showUserSearch ? (
            <UserSearch
              onSelectUser={handleStartConversation}
              onClose={() => setShowUserSearch(false)}
            />
          ) : (
            <ConversationList
              conversations={conversations}
              activeConversation={selectedConversation}
              onSelectConversation={handleConversationSelect}
              loading={loading}
            />
          )}
        </div>

        {/* Main chat area */}
        <div style={styles.chatArea}>
          {selectedConversation ? (
            <ChatWindow conversation={selectedConversation} />
          ) : (
            <div style={styles.emptyState}>
              <div style={styles.emptyStateContent}>
                <h3 style={styles.emptyStateTitle}>Select a conversation</h3>
                <p style={styles.emptyStateText}>
                  Choose a conversation from the sidebar or start a new chat
                </p>
                <button
                  style={styles.startChatButton}
                  onClick={() => setShowUserSearch(true)}
                >
                  Start New Chat
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {error && (
        <div style={styles.errorBanner}>
          {error}
        </div>
      )}
    </div>
  );
}

const styles = {
  container: {
    height: '100vh',
    display: 'flex',
    flexDirection: 'column',
    backgroundColor: '#f0f2f5',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '20px',
    backgroundColor: '#fff',
    borderBottom: '1px solid #e4e6ea',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
  },
  title: {
    margin: 0,
    fontSize: '24px',
    fontWeight: '600',
    color: '#1c1e21',
  },
  newChatButton: {
    backgroundColor: '#1877f2',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    padding: '10px 16px',
    fontSize: '14px',
    fontWeight: '600',
    cursor: 'pointer',
    transition: 'background-color 0.2s',
  },
  content: {
    display: 'flex',
    flex: 1,
    overflow: 'hidden',
  },
  sidebar: {
    width: '350px',
    backgroundColor: '#fff',
    borderRight: '1px solid #e4e6ea',
    display: 'flex',
    flexDirection: 'column',
  },
  chatArea: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
  },
  emptyState: {
    flex: 1,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#fff',
  },
  emptyStateContent: {
    textAlign: 'center',
    maxWidth: '300px',
  },
  emptyStateTitle: {
    margin: '0 0 10px 0',
    fontSize: '20px',
    fontWeight: '600',
    color: '#1c1e21',
  },
  emptyStateText: {
    margin: '0 0 20px 0',
    fontSize: '14px',
    color: '#65676b',
    lineHeight: '1.4',
  },
  startChatButton: {
    backgroundColor: '#1877f2',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    padding: '12px 24px',
    fontSize: '14px',
    fontWeight: '600',
    cursor: 'pointer',
  },
  loadingContainer: {
    height: '100vh',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#f0f2f5',
  },
  loadingText: {
    fontSize: '16px',
    color: '#65676b',
  },
  errorBanner: {
    position: 'fixed',
    bottom: '20px',
    left: '50%',
    transform: 'translateX(-50%)',
    backgroundColor: '#e74c3c',
    color: 'white',
    padding: '12px 20px',
    borderRadius: '8px',
    fontSize: '14px',
    zIndex: 1000,
  },
};