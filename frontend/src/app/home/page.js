'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import PostList from "../../components/posts/PostList";
import PostForm from "../../components/posts/PostForm";
import LogoutForm from "../../components/auth/LogoutForm";
import NotificationDropdown from "../../components/NotificationDropdown";
import EditProfile from "../../components/EditProfile";
import { useAuth } from '../../store/AuthContext';

// Constants
const API_BASE_URL = 'http://localhost:8080/api/v1';
const NAVBAR_HEIGHT = 80;

// Styles
const styles = {
  container: { display: 'flex', flexDirection: 'column', minHeight: '100vh' },
  navbar: {
    position: 'fixed', top: 0, left: 0, right: 0, zIndex: 1000,
    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
    padding: '16px 32px', background: '#b74115', borderBottom: '1px solid #a03a12',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
  },
  navbarTitle: { fontSize: '20px', fontWeight: 'bold', color: '#fff' },
  mainContent: { display: 'flex', flex: 1, marginTop: `${NAVBAR_HEIGHT}px` },
  sidebar: { flex: '0 0 280px', background: '#fafbfc', borderRight: '1px solid #eee', minHeight: '100vh', padding: '20px' },
  avatar: { width: 80, height: 80, borderRadius: '50%', objectFit: 'cover', border: '2px solid #eee' },
  loadingContainer: { textAlign: 'center', marginTop: 40 },
  errorContainer: { textAlign: 'center', marginTop: 40, color: 'red' }
};

// --- Custom hooks ---
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
      const res = await fetch(`${API_BASE_URL}/profile/${user.id}`, { credentials: 'include' });
      if (!res.ok) {
        setError('Failed to fetch profile.');
        setProfile(null);
      } else {
        const data = await res.json();
        setProfile(data);
      }
    } catch (err) {
      console.error('Profile fetch error:', err);
      setError('Failed to fetch profile.');
      setProfile(null);
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => { fetchProfile(); }, [fetchProfile]);

  return { profile, loading, error };
};

// --- Components ---
const LoadingSpinner = () => (
  <div style={styles.loadingContainer}>
    <div className="content-placeholder" style={{ width: 80, height: 80, borderRadius: '50%', margin: '0 auto' }}></div>
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
    setTimeout(() => {
      router.push(`/profile/${profile.id}`);
      onNavigate();
    }, 150);
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '20px' }}>
      <a href={`/profile/${profile.id}`} onClick={handleProfileClick} style={{ textDecoration: 'none' }}>
        <img 
          src={profile.avatarPath ? `/${profile.avatarPath}` : '/user.png'} 
          alt="avatar"
          style={{ ...styles.avatar, cursor: 'pointer', transform: isClicked ? 'scale(0.95)' : 'scale(1)', transition: 'transform 0.15s ease-in-out' }}
        />
      </a>
    </div>
  );
};

// --- Main HomePage ---
export default function HomePage() {
  const { user, isAuthenticated, loading: authLoading } = useAuth();
  const { profile, loading, error } = useProfile();
  const [isNavigating, setIsNavigating] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const router = useRouter();

  // NEW state for feed
  const [showForm, setShowForm] = useState(false);
  const [postsKey, setPostsKey] = useState(0);
  function refreshPosts() {
    setPostsKey((prev) => prev + 1);
    setShowForm(false);
  }

  // Redirect if not logged in
  useEffect(() => {
    if (!isAuthenticated && !authLoading) {
      router.push('/auth/login');
    }
  }, [isAuthenticated, authLoading, router]);

  // --- Conditional states ---
  if (authLoading) return <LoadingSpinner />;
  if (!isAuthenticated) return <div>Redirecting to login...</div>;
  if (loading) return <LoadingSpinner />;
  if (error) return <ErrorMessage message={error} />;
  if (!profile) return <LoadingSpinner />;

  // --- Render main page ---
  return (
    <div className="page-enter" style={styles.container}>
      {/* Navbar */}
      <nav style={styles.navbar}>
        <div style={styles.navbarTitle}>Social Network</div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <Link href="/messages" style={{ color: '#fff', textDecoration: 'none' }}>Messages</Link>
          <NotificationDropdown />
          <LogoutForm />
        </div>
      </nav>

      {/* Main Content */}
      <div style={styles.mainContent}>
        {/* Sidebar */}
        <div style={styles.sidebar}>
          <ProfileButton profile={profile} onNavigate={() => setIsNavigating(true)} />
          {/* You can add Groups list here later */}
        </div>

        {/* Feed Section */}
        <div style={{ flex: 1, padding: "20px" }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <h2>Feed</h2>
            <button
              onClick={() => setShowForm((prev) => !prev)}
              style={{ background: "#b74115", color: "white", padding: "8px 16px", borderRadius: "8px", border: "none", cursor: "pointer" }}
            >
              {showForm ? "Cancel" : "+ New Post"}
            </button>
          </div>

          {showForm && <PostForm onSuccess={refreshPosts} />}
          <PostList key={postsKey} userId={profile.id} />
        </div>
      </div>

      {/* Edit Profile Modal */}
      <EditProfile 
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        onProfileUpdated={() => window.location.reload()}
      />
    </div>
  );
}
