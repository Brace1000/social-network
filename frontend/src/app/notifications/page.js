'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '../../store/AuthContext';
import { useFollowRequests } from '../../lib/hooks';

const styles = {
  container: {
    maxWidth: 800,
    margin: '0 auto',
    padding: '2rem',
    fontFamily: 'Arial, sans-serif'
  },
  header: {
    fontSize: '2rem',
    marginBottom: '2rem',
    color: '#333'
  },
  section: {
    marginBottom: '3rem'
  },
  sectionTitle: {
    fontSize: '1.5rem',
    marginBottom: '1rem',
    color: '#555'
  },
  requestCard: {
    border: '1px solid #ddd',
    borderRadius: 8,
    padding: '1rem',
    marginBottom: '1rem',
    backgroundColor: '#fff',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
  },
  requesterInfo: {
    display: 'flex',
    alignItems: 'center',
    marginBottom: '1rem'
  },
  avatar: {
    width: 50,
    height: 50,
    borderRadius: '50%',
    marginRight: '1rem',
    backgroundColor: '#f0f0f0',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '1.2rem',
    color: '#666'
  },
  requesterName: {
    fontSize: '1.1rem',
    fontWeight: 'bold',
    color: '#333'
  },
  requesterNickname: {
    color: '#666',
    fontSize: '0.9rem'
  },
  requestTime: {
    color: '#888',
    fontSize: '0.8rem',
    marginBottom: '1rem'
  },
  buttonGroup: {
    display: 'flex',
    gap: '0.5rem'
  },
  acceptButton: {
    backgroundColor: '#4CAF50',
    color: 'white',
    border: 'none',
    padding: '0.5rem 1rem',
    borderRadius: 4,
    cursor: 'pointer',
    fontSize: '0.9rem'
  },
  declineButton: {
    backgroundColor: '#f44336',
    color: 'white',
    border: 'none',
    padding: '0.5rem 1rem',
    borderRadius: 4,
    cursor: 'pointer',
    fontSize: '0.9rem'
  },
  loadingButton: {
    opacity: 0.6,
    cursor: 'not-allowed'
  },
  emptyState: {
    textAlign: 'center',
    color: '#888',
    padding: '2rem'
  },
  error: {
    color: '#f44336',
    textAlign: 'center',
    padding: '1rem'
  }
};

export default function NotificationsPage() {
  const { user, isAuthenticated, loading: authLoading } = useAuth();
  const { 
    followRequests, 
    loading, 
    error, 
    handleFollowRequestAction 
  } = useFollowRequests();
  const [processingRequests, setProcessingRequests] = useState(new Set());

  const handleFollowRequest = async (requestId, action) => {
    try {
      setProcessingRequests(prev => new Set(prev).add(requestId));
      await handleFollowRequestAction(requestId, action);
    } catch (err) {
      console.error(`Error ${action}ing follow request:`, err);
    } finally {
      setProcessingRequests(prev => {
        const newSet = new Set(prev);
        newSet.delete(requestId);
        return newSet;
      });
    }
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const getInitials = (firstName, lastName) => {
    return (firstName?.charAt(0) || '') + (lastName?.charAt(0) || '');
  };

  // Show loading while checking authentication
  if (authLoading) {
    return (
      <div style={styles.container}>
        <div style={styles.loadingState}>Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div style={styles.container}>
        <div style={styles.error}>Please log in to view notifications.</div>
      </div>
    );
  }

  if (loading) {
    return (
      <div style={styles.container}>
        <div style={styles.emptyState}>Loading notifications...</div>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      <h1 style={styles.header}>Notifications</h1>
      
      {error && <div style={styles.error}>{error}</div>}
      
      <div style={styles.section}>
        <h2 style={styles.sectionTitle}>Follow Requests ({followRequests.length})</h2>
        
        {followRequests.length === 0 ? (
          <div style={styles.emptyState}>
            {error === 'No follow requests found' ? 'No follow requests found.' : 'No pending follow requests.'}
          </div>
        ) : (
          followRequests.map(request => {
            const isProcessing = processingRequests.has(request.id);
            const requester = request.requester;
            
            return (
              <div key={request.id} style={styles.requestCard}>
                <div style={styles.requesterInfo}>
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
                  <div>
                    <div style={styles.requesterName}>
                      {requester.firstName} {requester.lastName}
                    </div>
                    {requester.nickname && (
                      <div style={styles.requesterNickname}>
                        @{requester.nickname}
                      </div>
                    )}
                  </div>
                </div>
                
                <div style={styles.requestTime}>
                  Requested on {formatDate(request.createdAt)}
                </div>
                
                <div style={styles.buttonGroup}>
                  <button
                    style={{
                      ...styles.acceptButton,
                      ...(isProcessing && styles.loadingButton)
                    }}
                    onClick={() => handleFollowRequest(request.id, 'accept')}
                    disabled={isProcessing}
                  >
                    {isProcessing ? 'Accepting...' : 'Accept'}
                  </button>
                  
                  <button
                    style={{
                      ...styles.declineButton,
                      ...(isProcessing && styles.loadingButton)
                    }}
                    onClick={() => handleFollowRequest(request.id, 'decline')}
                    disabled={isProcessing}
                  >
                    {isProcessing ? 'Declining...' : 'Decline'}
                  </button>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}

