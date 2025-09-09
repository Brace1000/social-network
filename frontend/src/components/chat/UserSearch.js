'use client';

import React, { useState, useEffect, useRef } from 'react';
import { chatAPI } from '../../lib/api';

const UserSearch = ({ onSelectUser, onClose }) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const searchRef = useRef(null);
  const inputRef = useRef(null);

  // Handle clicks outside the search
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (searchRef.current && !searchRef.current.contains(event.target)) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [onClose]);

  // Focus input when component mounts
  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  // Search users with debouncing
  useEffect(() => {
    if (!searchQuery.trim()) {
      setSearchResults([]);
      return;
    }

    const timeoutId = setTimeout(async () => {
      try {
        setLoading(true);
        setError(null);
        const results = await chatAPI.searchUsers(searchQuery.trim());
        setSearchResults(results || []);
      } catch (err) {
        console.error('Error searching users:', err);
        setError('Failed to search users');
        setSearchResults([]);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [searchQuery]);

  const handleUserSelect = (user) => {
    onSelectUser(user);
    onClose();
  };

  const getInitials = (firstName, lastName) => {
    return `${firstName[0]}${lastName[0]}`.toUpperCase();
  };

  const styles = {
    overlay: {
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: 'rgba(0,0,0,0.5)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 1000,
    },
    container: {
      backgroundColor: 'white',
      borderRadius: '12px',
      boxShadow: '0 4px 20px rgba(0,0,0,0.15)',
      width: '90%',
      maxWidth: '500px',
      maxHeight: '80vh',
      overflow: 'hidden',
    },
    header: {
      padding: '20px',
      borderBottom: '1px solid #eee',
      backgroundColor: '#f8f9fa',
    },
    title: {
      fontSize: '18px',
      fontWeight: '600',
      color: '#333',
      margin: '0 0 16px 0',
    },
    searchInput: {
      width: '100%',
      padding: '12px 16px',
      border: '1px solid #ddd',
      borderRadius: '8px',
      fontSize: '14px',
      outline: 'none',
    },
    content: {
      maxHeight: '400px',
      overflowY: 'auto',
    },
    loadingState: {
      padding: '40px 20px',
      textAlign: 'center',
      color: '#666',
    },
    emptyState: {
      padding: '40px 20px',
      textAlign: 'center',
      color: '#666',
    },
    errorState: {
      padding: '40px 20px',
      textAlign: 'center',
      color: '#dc3545',
    },
    userItem: {
      padding: '16px 20px',
      borderBottom: '1px solid #f0f0f0',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      gap: '12px',
      transition: 'background-color 0.2s',
    },
    userItemHover: {
      backgroundColor: '#f5f5f5',
    },
    userItemDisabled: {
      cursor: 'not-allowed',
      opacity: 0.6,
    },
    avatar: {
      width: '48px',
      height: '48px',
      borderRadius: '50%',
      backgroundColor: '#007bff',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      color: 'white',
      fontSize: '16px',
      fontWeight: '600',
      flexShrink: 0,
    },
    userInfo: {
      flex: 1,
      minWidth: 0,
    },
    userName: {
      fontSize: '14px',
      fontWeight: '600',
      color: '#333',
      marginBottom: '4px',
    },
    userDetails: {
      fontSize: '12px',
      color: '#666',
      display: 'flex',
      alignItems: 'center',
      gap: '8px',
    },
    badge: {
      backgroundColor: '#e3f2fd',
      color: '#1976d2',
      padding: '2px 6px',
      borderRadius: '4px',
      fontSize: '10px',
      fontWeight: '600',
    },
    cannotMessageBadge: {
      backgroundColor: '#ffebee',
      color: '#c62828',
    },
    footer: {
      padding: '16px 20px',
      borderTop: '1px solid #eee',
      backgroundColor: '#f8f9fa',
      textAlign: 'right',
    },
    closeButton: {
      backgroundColor: '#6c757d',
      color: 'white',
      border: 'none',
      borderRadius: '6px',
      padding: '8px 16px',
      fontSize: '14px',
      cursor: 'pointer',
      transition: 'background-color 0.2s',
    },
  };

  return (
    <div style={styles.overlay}>
      <div ref={searchRef} style={styles.container}>
        <div style={styles.header}>
          <h3 style={styles.title}>Start a new conversation</h3>
          <input
            ref={inputRef}
            type="text"
            style={styles.searchInput}
            placeholder="Search for users by name..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        <div style={styles.content}>
          {loading ? (
            <div style={styles.loadingState}>Searching users...</div>
          ) : error ? (
            <div style={styles.errorState}>{error}</div>
          ) : !searchQuery.trim() ? (
            <div style={styles.emptyState}>
              <p>Type a name to search for users</p>
            </div>
          ) : searchResults.length === 0 ? (
            <div style={styles.emptyState}>
              <p>No users found</p>
              <p style={{ fontSize: '12px', marginTop: '8px' }}>
                Try searching with a different name
              </p>
            </div>
          ) : (
            searchResults.map((user) => (
              <div
                key={user.id}
                style={{
                  ...styles.userItem,
                  ...(user.canMessage ? {} : styles.userItemDisabled),
                }}
                onClick={() => user.canMessage && handleUserSelect(user)}
                onMouseEnter={(e) => {
                  if (user.canMessage) {
                    e.currentTarget.style.backgroundColor = styles.userItemHover.backgroundColor;
                  }
                }}
                onMouseLeave={(e) => {
                  if (user.canMessage) {
                    e.currentTarget.style.backgroundColor = 'transparent';
                  }
                }}
              >
                <div style={styles.avatar}>
                  {user.avatarPath ? (
                    <img 
                      src={user.avatarPath} 
                      alt={`${user.firstName} ${user.lastName}`}
                      style={{ width: '100%', height: '100%', borderRadius: '50%' }}
                    />
                  ) : (
                    getInitials(user.firstName, user.lastName)
                  )}
                </div>
                
                <div style={styles.userInfo}>
                  <div style={styles.userName}>
                    {user.firstName} {user.lastName}
                    {user.nickname && (
                      <span style={{ fontWeight: 'normal', color: '#666' }}>
                        {' '}({user.nickname})
                      </span>
                    )}
                  </div>
                  <div style={styles.userDetails}>
                    {user.isPublic && (
                      <span style={styles.badge}>Public Profile</span>
                    )}
                    {user.canMessage ? (
                      <span style={styles.badge}>Can Message</span>
                    ) : (
                      <span style={{...styles.badge, ...styles.cannotMessageBadge}}>
                        Cannot Message
                      </span>
                    )}
                  </div>
                </div>
              </div>
            ))
          )}
        </div>

        <div style={styles.footer}>
          <button
            style={styles.closeButton}
            onClick={onClose}
            onMouseEnter={(e) => {
              e.target.style.backgroundColor = '#5a6268';
            }}
            onMouseLeave={(e) => {
              e.target.style.backgroundColor = '#6c757d';
            }}
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

export default UserSearch;
