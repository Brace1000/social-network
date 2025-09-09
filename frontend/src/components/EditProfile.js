'use client';

import React, { useState, useEffect } from 'react';
import { useAuth } from '../store/AuthContext';

const styles = {
  overlay: {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    zIndex: 1000,
  },
  modal: {
    background: '#fff',
    borderRadius: '12px',
    padding: '24px',
    width: '90%',
    maxWidth: '500px',
    maxHeight: '90vh',
    overflowY: 'auto',
    boxShadow: '0 10px 25px rgba(0, 0, 0, 0.2)',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '24px',
    paddingBottom: '16px',
    borderBottom: '1px solid #eee',
  },
  title: {
    fontSize: '20px',
    fontWeight: '600',
    color: '#333',
    margin: 0,
  },
  closeButton: {
    background: 'none',
    border: 'none',
    fontSize: '24px',
    cursor: 'pointer',
    color: '#666',
    padding: '4px',
    borderRadius: '50%',
    width: '32px',
    height: '32px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  form: {
    display: 'flex',
    flexDirection: 'column',
    gap: '16px',
  },
  field: {
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
  },
  label: {
    fontSize: '14px',
    fontWeight: '500',
    color: '#333',
  },
  input: {
    padding: '12px 16px',
    border: '2px solid #e0e0e0',
    borderRadius: '8px',
    fontSize: '16px',
    outline: 'none',
    transition: 'border-color 0.2s',
  },
  textarea: {
    padding: '12px 16px',
    border: '2px solid #e0e0e0',
    borderRadius: '8px',
    fontSize: '16px',
    outline: 'none',
    transition: 'border-color 0.2s',
    minHeight: '100px',
    resize: 'vertical',
    fontFamily: 'inherit',
  },
  fileInput: {
    padding: '12px 16px',
    border: '2px solid #e0e0e0',
    borderRadius: '8px',
    fontSize: '16px',
    outline: 'none',
  },
  avatarPreview: {
    width: '80px',
    height: '80px',
    borderRadius: '50%',
    objectFit: 'cover',
    border: '2px solid #eee',
    marginTop: '8px',
  },
  privacyToggle: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
    padding: '12px 16px',
    border: '2px solid #e0e0e0',
    borderRadius: '8px',
    cursor: 'pointer',
    transition: 'border-color 0.2s',
  },
  toggle: {
    position: 'relative',
    width: '44px',
    height: '24px',
    backgroundColor: '#ccc',
    borderRadius: '12px',
    transition: 'background-color 0.2s',
  },
  toggleActive: {
    backgroundColor: '#b74115',
  },
  toggleThumb: {
    position: 'absolute',
    top: '2px',
    left: '2px',
    width: '20px',
    height: '20px',
    backgroundColor: '#fff',
    borderRadius: '50%',
    transition: 'transform 0.2s',
  },
  toggleThumbActive: {
    transform: 'translateX(20px)',
  },
  buttonGroup: {
    display: 'flex',
    gap: '12px',
    marginTop: '24px',
  },
  saveButton: {
    flex: 1,
    background: '#b74115',
    color: '#fff',
    border: 'none',
    borderRadius: '8px',
    padding: '12px 24px',
    fontSize: '16px',
    fontWeight: '600',
    cursor: 'pointer',
    transition: 'background-color 0.2s',
  },
  cancelButton: {
    flex: 1,
    background: '#f5f5f5',
    color: '#333',
    border: '2px solid #e0e0e0',
    borderRadius: '8px',
    padding: '12px 24px',
    fontSize: '16px',
    fontWeight: '600',
    cursor: 'pointer',
    transition: 'background-color 0.2s',
  },
  loadingButton: {
    opacity: 0.6,
    cursor: 'not-allowed',
  },
  error: {
    color: '#e74c3c',
    fontSize: '14px',
    marginTop: '8px',
  },
  success: {
    color: '#27ae60',
    fontSize: '14px',
    marginTop: '8px',
  },
};

export default function EditProfile({ isOpen, onClose, onProfileUpdated }) {
  const { user } = useAuth();
  const [form, setForm] = useState({
    firstName: '',
    lastName: '',
    nickname: '',
    aboutMe: '',
    isPublic: true,
  });
  const [avatar, setAvatar] = useState(null);
  const [avatarPreview, setAvatarPreview] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    if (isOpen && user) {
      // Load current profile data
      fetchProfile();
    }
  }, [isOpen, user]);

  const fetchProfile = async () => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/profile/${user.id}`, {
        credentials: 'include',
      });
      
      if (response.ok) {
        const profile = await response.json();
        setForm({
          firstName: profile.firstName || '',
          lastName: profile.lastName || '',
          nickname: profile.nickname || '',
          aboutMe: profile.aboutMe || '',
          isPublic: profile.isPublic !== undefined ? profile.isPublic : true,
        });
        setAvatarPreview(profile.avatarPath);
      }
    } catch (err) {
      console.error('Failed to fetch profile:', err);
      setError('Failed to load profile data');
    }
  };

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setForm(prev => ({ ...prev, [name]: value }));
  };

  const handleAvatarChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      setAvatar(file);
      const reader = new FileReader();
      reader.onload = (e) => setAvatarPreview(e.target.result);
      reader.readAsDataURL(file);
    }
  };

  const handlePrivacyToggle = () => {
    setForm(prev => ({ ...prev, isPublic: !prev.isPublic }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess('');

    try {
      // Update profile information
      const profileResponse = await fetch('http://localhost:8080/api/v1/profile', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(form),
      });

      if (!profileResponse.ok) {
        const errorData = await profileResponse.text();
        throw new Error(`Failed to update profile: ${profileResponse.status} ${errorData}`);
      }

      // Upload avatar if selected
      if (avatar) {
        const formData = new FormData();
        formData.append('avatar', avatar);

        const avatarResponse = await fetch('http://localhost:8080/api/v1/profile/avatar', {
          method: 'POST',
          credentials: 'include',
          body: formData,
        });

        if (!avatarResponse.ok) {
          const errorData = await avatarResponse.text();
          throw new Error(`Failed to upload avatar: ${avatarResponse.status} ${errorData}`);
        }
      }

      setSuccess('Profile updated successfully!');
      
      // Close modal after a short delay
      setTimeout(() => {
        onClose();
        // Call the callback after closing the modal to avoid any potential issues
        if (onProfileUpdated) {
          onProfileUpdated();
        }
      }, 1500);

    } catch (err) {
      console.error('Profile update error:', err);
      // Show a more specific error message based on the error
      if (err.message.includes('Failed to upload avatar')) {
        setError('Failed to upload profile picture. Please try a different image.');
      } else if (err.message.includes('Failed to update profile')) {
        setError('Failed to update profile information. Please check your input and try again.');
      } else {
        setError('Failed to update profile. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div style={styles.overlay} onClick={onClose}>
      <div style={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div style={styles.header}>
          <h2 style={styles.title}>Edit Profile</h2>
          <button style={styles.closeButton} onClick={onClose}>
            ×
          </button>
        </div>

        <form style={styles.form} onSubmit={handleSubmit}>
          <div style={styles.field}>
            <label style={styles.label}>First Name</label>
            <input
              type="text"
              name="firstName"
              value={form.firstName}
              onChange={handleInputChange}
              style={styles.input}
              required
            />
          </div>

          <div style={styles.field}>
            <label style={styles.label}>Last Name</label>
            <input
              type="text"
              name="lastName"
              value={form.lastName}
              onChange={handleInputChange}
              style={styles.input}
              required
            />
          </div>

          <div style={styles.field}>
            <label style={styles.label}>Nickname (Optional)</label>
            <input
              type="text"
              name="nickname"
              value={form.nickname}
              onChange={handleInputChange}
              style={styles.input}
              placeholder="@username"
            />
          </div>

          <div style={styles.field}>
            <label style={styles.label}>About Me</label>
            <textarea
              name="aboutMe"
              value={form.aboutMe}
              onChange={handleInputChange}
              style={styles.textarea}
              placeholder="Tell us about yourself..."
              maxLength={500}
            />
          </div>

          <div style={styles.field}>
            <label style={styles.label}>Profile Picture</label>
            <input
              type="file"
              accept="image/*"
              onChange={handleAvatarChange}
              style={styles.fileInput}
            />
            {avatarPreview && (
              <img
                src={avatarPreview}
                alt="Avatar preview"
                style={styles.avatarPreview}
              />
            )}
          </div>

          <div style={styles.field}>
            <label style={styles.label}>Privacy</label>
            <div style={styles.privacyToggle} onClick={handlePrivacyToggle}>
              <div style={{
                ...styles.toggle,
                ...(form.isPublic ? styles.toggleActive : {})
              }}>
                <div style={{
                  ...styles.toggleThumb,
                  ...(form.isPublic ? styles.toggleThumbActive : {})
                }} />
              </div>
              <span>{form.isPublic ? 'Public Profile' : 'Private Profile'}</span>
            </div>
          </div>

          {error && <div style={styles.error}>{error}</div>}
          {success && <div style={styles.success}>{success}</div>}

          <div style={styles.buttonGroup}>
            <button
              type="button"
              style={styles.cancelButton}
              onClick={onClose}
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              style={{
                ...styles.saveButton,
                ...(loading && styles.loadingButton)
              }}
              disabled={loading}
            >
              {loading ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
} 