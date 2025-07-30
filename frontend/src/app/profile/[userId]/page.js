'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { followAPI, userAPI } from '../../../lib/api';
import { useAuth } from '../../../store/AuthContext';
import { useMyFollowRequests } from '../../../lib/hooks';
import EditProfile from '../../../components/EditProfile';

export default function ProfilePage() {
  const { user, isAuthenticated, loading: authLoading } = useAuth();
  const params = useParams();
  const router = useRouter();
  const userId = params.userId;
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isOwner, setIsOwner] = useState(false);
  const [toggleLoading, setToggleLoading] = useState(false);

  const [followLoading, setFollowLoading] = useState(false);
  const [followStatus, setFollowStatus] = useState('not_following'); // 'not_following', 'following', 'request_sent'
  const [activeTab, setActiveTab] = useState('about');
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const { myFollowRequests } = useMyFollowRequests();

  // Check follow status function
  const checkFollowStatus = async () => {
    if (!isAuthenticated || !user || isOwner) return;
    
    try {
      const response = await fetch(`http://localhost:8080/api/v1/follow-status/${userId}`, {
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        setFollowStatus(data.status);
      } else {
        setFollowStatus('not_following');
      }
    } catch (err) {
      console.error('Error checking follow status:', err);
      setFollowStatus('not_following');
    }
  };

  // Fetch profile function
  const fetchProfile = async () => {
    if (!isAuthenticated) return;
    
    setLoading(true);
    setError('');
    try {
      const data = await userAPI.getUserProfile(userId);
      setProfile(data);
      // Check if the current user is the owner
      setIsOwner(user && user.id === data.id);
      
      // Check follow status if not the owner
      if (!isOwner) {
        await checkFollowStatus();
      }
    } catch (err) {
      console.error('Error fetching profile:', err);
      if (err.message.includes('403')) {
        setError('This profile is private.');
        setProfile(null);
      } else if (err.message.includes('401')) {
        // Redirect to login page if not authenticated
        router.push('/auth/login');
        return;
      } else {
        setError('Failed to fetch profile');
        setProfile(null);
      }
    } finally {
      setLoading(false);
    }
  };

  // Users Section Component
  const UsersSection = () => {
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [followLoading, setFollowLoading] = useState(new Set());
    const { myFollowRequests, cancelFollowRequest, refresh: refreshMyFollowRequests } = useMyFollowRequests();

    const fetchUsers = useCallback(async () => {
      try {
        setLoading(true);
        setError('');
        const response = await userAPI.getAllUsers();
        
        // Handle null or undefined response
        if (!response || !Array.isArray(response)) {
          setUsers([]);
          setError('No users found');
          return;
        }
        
        // Just use the users as they are, without follow status
        const usersWithStatus = response.map(user => ({
          ...user,
          isFollowing: false // Default to not following
        }));
        
        setUsers(usersWithStatus);
        if (usersWithStatus.length === 0) {
          setError('No users found');
        }
      } catch (err) {
        console.error('Error fetching users:', err);
        setUsers([]);
        setError('No users found');
      } finally {
        setLoading(false);
      }
    }, []);

    useEffect(() => {
      fetchUsers();
    }, [fetchUsers]);

    const handleFollowAction = async (userId, isFollowing) => {
      try {
        setFollowLoading(prev => new Set(prev).add(userId));
        
        if (isFollowing) {
          await followAPI.unfollowUser(userId);
          setUsers(prev => prev.map(user => 
            user.id === userId ? { ...user, isFollowing: false } : user
          ));
        } else {
          try {
            await followAPI.followUser(userId);
            // Update the user's follow status based on their profile privacy
            const user = users.find(u => u.id === userId);
            if (user && user.isPublic) {
              setUsers(prev => prev.map(u => 
                u.id === userId ? { ...u, isFollowing: true } : u
              ));
            } else {
              // For private profiles, show "Request Sent" status
              setUsers(prev => prev.map(u => 
                u.id === userId ? { ...u, isFollowing: 'request_sent' } : u
              ));
              // Refresh my follow requests list to include the new request
              await refreshMyFollowRequests();
            }
          } catch (followError) {
            // If the error is "Follow request already sent", update the status
            if (followError.message && followError.message.includes('Follow request already sent')) {
              setUsers(prev => prev.map(u => 
                u.id === userId ? { ...u, isFollowing: 'request_sent' } : u
              ));
              // Refresh my follow requests list
              await refreshMyFollowRequests();
            } else {
              throw followError; // Re-throw other errors
            }
          }
        }
      } catch (error) {
        console.error('Follow action failed:', error);
        alert(error.message || 'Failed to perform follow action');
      } finally {
        setFollowLoading(prev => {
          const newSet = new Set(prev);
          newSet.delete(userId);
          return newSet;
        });
      }
    };

    const handleCancelRequest = async (userId) => {
      try {
        setFollowLoading(prev => new Set(prev).add(userId));
        
        // Find the follow request for this user
        const followRequest = myFollowRequests.find(req => req.recipient.id === userId);
        if (!followRequest) {
          throw new Error('Follow request not found');
        }
        
        await cancelFollowRequest(followRequest.id);
        
        // Update the user's follow status
        setUsers(prev => prev.map(u => 
          u.id === userId ? { ...u, isFollowing: false } : u
        ));
      } catch (error) {
        console.error('Cancel request failed:', error);
        alert(error.message || 'Failed to cancel follow request');
      } finally {
        setFollowLoading(prev => {
          const newSet = new Set(prev);
          newSet.delete(userId);
          return newSet;
        });
      }
    };

    if (loading) {
      return (
        <div style={{ textAlign: 'center', padding: '40px 20px' }}>
          <div style={{ fontSize: '14px', color: '#666' }}>Loading users...</div>
        </div>
      );
    }

    if (error) {
      return (
        <div>
          <div style={{ textAlign: 'center', padding: '20px', color: '#666' }}>
            <div style={{ fontSize: '14px' }}>{error}</div>
            {error === 'No users found' && (
              <div style={{ fontSize: '12px', marginTop: '8px', color: '#999' }}>
                There are no other users to display at the moment.
              </div>
            )}
          </div>
        </div>
      );
    }

    return (
      <div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          {Array.isArray(users) && users.length > 0 ? (
            users.map(user => {
              const isLoading = followLoading.has(user.id);
              
              return (
                <div key={user.id} style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '12px',
                  padding: '12px',
                  backgroundColor: '#fff',
                  borderRadius: '8px',
                  border: '1px solid #eee'
                }}>
                  <div style={{
                    width: '40px',
                    height: '40px',
                    borderRadius: '50%',
                    backgroundColor: '#f0f0f0',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '14px',
                    color: '#666'
                  }}>
                    {user.avatarPath ? (
                      <img 
                        src={`/${user.avatarPath}`} 
                        alt="Avatar" 
                        style={{ width: '100%', height: '100%', borderRadius: '50%', objectFit: 'cover' }}
                      />
                    ) : (
                      (user.firstName?.charAt(0) || '') + (user.lastName?.charAt(0) || '')
                    )}
                  </div>
                  
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: '500', fontSize: '14px', color: '#333' }}>
                      {user.firstName} {user.lastName}
                    </div>
                    {user.nickname && (
                      <div style={{ fontSize: '12px', color: '#666' }}>
                        @{user.nickname}
                      </div>
                    )}
                  </div>
                  
                  {user.isFollowing === 'request_sent' ? (
                    <div style={{ display: 'flex', gap: '6px' }}>
                      <button
                        style={{
                          backgroundColor: '#ff9800',
                          color: 'white',
                          border: 'none',
                          borderRadius: '6px',
                          padding: '6px 8px',
                          fontSize: '11px',
                          fontWeight: '500',
                          cursor: 'default',
                          whiteSpace: 'nowrap'
                        }}
                        disabled
                      >
                        Request Sent
                      </button>
                      <button
                        style={{
                          backgroundColor: '#f44336',
                          color: 'white',
                          border: 'none',
                          borderRadius: '6px',
                          padding: '6px 8px',
                          fontSize: '11px',
                          fontWeight: '500',
                          cursor: isLoading ? 'not-allowed' : 'pointer',
                          opacity: isLoading ? 0.6 : 1,
                          whiteSpace: 'nowrap'
                        }}
                        onClick={() => handleCancelRequest(user.id)}
                        disabled={isLoading}
                      >
                        {isLoading ? '...' : 'Cancel'}
                      </button>
                    </div>
                  ) : (
                    <button
                      style={{
                        backgroundColor: user.isFollowing === true ? '#f44336' : '#4CAF50',
                        color: 'white',
                        border: 'none',
                        borderRadius: '6px',
                        padding: '6px 12px',
                        fontSize: '12px',
                        fontWeight: '500',
                        cursor: isLoading ? 'not-allowed' : 'pointer',
                        opacity: isLoading ? 0.6 : 1,
                        whiteSpace: 'nowrap'
                      }}
                      onClick={() => handleFollowAction(user.id, user.isFollowing)}
                      disabled={isLoading}
                    >
                      {isLoading ? 'Loading...' : 
                       user.isFollowing === true ? 'Unfollow' : 'Follow'}
                    </button>
                  )}
                </div>
              );
            })
          ) : (
            <div style={{ 
              textAlign: 'center', 
              padding: '40px 20px', 
              color: '#666',
              fontSize: '14px'
            }}>
              No users found.
            </div>
          )}
        </div>
      </div>
    );
  };

  useEffect(() => {
    if (isAuthenticated) {
      fetchProfile();
    }
  }, [userId, router, isAuthenticated, user, isOwner]);

  const handleTogglePrivacy = async () => {
    if (!profile) return;
    setToggleLoading(true);
    try {
      const data = await userAPI.togglePrivacy();
      setProfile(prev => ({ ...prev, isPublic: data.isPublic }));
    } catch (err) {
      console.error('Failed to toggle privacy:', err);
      alert('Failed to update privacy');
    } finally {
      setToggleLoading(false);
    }
  };

  const handleFollowAction = async () => {
    if (!profile || isOwner) return;
    
    setFollowLoading(true);
    try {
      if (followStatus === 'following') {
        // Unfollow
        await followAPI.unfollowUser(userId);
        setFollowStatus('not_following');
      } else if (followStatus === 'request_sent') {
        // Cancel follow request
        const followRequest = myFollowRequests.find(req => req.recipient.id === userId);
        if (followRequest) {
          await followAPI.cancelFollowRequest(followRequest.id);
          setFollowStatus('not_following');
        } else {
          alert('Follow request not found. Please refresh the page.');
        }
      } else {
        // Follow or send request
        try {
          await followAPI.followUser(userId);
          if (profile.isPublic) {
            setFollowStatus('following');
          } else {
            setFollowStatus('request_sent');
          }
        } catch (followError) {
          // If the error is "Follow request already sent", update the status
          if (followError.message && followError.message.includes('Follow request already sent')) {
            setFollowStatus('request_sent');
          } else {
            throw followError; // Re-throw other errors
          }
        }
      }
    } catch (error) {
      console.error('Follow action failed:', error);
      alert(error.message || 'Failed to perform follow action');
    } finally {
      setFollowLoading(false);
    }
  };

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
    <div className="page-enter" style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      minHeight: '100vh',
      background: '#f0f2f5'
    }}>
      <div className="content-placeholder" style={{ 
        width: '80px', 
        height: '80px', 
        borderRadius: '50%'
      }}></div>
    </div>
  );
  
  if (error) return (
    <div className="page-enter" style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      minHeight: '100vh',
      background: '#f0f2f5'
    }}>
      <div style={{ fontSize: '18px', color: '#e74c3c' }}>{error}</div>
    </div>
  );
  
  if (!profile) return (
    <div className="page-enter" style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      minHeight: '100vh',
      background: '#f0f2f5'
    }}>
      <div style={{ fontSize: '18px', color: '#666' }}>Profile not found...</div>
    </div>
  );

  return (
    <div className="page-slide-right" style={{ 
      minHeight: '100vh', 
      background: '#f0f2f5',
      animation: 'slideInFromRight 0.4s ease-out'
    }}>
      {/* Navigation Bar */}
      <nav style={{
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
      }}>
        <Link href="/home" className="nav-link" style={{ 
          textDecoration: 'none', 
          color: '#fff', 
          fontSize: '20px', 
          fontWeight: 'bold',
          display: 'flex',
          alignItems: 'center',
          gap: '12px'
        }}>
          <span style={{ fontSize: '24px' }}>←</span>
          Social Network
        </Link>
        <div style={{ color: '#fff', fontSize: '16px' }}>
          Profile
        </div>
      </nav>

      {/* Main Content */}
      <div style={{ paddingTop: '80px' }}>
        {/* Profile Header */}
        <div style={{
          background: '#fff',
          borderBottom: '1px solid #ddd',
          padding: '32px 0'
        }}>
          <div style={{
            maxWidth: '1000px',
            margin: '0 auto',
            padding: '0 32px'
          }}>
            <div style={{
              display: 'flex',
              alignItems: 'flex-end',
              gap: '24px',
              marginBottom: '24px'
            }}>
              <img 
                src={profile.avatarPath ? `/${profile.avatarPath}` : '/user.png'} 
                alt="avatar" 
                className="profile-image"
                style={{ 
                  width: '168px', 
                  height: '168px', 
                  borderRadius: '50%', 
                  objectFit: 'cover', 
                  border: '4px solid #fff',
                  boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
                }} 
              />
              <div style={{ flex: 1, marginBottom: '16px' }}>
                <h1 style={{ 
                  margin: '0 0 8px 0', 
                  fontSize: '32px', 
                  fontWeight: '700',
                  color: '#050505'
                }}>
                  {profile.firstName} {profile.lastName}
                  {profile.nickname && (
                    <span style={{ 
                      color: '#65676b', 
                      fontWeight: '400',
                      marginLeft: '8px'
                    }}>
                      ({profile.nickname})
                    </span>
                  )}
                </h1>
                <div style={{ 
                  color: '#65676b', 
                  fontSize: '17px',
                  marginBottom: '8px'
                }}>
                  {profile.email}
                </div>
                <div style={{ 
                  color: '#65676b', 
                  fontSize: '15px',
                  marginBottom: '16px'
                }}>
                  Date of Birth: {profile.dateOfBirth}
                </div>
                <div style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  gap: '12px',
                  marginBottom: '16px'
                }}>
                  <span style={{ 
                    fontWeight: '500', 
                    color: profile.isPublic ? '#2e7d32' : '#b74115',
                    fontSize: '15px'
                  }}>
                    {profile.isPublic ? 'Public Profile' : 'Private Profile'}
                  </span>

                </div>
                {isOwner && (
                  <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
                    <button
                      className="btn-smooth"
                      style={{ 
                        background: '#b74115',
                        color: '#fff', 
                        border: 'none', 
                        borderRadius: '8px', 
                        padding: '12px 24px', 
                        fontWeight: '600', 
                        cursor: 'pointer',
                        fontSize: '15px'
                      }}
                      onClick={() => setIsEditModalOpen(true)}
                    >
                      Edit Profile
                    </button>
                    <select
                      value={profile.isPublic ? 'public' : 'private'}
                      onChange={(e) => handleTogglePrivacy()}
                      disabled={toggleLoading}
                      className="btn-smooth"
                      style={{ 
                        padding: '6px 12px', 
                        borderRadius: '6px', 
                        border: '1px solid #ddd', 
                        fontSize: '14px',
                        background: '#fff'
                      }}
                    >
                      <option value="public">Public</option>
                      <option value="private">Private</option>
                    </select>
                  </div>
                )}
                {!isOwner && (
                  <button
                    className="btn-smooth"
                    style={{ 
                      background: followStatus === 'following' ? '#f44336' : 
                                 followStatus === 'request_sent' ? '#ff9800' : '#4CAF50', 
                      color: '#fff', 
                      border: 'none', 
                      borderRadius: '8px', 
                      padding: '12px 24px', 
                      fontWeight: '600', 
                      cursor: followLoading ? 'not-allowed' : 'pointer', 
                      opacity: followLoading ? 0.7 : 1,
                      fontSize: '15px'
                    }}
                    onClick={handleFollowAction}
                    disabled={followLoading}
                  >
                    {followLoading ? 'Loading...' : 
                      followStatus === 'following' ? 'Unfollow' : 
                      followStatus === 'request_sent' ? 'Request Sent' : 
                      'Follow'
                    }
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Profile Content */}
        <div style={{
          maxWidth: '1000px',
          margin: '0 auto',
          padding: '32px'
        }}>
          {/* Tabs */}
          <div style={{ 
            display: 'flex', 
            gap: '48px', 
            borderBottom: '1px solid #ddd',
            marginBottom: '32px'
          }}>
            <button
              className="nav-link"
              style={{
                background: 'none',
                border: 'none',
                borderBottom: activeTab === 'about' ? '3px solid #b74115' : '3px solid transparent',
                color: activeTab === 'about' ? '#b74115' : '#65676b',
                fontWeight: activeTab === 'about' ? '700' : '500',
                fontSize: '17px',
                padding: '16px 0',
                cursor: 'pointer',
                outline: 'none'
              }}
              onClick={() => setActiveTab('about')}
            >
              About
            </button>
            <button
              className="nav-link"
              style={{
                background: 'none',
                border: 'none',
                borderBottom: activeTab === 'followers' ? '3px solid #b74115' : '3px solid transparent',
                color: activeTab === 'followers' ? '#b74115' : '#65676b',
                fontWeight: activeTab === 'followers' ? '700' : '500',
                fontSize: '17px',
                padding: '16px 0',
                cursor: 'pointer',
                outline: 'none'
              }}
              onClick={() => setActiveTab('followers')}
            >
              Followers ({profile.followers.length})
            </button>
                         <button
               className="nav-link"
               style={{
                 background: 'none',
                 border: 'none',
                 borderBottom: activeTab === 'following' ? '3px solid #b74115' : '3px solid transparent',
                 color: activeTab === 'following' ? '#b74115' : '#65676b',
                 fontWeight: activeTab === 'following' ? '700' : '500',
                 fontSize: '17px',
                 padding: '16px 0',
                 cursor: 'pointer',
                 outline: 'none'
               }}
               onClick={() => setActiveTab('following')}
             >
               Following ({profile.following.length})
             </button>
             <button
               className="nav-link"
               style={{
                 background: 'none',
                 border: 'none',
                 borderBottom: activeTab === 'discover' ? '3px solid #b74115' : '3px solid transparent',
                 color: activeTab === 'discover' ? '#b74115' : '#65676b',
                 fontWeight: activeTab === 'discover' ? '700' : '500',
                 fontSize: '17px',
                 padding: '16px 0',
                 cursor: 'pointer',
                 outline: 'none'
               }}
               onClick={() => setActiveTab('discover')}
             >
               People You May Know
             </button>
          </div>

          {/* Tab Content */}
          <div className="card-smooth" style={{ 
            background: '#fff', 
            borderRadius: '8px', 
            padding: '24px',
            boxShadow: '0 1px 2px rgba(0,0,0,0.1)'
          }}>
            {activeTab === 'about' && (
              <div>
                <h3 style={{ 
                  margin: '0 0 16px 0', 
                  fontSize: '20px', 
                  fontWeight: '600',
                  color: '#050505'
                }}>
                  About
                </h3>
                <div style={{ 
                  color: '#65676b', 
                  fontSize: '15px',
                  lineHeight: '1.5'
                }}>
                  {profile.aboutMe || 'No information provided.'}
                </div>
              </div>
            )}
            
            {activeTab === 'followers' && (
              <div>
                <h3 style={{ 
                  margin: '0 0 16px 0', 
                  fontSize: '20px', 
                  fontWeight: '600',
                  color: '#050505'
                }}>
                  Followers ({profile.followers.length})
                </h3>
                {profile.followers.length === 0 ? (
                  <div style={{ 
                    color: '#65676b', 
                    textAlign: 'center', 
                    padding: '40px 20px', 
                    fontSize: '15px' 
                  }}>
                    No followers yet.
                  </div>
                ) : (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                    {profile.followers.map(user => (
                      <div key={user.id} className="card-smooth" style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '12px',
                        padding: '12px',
                        borderRadius: '8px',
                        background: '#fff',
                        border: '1px solid #eee'
                      }}>
                        <img 
                          src={user.avatarPath ? `/${user.avatarPath}` : '/user.png'} 
                          alt="avatar" 
                          className="profile-image"
                          style={{ 
                            width: '40px', 
                            height: '40px', 
                            borderRadius: '50%', 
                            objectFit: 'cover' 
                          }} 
                        />
                        <div>
                          <div style={{ 
                            fontWeight: '600', 
                            fontSize: '15px',
                            color: '#050505'
                          }}>
                            {user.firstName} {user.lastName}
                          </div>
                          {user.nickname && (
                            <div style={{ 
                              fontSize: '13px', 
                              color: '#65676b' 
                            }}>
                              @{user.nickname}
                            </div>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
            
            {activeTab === 'following' && (
              <div>
                <h3 style={{ 
                  margin: '0 0 16px 0', 
                  fontSize: '20px', 
                  fontWeight: '600',
                  color: '#050505'
                }}>
                  Following ({profile.following.length})
                </h3>
                {profile.following.length === 0 ? (
                  <div style={{ 
                    color: '#65676b', 
                    textAlign: 'center', 
                    padding: '40px 20px', 
                    fontSize: '15px' 
                  }}>
                    Not following anyone.
                  </div>
                ) : (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                    {profile.following.map(user => (
                      <div key={user.id} className="card-smooth" style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '12px',
                        padding: '12px',
                        borderRadius: '8px',
                        background: '#fff',
                        border: '1px solid #eee'
                      }}>
                        <img 
                          src={user.avatarPath ? `/${user.avatarPath}` : '/user.png'} 
                          alt="avatar" 
                          className="profile-image"
                          style={{ 
                            width: '40px', 
                            height: '40px', 
                            borderRadius: '50%', 
                            objectFit: 'cover' 
                          }} 
                        />
                        <div>
                          <div style={{ 
                            fontWeight: '600', 
                            fontSize: '15px',
                            color: '#050505'
                          }}>
                            {user.firstName} {user.lastName}
                          </div>
                          {user.nickname && (
                            <div style={{ 
                              fontSize: '13px', 
                              color: '#65676b' 
                            }}>
                              @{user.nickname}
                            </div>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                                 )}
               </div>
             )}
             
             {activeTab === 'discover' && (
               <div>
                 <h3 style={{ 
                   margin: '0 0 16px 0', 
                   fontSize: '20px', 
                   fontWeight: '600',
                   color: '#050505'
                 }}>
                   People You May Know
                 </h3>
                 <UsersSection />
               </div>
             )}
           </div>
         </div>
      </div>

      {/* Edit Profile Modal */}
      <EditProfile 
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        onProfileUpdated={() => {
          // Refresh the profile data after successful update
          try {
            fetchProfile();
          } catch (error) {
            console.error('Error refreshing profile:', error);
            // Silently handle the error to avoid pop-ups
          }
        }}
      />

    </div>
  );
}


