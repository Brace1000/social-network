'use client';

import React, { useState, useRef, useEffect } from 'react';
import { useFollowRequests, useNotifications } from '../lib/hooks';
import { useRouter } from 'next/navigation';

const styles = {
  container: {
    position: 'relative',
  },
  notificationButton: {
    background: 'none',
    border: 'none',
    cursor: 'pointer',
    position: 'relative',
    padding: '8px',
    borderRadius: '50%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '20px',
    color: '#fff',
    transition: 'background-color 0.2s',
    width: '48px',
    height: '48px',
  },
  bellIcon: {
    width: '28px',
    height: '28px',
    fill: 'currentColor',
  },
  badge: {
    position: 'absolute',
    top: '-2px',
    right: '-2px',
    background: '#ff4444',
    color: 'white',
    borderRadius: '50%',
    width: '18px',
    height: '18px',
    fontSize: '11px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontWeight: 'bold',
    border: '2px solid #b74115',
  },
  dropdown: {
    position: 'absolute',
    top: '100%',
    right: 0,
    width: '360px',
    maxHeight: '500px',
    background: '#fff',
    border: '1px solid #ddd',
    borderRadius: '8px',
    boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
    zIndex: 1000,
    overflow: 'hidden',
  },
  header: {
    padding: '16px',
    borderBottom: '1px solid #eee',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  headerTitle: {
    fontSize: '16px',
    fontWeight: '600',
    color: '#333',
  },
  clearButton: {
    background: 'none',
    border: 'none',
    color: '#1877f2',
    cursor: 'pointer',
    fontSize: '14px',
    fontWeight: '500',
  },
  content: {
    maxHeight: '400px',
    overflowY: 'auto',
  },
  requestItem: {
    padding: '16px',
    borderBottom: '1px solid #f0f0f0',
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
  },
  avatar: {
    width: '48px',
    height: '48px',
    borderRadius: '50%',
    backgroundColor: '#f0f0f0',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '18px',
    color: '#666',
    flexShrink: 0,
  },
  requestInfo: {
    flex: 1,
    minWidth: 0,
  },
  requesterName: {
    fontSize: '14px',
    fontWeight: '600',
    color: '#333',
    marginBottom: '2px',
  },
  requestText: {
    fontSize: '13px',
    color: '#666',
    marginBottom: '8px',
  },
  buttonGroup: {
    display: 'flex',
    gap: '8px',
  },
  acceptButton: {
    background: '#1877f2',
    color: 'white',
    border: 'none',
    padding: '6px 16px',
    borderRadius: '6px',
    fontSize: '13px',
    fontWeight: '600',
    cursor: 'pointer',
    transition: 'background-color 0.2s',
  },
  declineButton: {
    background: '#e4e6eb',
    color: '#050505',
    border: 'none',
    padding: '6px 16px',
    borderRadius: '6px',
    fontSize: '13px',
    fontWeight: '600',
    cursor: 'pointer',
    transition: 'background-color 0.2s',
  },
  loadingButton: {
    opacity: 0.6,
    cursor: 'not-allowed',
  },
  emptyState: {
    padding: '32px 16px',
    textAlign: 'center',
    color: '#666',
  },
  loadingState: {
    padding: '32px 16px',
    textAlign: 'center',
    color: '#666',
  },
  viewAllButton: {
    width: '100%',
    padding: '12px',
    background: '#f0f2f5',
    border: 'none',
    color: '#1877f2',
    fontSize: '14px',
    fontWeight: '600',
    cursor: 'pointer',
    transition: 'background-color 0.2s',
  },
};

export default function NotificationDropdown() {
  const { 
    followRequests, 
    loading: followRequestsLoading, 
    handleFollowRequestAction
  } = useFollowRequests();
  
  const {
    notifications,
    loading: notificationsLoading,
    markAsRead,
    unreadCount: notificationsUnreadCount
  } = useNotifications();
  
  // Calculate total unread count (follow requests + notifications)
  const followRequestsCount = Array.isArray(followRequests) ? followRequests.length : 0;
  const totalUnreadCount = followRequestsCount + notificationsUnreadCount;
  
  // Clear all follow requests (this would need backend support)
  const clearAll = () => {
    // For now, this is a placeholder. In a real app, you might want to
    // implement bulk accept/decline functionality
    console.log('Clear all functionality not implemented');
  };
  
  // Don't render if we're not in a browser environment
  if (typeof window === 'undefined') {
    return null;
  }
  
  const [isOpen, setIsOpen] = useState(false);
  const [processingRequests, setProcessingRequests] = useState(new Set());
  const dropdownRef = useRef(null);
  const router = useRouter();

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleRequestAction = async (requestId, action) => {
    try {
      setProcessingRequests(prev => new Set(prev).add(requestId));
      await handleFollowRequestAction(requestId, action);
    } catch (error) {
      console.error(`Failed to ${action} request:`, error);
    } finally {
      setProcessingRequests(prev => {
        const newSet = new Set(prev);
        newSet.delete(requestId);
        return newSet;
      });
    }
  };

  const handleNotificationClick = async (notification) => {
    if (!notification.read) {
      try {
        await markAsRead(notification.id);
      } catch (error) {
        console.error('Failed to mark notification as read:', error);
      }
    }
  };

  const getInitials = (firstName, lastName) => {
    return (firstName?.charAt(0) || '') + (lastName?.charAt(0) || '');
  };

  const formatTime = (dateString) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInHours = (now - date) / (1000 * 60 * 60);
    
    if (diffInHours < 1) {
      return 'Just now';
    } else if (diffInHours < 24) {
      const hours = Math.floor(diffInHours);
      return `${hours}h ago`;
    } else {
      const days = Math.floor(diffInHours / 24);
      return `${days}d ago`;
    }
  };

  const loading = followRequestsLoading || notificationsLoading;

  return (
    <div style={styles.container} ref={dropdownRef}>
      <button
        style={styles.notificationButton}
        onClick={() => setIsOpen(!isOpen)}
        onMouseEnter={(e) => e.target.style.backgroundColor = 'rgba(255,255,255,0.1)'}
        onMouseLeave={(e) => e.target.style.backgroundColor = 'transparent'}
      >
        <svg style={styles.bellIcon} viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 22c1.1 0 2-.9 2-2h-4c0 1.1.89 2 2 2zm6-6v-5c0-3.07-1.64-5.64-4.5-6.32V4c0-.83-.67-1.5-1.5-1.5s-1.5.67-1.5 1.5v.68C7.63 5.36 6 7.92 6 11v5l-2 2v1h16v-1l-2-2z"/>
        </svg>
        {totalUnreadCount > 0 && (
          <span style={styles.badge}>
            {totalUnreadCount > 99 ? '99+' : totalUnreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <div style={styles.dropdown}>
          <div style={styles.header}>
            <span style={styles.headerTitle}>
              Notifications ({totalUnreadCount})
            </span>
            {totalUnreadCount > 0 && (
              <button 
                style={styles.clearButton}
                onClick={clearAll}
              >
                Clear all
              </button>
            )}
          </div>

          <div style={styles.content}>
            {loading ? (
              <div style={styles.loadingState}>Loading...</div>
            ) : totalUnreadCount === 0 ? (
              <div style={styles.emptyState}>
                No new notifications
              </div>
            ) : (
              <>
                {/* Follow Requests Section */}
                {Array.isArray(followRequests) && followRequests.length > 0 && (
                  <>
                    {followRequests.length > 0 && (
                      <div style={{ padding: '12px 16px', fontSize: '14px', fontWeight: '600', color: '#666', borderBottom: '1px solid #f0f0f0' }}>
                        Follow Requests ({followRequests.length})
                      </div>
                    )}
                    {followRequests.slice(0, 3).map(request => {
                      const isProcessing = processingRequests.has(request.id);
                      const requester = request.requester;
                      
                      return (
                        <div key={request.id} style={styles.requestItem}>
                          <div style={styles.avatar}>
                            {requester.avatarPath ? (
                              <img 
                                src={requester.avatarPath} 
                                alt="Avatar" 
                                style={{ width: '100%', height: '100%', borderRadius: '50%', objectFit: 'cover' }}
                              />
                            ) : (
                              getInitials(requester.firstName, requester.lastName)
                            )}
                          </div>
                          
                          <div style={styles.requestInfo}>
                            <div style={styles.requesterName}>
                              {requester.firstName} {requester.lastName}
                            </div>
                            <div style={styles.requestText}>
                              Wants to follow you • {formatTime(request.createdAt)}
                            </div>
                            
                            <div style={styles.buttonGroup}>
                              <button
                                style={{
                                  ...styles.acceptButton,
                                  ...(isProcessing && styles.loadingButton)
                                }}
                                onClick={() => handleRequestAction(request.id, 'accept')}
                                disabled={isProcessing}
                              >
                                {isProcessing ? 'Accepting...' : 'Accept'}
                              </button>
                              
                              <button
                                style={{
                                  ...styles.declineButton,
                                  ...(isProcessing && styles.loadingButton)
                                }}
                                onClick={() => handleRequestAction(request.id, 'decline')}
                                disabled={isProcessing}
                              >
                                {isProcessing ? 'Declining...' : 'Decline'}
                              </button>
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </>
                )}

                {/* General Notifications Section */}
                {Array.isArray(notifications) && notifications.length > 0 && (
                  <>
                    {notifications.length > 0 && (
                      <div style={{ padding: '12px 16px', fontSize: '14px', fontWeight: '600', color: '#666', borderBottom: '1px solid #f0f0f0' }}>
                        Other Notifications ({notificationsUnreadCount})
                      </div>
                    )}
                    {notifications.slice(0, 3).map(notification => (
                      <div 
                        key={notification.id} 
                        style={{
                          ...styles.requestItem,
                          backgroundColor: notification.read ? 'transparent' : '#f8f9ff',
                          cursor: 'pointer'
                        }}
                        onClick={() => handleNotificationClick(notification)}
                      >
                        <div style={styles.avatar}>
                          {notification.actor?.avatarPath ? (
                            <img 
                              src={notification.actor.avatarPath} 
                              alt="Avatar" 
                              style={{ width: '100%', height: '100%', borderRadius: '50%', objectFit: 'cover' }}
                            />
                          ) : notification.actor ? (
                            getInitials(notification.actor.firstName, notification.actor.lastName)
                          ) : (
                            '📢'
                          )}
                        </div>
                        
                        <div style={styles.requestInfo}>
                          <div style={styles.requesterName}>
                            {notification.actor ? `${notification.actor.firstName} ${notification.actor.lastName}` : 'System'}
                          </div>
                          <div style={styles.requestText}>
                            {notification.message} • {formatTime(notification.createdAt)}
                          </div>
                        </div>
                      </div>
                    ))}
                  </>
                )}
              </>
            )}
          </div>

          {(followRequestsCount > 3 || notificationsUnreadCount > 3) && (
            <button 
              style={styles.viewAllButton}
              onClick={() => {
                setIsOpen(false);
                router.push('/notifications');
              }}
            >
              View all notifications
            </button>
          )}
        </div>
      )}
    </div>
  );
} 