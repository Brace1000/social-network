'use client';

import React, { useRef, useEffect } from 'react';

const EmojiPicker = ({ onEmojiSelect, onClose }) => {
  const pickerRef = useRef(null);

  // Common emojis organized by category
  const emojiCategories = {
    'Smileys': [
      '😀', '😃', '😄', '😁', '😆', '😅', '😂', '🤣', '😊', '😇',
      '🙂', '🙃', '😉', '😌', '😍', '🥰', '😘', '😗', '😙', '😚',
      '😋', '😛', '😝', '😜', '🤪', '🤨', '🧐', '🤓', '😎', '🤩',
      '🥳', '😏', '😒', '😞', '😔', '😟', '😕', '🙁', '☹️', '😣',
      '😖', '😫', '😩', '🥺', '😢', '😭', '😤', '😠', '😡', '🤬'
    ],
    'Gestures': [
      '👍', '👎', '👌', '🤌', '🤏', '✌️', '🤞', '🤟', '🤘', '🤙',
      '👈', '👉', '👆', '🖕', '👇', '☝️', '👋', '🤚', '🖐️', '✋',
      '🖖', '👏', '🙌', '🤲', '🤝', '🙏', '✍️', '💪', '🦾', '🦿'
    ],
    'Hearts': [
      '❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔',
      '❣️', '💕', '💞', '💓', '💗', '💖', '💘', '💝', '💟', '♥️'
    ],
    'Objects': [
      '🎉', '🎊', '🎈', '🎁', '🎀', '🎂', '🍰', '🧁', '🍕', '🍔',
      '🌮', '🍟', '🍿', '🍩', '🍪', '🍫', '🍬', '🍭', '🍯', '🍼',
      '☕', '🍵', '🧃', '🥤', '🍺', '🍻', '🥂', '🍷', '🥃', '🍸'
    ]
  };

  // Handle clicks outside the picker
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (pickerRef.current && !pickerRef.current.contains(event.target)) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [onClose]);

  const styles = {
    container: {
      backgroundColor: 'white',
      border: '1px solid #ddd',
      borderRadius: '12px',
      boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
      width: '320px',
      maxHeight: '300px',
      overflow: 'hidden',
      zIndex: 1000,
    },
    header: {
      padding: '12px 16px',
      borderBottom: '1px solid #eee',
      backgroundColor: '#f8f9fa',
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
    },
    title: {
      fontSize: '14px',
      fontWeight: '600',
      color: '#333',
      margin: 0,
    },
    closeButton: {
      background: 'none',
      border: 'none',
      fontSize: '18px',
      cursor: 'pointer',
      color: '#666',
      padding: '0',
      width: '24px',
      height: '24px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      borderRadius: '50%',
    },
    content: {
      maxHeight: '240px',
      overflowY: 'auto',
      padding: '8px',
    },
    category: {
      marginBottom: '16px',
    },
    categoryTitle: {
      fontSize: '12px',
      fontWeight: '600',
      color: '#666',
      marginBottom: '8px',
      paddingLeft: '4px',
    },
    emojiGrid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(8, 1fr)',
      gap: '4px',
    },
    emojiButton: {
      background: 'none',
      border: 'none',
      fontSize: '20px',
      cursor: 'pointer',
      padding: '6px',
      borderRadius: '6px',
      transition: 'background-color 0.2s',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      aspectRatio: '1',
    },
    emojiButtonHover: {
      backgroundColor: '#f0f0f0',
    },
  };

  return (
    <div ref={pickerRef} style={styles.container}>
      <div style={styles.header}>
        <h4 style={styles.title}>Choose an emoji</h4>
        <button
          style={styles.closeButton}
          onClick={onClose}
          onMouseEnter={(e) => {
            e.target.style.backgroundColor = '#f0f0f0';
          }}
          onMouseLeave={(e) => {
            e.target.style.backgroundColor = 'transparent';
          }}
        >
          ×
        </button>
      </div>
      
      <div style={styles.content}>
        {Object.entries(emojiCategories).map(([categoryName, emojis]) => (
          <div key={categoryName} style={styles.category}>
            <div style={styles.categoryTitle}>{categoryName}</div>
            <div style={styles.emojiGrid}>
              {emojis.map((emoji, index) => (
                <button
                  key={`${categoryName}-${index}`}
                  style={styles.emojiButton}
                  onClick={() => onEmojiSelect(emoji)}
                  onMouseEnter={(e) => {
                    e.target.style.backgroundColor = styles.emojiButtonHover.backgroundColor;
                  }}
                  onMouseLeave={(e) => {
                    e.target.style.backgroundColor = 'transparent';
                  }}
                  title={emoji}
                >
                  {emoji}
                </button>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default EmojiPicker;
