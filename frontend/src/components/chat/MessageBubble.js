'use client';

import React from 'react';

const MessageBubble = ({ message, isOwn, senderName }) => {
  const formatTime = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const styles = {
    container: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: isOwn ? 'flex-end' : 'flex-start',
      marginBottom: '12px',
      maxWidth: '70%',
      alignSelf: isOwn ? 'flex-end' : 'flex-start',
    },
    senderName: {
      fontSize: '12px',
      color: '#666',
      marginBottom: '4px',
      paddingLeft: isOwn ? '0' : '12px',
      paddingRight: isOwn ? '12px' : '0',
    },
    bubble: {
      backgroundColor: isOwn ? '#007bff' : '#f1f3f4',
      color: isOwn ? 'white' : '#333',
      padding: '12px 16px',
      borderRadius: '18px',
      borderTopRightRadius: isOwn ? '6px' : '18px',
      borderTopLeftRadius: isOwn ? '18px' : '6px',
      wordWrap: 'break-word',
      maxWidth: '100%',
      position: 'relative',
    },
    content: {
      fontSize: '14px',
      lineHeight: '1.4',
      margin: 0,
    },
    timestamp: {
      fontSize: '11px',
      color: isOwn ? 'rgba(255,255,255,0.7)' : '#999',
      marginTop: '4px',
      textAlign: isOwn ? 'right' : 'left',
    },
  };

  return (
    <div style={styles.container}>
      {!isOwn && senderName && (
        <div style={styles.senderName}>{senderName}</div>
      )}
      <div style={styles.bubble}>
        <div style={styles.content}>{message.content}</div>
        <div style={styles.timestamp}>
          {formatTime(message.createdAt)}
        </div>
      </div>
    </div>
  );
};

export default MessageBubble;
