'use client';

import React from 'react';

const ConversationList = ({ conversations, activeConversation, onSelectConversation, loading }) => {
  const formatTime = (timestamp) => {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    const now = new Date();
    const diffInHours = (now - date) / (1000 * 60 * 60);
    
    if (diffInHours < 24) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
    }
  };

  const truncateMessage = (message, maxLength = 50) => {
    if (!message) return 'No messages yet';
    return message.length > maxLength ? message.substring(0, maxLength) + '...' : message;
  };

  const getConversationId = (conversation) => {
    return conversation.type === 'private' ? conversation.userId : conversation.groupId;
  };

  const styles = {
    container: {
      height: '100%',
      backgroundColor: 'white',
      borderRight: '1px solid #e1e5e9',
      display: 'flex',
      flexDirection: 'column',
    },
    header: {
      padding: '20px',
      borderBottom: '1px solid #e1e5e9',
      backgroundColor: '#f8f9fa',
    },
    title: {
      fontSize: '18px',
      fontWeight: '600',
      color: '#333',
      margin: 0,
    },
    conversationsList: {
      flex: 1,
      overflowY: 'auto',
    },
    conversationItem: {
      padding: '16px 20px',
      borderBottom: '1px solid #f0f0f0',
      cursor: 'pointer',
      transition: 'background-color 0.2s',
      display: 'flex',
      alignItems: 'center',
      gap: '12px',
    },
    conversationItemActive: {
      backgroundColor: '#e3f2fd',
      borderRight: '3px solid #007bff',
    },
    conversationItemHover: {
      backgroundColor: '#f5f5f5',
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
      fontSize: '18px',
      fontWeight: '600',
      flexShrink: 0,
    },
    conversationInfo: {
      flex: 1,
      minWidth: 0,
    },
    conversationName: {
      fontSize: '14px',
      fontWeight: '600',
      color: '#333',
      marginBottom: '4px',
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      whiteSpace: 'nowrap',
    },
    lastMessage: {
      fontSize: '13px',
      color: '#666',
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      whiteSpace: 'nowrap',
    },
    conversationMeta: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'flex-end',
      gap: '4px',
    },
    timestamp: {
      fontSize: '12px',
      color: '#999',
    },
    unreadBadge: {
      backgroundColor: '#007bff',
      color: 'white',
      borderRadius: '10px',
      padding: '2px 6px',
      fontSize: '11px',
      fontWeight: '600',
      minWidth: '18px',
      textAlign: 'center',
    },
    emptyState: {
      padding: '40px 20px',
      textAlign: 'center',
      color: '#666',
    },
    loadingState: {
      padding: '20px',
      textAlign: 'center',
      color: '#666',
    },
  };

  const getInitials = (name) => {
    return name
      .split(' ')
      .map(word => word[0])
      .join('')
      .toUpperCase()
      .substring(0, 2);
  };

  if (loading) {
    return (
      <div style={styles.container}>
        <div style={styles.header}>
          <h2 style={styles.title}>Messages</h2>
        </div>
        <div style={styles.loadingState}>Loading conversations...</div>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h2 style={styles.title}>Messages</h2>
      </div>
      
      <div style={styles.conversationsList}>
        {conversations.length === 0 ? (
          <div style={styles.emptyState}>
            <p>No conversations yet</p>
            <p style={{ fontSize: '12px', marginTop: '8px' }}>
              Start a conversation by searching for users
            </p>
          </div>
        ) : (
          conversations.map((conversation) => {
            const conversationId = getConversationId(conversation);
            const isActive = activeConversation && 
              ((activeConversation.type === conversation.type) &&
               (getConversationId(activeConversation) === conversationId));

            return (
              <div
                key={`${conversation.type}_${conversationId}`}
                style={{
                  ...styles.conversationItem,
                  ...(isActive ? styles.conversationItemActive : {}),
                }}
                onClick={() => onSelectConversation(conversation)}
                onMouseEnter={(e) => {
                  if (!isActive) {
                    e.target.style.backgroundColor = styles.conversationItemHover.backgroundColor;
                  }
                }}
                onMouseLeave={(e) => {
                  if (!isActive) {
                    e.target.style.backgroundColor = 'transparent';
                  }
                }}
              >
                <div style={styles.avatar}>
                  {conversation.avatarPath ? (
                    <img 
                      src={conversation.avatarPath} 
                      alt={conversation.name}
                      style={{ width: '100%', height: '100%', borderRadius: '50%' }}
                    />
                  ) : (
                    getInitials(conversation.name)
                  )}
                </div>
                
                <div style={styles.conversationInfo}>
                  <div style={styles.conversationName}>
                    {conversation.name}
                    {conversation.type === 'group' && (
                      <span style={{ fontSize: '12px', color: '#666', marginLeft: '4px' }}>
                        (Group)
                      </span>
                    )}
                  </div>
                  <div style={styles.lastMessage}>
                    {truncateMessage(conversation.lastMessage)}
                  </div>
                </div>
                
                <div style={styles.conversationMeta}>
                  <div style={styles.timestamp}>
                    {formatTime(conversation.lastMessageTime)}
                  </div>
                  {conversation.unreadCount > 0 && (
                    <div style={styles.unreadBadge}>
                      {conversation.unreadCount}
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};

export default ConversationList;
