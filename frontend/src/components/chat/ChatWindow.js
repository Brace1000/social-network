'use client';

import React, { useState, useEffect, useRef } from 'react';
import MessageBubble from './MessageBubble';
import EmojiPicker from './EmojiPicker';
import { useAuth } from '../../store/AuthContext';
import { useChat } from '../../store/ChatContext';

const ChatWindow = ({ conversation }) => {
  const [messageText, setMessageText] = useState('');
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [loading, setLoading] = useState(false);
  const messagesEndRef = useRef(null);
  const inputRef = useRef(null);
  const { user } = useAuth();
  const { messages, loadMessages, sendMessage } = useChat();

  const conversationKey = conversation 
    ? `${conversation.type}_${conversation.type === 'private' ? conversation.userId : conversation.groupId}`
    : null;
  
  const conversationMessages = conversationKey ? messages[conversationKey] || [] : [];

  // Load messages when conversation changes
  useEffect(() => {
    if (conversation) {
      const conversationId = conversation.type === 'private' ? conversation.userId : conversation.groupId;
      setLoading(true);
      loadMessages(conversationId, conversation.type).finally(() => {
        setLoading(false);
      });
    }
  }, [conversation]); // Removed loadMessages from dependencies to prevent infinite loop

  // Scroll to bottom when new messages arrive
  useEffect(() => {
    scrollToBottom();
  }, [conversationMessages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSendMessage = (e) => {
    e.preventDefault();
    
    if (!messageText.trim() || !conversation) return;

    const recipientId = conversation.type === 'private' ? conversation.userId : null;
    const groupId = conversation.type === 'group' ? conversation.groupId : null;

    const success = sendMessage(messageText.trim(), recipientId, groupId);
    
    if (success) {
      setMessageText('');
      setShowEmojiPicker(false);
    }
  };

  const handleEmojiSelect = (emoji) => {
    setMessageText(prev => prev + emoji);
    setShowEmojiPicker(false);
    inputRef.current?.focus();
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage(e);
    }
  };

  const getSenderName = (message) => {
    if (message.senderId === user?.id) return null;
    // For group messages, we'd need to fetch user info
    // For now, just show the sender ID
    return message.senderId;
  };

  const styles = {
    container: {
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
      backgroundColor: 'white',
    },
    header: {
      padding: '16px 20px',
      borderBottom: '1px solid #e1e5e9',
      backgroundColor: '#f8f9fa',
      display: 'flex',
      alignItems: 'center',
      gap: '12px',
    },
    avatar: {
      width: '40px',
      height: '40px',
      borderRadius: '50%',
      backgroundColor: '#007bff',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      color: 'white',
      fontSize: '16px',
      fontWeight: '600',
    },
    headerInfo: {
      flex: 1,
    },
    conversationName: {
      fontSize: '16px',
      fontWeight: '600',
      color: '#333',
      margin: 0,
    },
    conversationStatus: {
      fontSize: '12px',
      color: '#666',
      marginTop: '2px',
    },
    messagesContainer: {
      flex: 1,
      overflowY: 'auto',
      padding: '20px',
      display: 'flex',
      flexDirection: 'column',
      gap: '8px',
    },
    loadingState: {
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      height: '100%',
      color: '#666',
    },
    emptyState: {
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      height: '100%',
      color: '#666',
      textAlign: 'center',
    },
    inputContainer: {
      padding: '16px 20px',
      borderTop: '1px solid #e1e5e9',
      backgroundColor: 'white',
      position: 'relative',
    },
    inputForm: {
      display: 'flex',
      alignItems: 'flex-end',
      gap: '12px',
    },
    inputWrapper: {
      flex: 1,
      position: 'relative',
    },
    messageInput: {
      width: '100%',
      minHeight: '40px',
      maxHeight: '120px',
      padding: '10px 40px 10px 12px',
      border: '1px solid #ddd',
      borderRadius: '20px',
      fontSize: '14px',
      resize: 'none',
      outline: 'none',
      fontFamily: 'inherit',
    },
    emojiButton: {
      position: 'absolute',
      right: '8px',
      top: '50%',
      transform: 'translateY(-50%)',
      background: 'none',
      border: 'none',
      fontSize: '18px',
      cursor: 'pointer',
      padding: '4px',
      borderRadius: '50%',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    },
    sendButton: {
      backgroundColor: '#007bff',
      color: 'white',
      border: 'none',
      borderRadius: '50%',
      width: '40px',
      height: '40px',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: '16px',
      transition: 'background-color 0.2s',
    },
    sendButtonDisabled: {
      backgroundColor: '#ccc',
      cursor: 'not-allowed',
    },
    emojiPickerContainer: {
      position: 'absolute',
      bottom: '60px',
      right: '20px',
      zIndex: 1000,
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

  if (!conversation) {
    return (
      <div style={styles.container}>
        <div style={styles.emptyState}>
          <div>
            <h3>Select a conversation</h3>
            <p>Choose a conversation from the list to start messaging</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      {/* Header */}
      <div style={styles.header}>
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
        <div style={styles.headerInfo}>
          <h3 style={styles.conversationName}>{conversation.name}</h3>
          <div style={styles.conversationStatus}>
            {conversation.type === 'group' ? 'Group Chat' : 'Private Chat'}
          </div>
        </div>
      </div>

      {/* Messages */}
      <div style={styles.messagesContainer}>
        {loading ? (
          <div style={styles.loadingState}>Loading messages...</div>
        ) : conversationMessages.length === 0 ? (
          <div style={styles.emptyState}>
            <div>
              <p>No messages yet</p>
              <p style={{ fontSize: '14px', color: '#999' }}>
                Start the conversation by sending a message
              </p>
            </div>
          </div>
        ) : (
          conversationMessages.map((message) => (
            <MessageBubble
              key={message.id}
              message={message}
              isOwn={message.senderId === user?.id}
              senderName={getSenderName(message)}
            />
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div style={styles.inputContainer}>
        {showEmojiPicker && (
          <div style={styles.emojiPickerContainer}>
            <EmojiPicker onEmojiSelect={handleEmojiSelect} onClose={() => setShowEmojiPicker(false)} />
          </div>
        )}
        
        <form style={styles.inputForm} onSubmit={handleSendMessage}>
          <div style={styles.inputWrapper}>
            <textarea
              ref={inputRef}
              style={styles.messageInput}
              value={messageText}
              onChange={(e) => setMessageText(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="Type a message..."
              rows={1}
            />
            <button
              type="button"
              style={styles.emojiButton}
              onClick={() => setShowEmojiPicker(!showEmojiPicker)}
              title="Add emoji"
            >
              😊
            </button>
          </div>
          <button
            type="submit"
            style={{
              ...styles.sendButton,
              ...(messageText.trim() ? {} : styles.sendButtonDisabled),
            }}
            disabled={!messageText.trim()}
            title="Send message"
          >
            ➤
          </button>
        </form>
      </div>
    </div>
  );
};

export default ChatWindow;
