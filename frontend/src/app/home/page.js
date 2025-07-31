'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import LogoutForm from "../../components/auth/LogoutForm";
import NotificationDropdown from "../../components/NotificationDropdown";
import EditProfile from "../../components/EditProfile";
import { useWebSocketConnection } from '../../lib/websocket';
import { useAuth } from '../../store/AuthContext';

// Constants
const API_BASE_URL = 'http://localhost:8080/api/v1';
const NAVBAR_HEIGHT = 80;

// Styles
const styles = {
  container: {
    display: 'flex',
    flexDirection: 'column',
    minHeight: '100vh'
  },
  navbar: {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    zIndex: 1000,
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '16px 32px',
    background: '#b74115',
    borderBottom: '1px solid #a03a12',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
  },
  navbarTitle: {
    fontSize: '20px',
    fontWeight: 'bold',
    color: '#fff'
  },
  mainContent: {
    display: 'flex',
    flex: 1,
    marginTop: `${NAVBAR_HEIGHT}px`
  },
  sidebar: {
    flex: '0 0 280px',
    background: '#fafbfc',
    borderRight: '1px solid #eee',
    minHeight: '100vh',
    padding: '20px'
  },
  avatar: {
    width: 80,
    height: 80,
    borderRadius: '50%',
    objectFit: 'cover',
    border: '2px solid #eee'
  },
  loadingContainer: {
    textAlign: 'center',
    marginTop: 40
  },
  errorContainer: {
    textAlign: 'center',
    marginTop: 40,
    color: 'red'
  }
};

// Custom hooks
const useProfile = () => {
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { user } = useAuth();

  const fetchProfile = useCallback(async () => {
    if (!user) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError('');
    
    try {
      const profileRes = await fetch(`${API_BASE_URL}/profile/${user.id}`, { credentials: 'include' });
      
      if (!profileRes.ok) {
        setError('Failed to fetch profile.');
        setProfile(null);
      } else {
        const data = await profileRes.json();
        setProfile(data);
      }
    } catch (error) {
      console.error('Profile fetch error:', error);
      setError('Failed to fetch profile.');
      setProfile(null);
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => {
    fetchProfile();
  }, [fetchProfile]);

  return { profile, loading, error };
};

// Components
const LoadingSpinner = () => (
  <div style={styles.loadingContainer}>
    <div className="content-placeholder" style={{ 
      width: '80px', 
      height: '80px', 
      borderRadius: '50%',
      margin: '0 auto'
    }}></div>
  </div>
);

const ErrorMessage = ({ message }) => (
  <div style={styles.errorContainer}>{message}</div>
);

const ProfileButton = ({ profile, onNavigate }) => {
  const [isClicked, setIsClicked] = useState(false);
  const router = useRouter();

  const handleProfileClick = (e) => {
    e.preventDefault();
    setIsClicked(true);
    
    // Add a small delay to show the click effect before navigation
    setTimeout(() => {
      router.push(`/profile/${profile.id}`);
      onNavigate(); // Notify parent that navigation is starting
    }, 150);
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '20px' }}>
      <a 
        href={`/profile/${profile.id}`}
        onClick={handleProfileClick}
        style={{ textDecoration: 'none' }}
      >
        <img 
          src={profile.avatarPath ? `/${profile.avatarPath}` : '/user.png'} 
          alt="avatar" 
          className={`profile-image ${isClicked ? 'profile-clicked' : ''}`}
          style={{ 
            ...styles.avatar, 
            cursor: 'pointer',
            transform: isClicked ? 'scale(0.95)' : 'scale(1)',
            transition: 'transform 0.15s ease-in-out'
          }}
          title="Click to view profile"
        />
      </a>
    </div>
  );
};

// Main component
export default function HomePage() {
  const { user, isAuthenticated, loading: authLoading } = useAuth();
  const { profile, loading, error } = useProfile();
  const [isNavigating, setIsNavigating] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const router = useRouter();

  // Always call hooks at the top level!
  useEffect(() => {
    if (!isAuthenticated && !authLoading) {
      router.push('/auth/login');
    }
  }, [isAuthenticated, authLoading, router]);

  // Show loading while checking authentication
  if (authLoading) return (
    <div className="page-enter" style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      minHeight: '100vh',
      background: '#f0f2f5'
    }}>
      <div className="content-placeholder" style={{ 
        width: '60px', 
        height: '60px', 
        borderRadius: '50%'
      }}></div>
    </div>
  );
  
  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    return (
      <div className="page-enter" style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        minHeight: '100vh',
        background: '#f0f2f5'
      }}>
        <div style={{ fontSize: '18px', color: '#666' }}>Redirecting to login...</div>
      </div>
    );
  }

  if (loading) return (
    <div className="page-enter" style={styles.container}>
      <nav style={styles.navbar}>
        <div style={styles.navbarTitle}>
          Social Network
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <Link
            href="/messages"
            style={{
              color: '#fff',
              textDecoration: 'none',
              padding: '8px 16px',
              borderRadius: '6px',
              backgroundColor: 'rgba(255,255,255,0.1)',
              transition: 'background-color 0.2s',
              fontSize: '14px',
              fontWeight: '500'
            }}
            onMouseEnter={(e) => e.target.style.backgroundColor = 'rgba(255,255,255,0.2)'}
            onMouseLeave={(e) => e.target.style.backgroundColor = 'rgba(255,255,255,0.1)'}
          >
            Messages
          </Link>
          <NotificationDropdown />
          <LogoutForm />
        </div>
      </nav>
      <div style={styles.mainContent}>
        <div style={styles.sidebar}>
          <LoadingSpinner />
        </div>
      </div>
    </div>
  );
  
  if (error) return (
    <div className="page-enter" style={styles.container}>
      <nav style={styles.navbar}>
        <div style={styles.navbarTitle}>
          Social Network
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <NotificationDropdown />
          <LogoutForm />
        </div>
      </nav>
      <div style={styles.mainContent}>
        <div style={styles.sidebar}>
          <ErrorMessage message={error} />
        </div>
      </div>
    </div>
  );
  
  if (!profile) return (
    <div className="page-enter" style={styles.container}>
      <nav style={styles.navbar}>
        <div style={styles.navbarTitle}>
          Social Network
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <Link
            href="/messages"
            style={{
              color: '#fff',
              textDecoration: 'none',
              padding: '8px 16px',
              borderRadius: '6px',
              backgroundColor: 'rgba(255,255,255,0.1)',
              transition: 'background-color 0.2s',
              fontSize: '14px',
              fontWeight: '500'
            }}
            onMouseEnter={(e) => e.target.style.backgroundColor = 'rgba(255,255,255,0.2)'}
            onMouseLeave={(e) => e.target.style.backgroundColor = 'rgba(255,255,255,0.1)'}
          >
            Messages
          </Link>
          <NotificationDropdown />
          <LogoutForm />
        </div>
      </nav>
      <div style={styles.mainContent}>
        <div style={styles.sidebar}>
          <LoadingSpinner />
        </div>
      </div>
    </div>
  );

  return (
    <div className="page-enter" style={styles.container}>
      {/* Navigation overlay */}
      {isNavigating && (
        <div 
          className="nav-preload active"
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: '#f0f2f5',
            zIndex: 9999,
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center'
          }}
        >
          <div className="content-placeholder" style={{ 
            width: '80px', 
            height: '80px', 
            borderRadius: '50%'
          }}></div>
        </div>
      )}
      
      {/* Navbar */}
      <nav style={styles.navbar}>
        <div style={styles.navbarTitle}>
          Social Network
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <Link
            href="/messages"
            style={{
              color: '#fff',
              textDecoration: 'none',
              padding: '8px 16px',
              borderRadius: '6px',
              backgroundColor: 'rgba(255,255,255,0.1)',
              transition: 'background-color 0.2s',
              fontSize: '14px',
              fontWeight: '500'
            }}
            onMouseEnter={(e) => e.target.style.backgroundColor = 'rgba(255,255,255,0.2)'}
            onMouseLeave={(e) => e.target.style.backgroundColor = 'rgba(255,255,255,0.1)'}
          >
            Messages
          </Link>
          <NotificationDropdown />
          <LogoutForm />
        </div>
      </nav>
      
      {/* Main Content */}
      <div style={styles.mainContent}>
        {/* Profile Section */}
        <div style={styles.sidebar}>
          <div style={{ marginBottom: '20px' }}>
            <ProfileButton profile={profile} onNavigate={() => setIsNavigating(true)} />
          </div>

        </div>
      </div>

      {/* Edit Profile Modal */}
      <EditProfile 
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        onProfileUpdated={() => {
          // Refresh the profile data after successful update
          window.location.reload();
        }}
      />
    </div>
  );
}
