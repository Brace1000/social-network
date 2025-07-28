const API_BASE_URL = 'http://localhost:8080/api/v1';

// Helper function to make API calls
export async function apiCall(endpoint, method = 'GET', body = null) {
  const url = `${API_BASE_URL}${endpoint}`;
  const options = {
    method,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
  };

  if (body) {
    options.body = JSON.stringify(body);
  }

  try {
    const response = await fetch(url, options);
    
    if (!response.ok) {
      const errorText = await response.text();
      const errorMessage = errorText || `HTTP error! status: ${response.status}`;
      
      // Don't log 401 errors as they are expected for non-authenticated users
      if (response.status === 401) {
        console.log('DEBUG: Suppressing 401 error for endpoint:', endpoint);
      } else {
        console.error(`API call failed for ${endpoint}:`, errorMessage);
      }
      
      throw new Error(errorMessage);
    }

    // Handle empty responses
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      const data = await response.json();
      return data;
    }
    
    // For endpoints that should return JSON but don't have content-type header
    // Try to parse as JSON anyway
    const text = await response.text();
    if (text.trim()) {
      try {
        return JSON.parse(text);
      } catch (parseError) {
        console.warn('Failed to parse response as JSON:', parseError);
        return null;
      }
    }
    
    return null;
  } catch (error) {
    // Only log errors that weren't already handled above
    if (!error.message.includes('401') && !error.message.includes('Unauthorized')) {
      console.error(`API call failed for ${endpoint}:`, error);
    }
    throw error;
  }
}

// User-related API calls
export const userAPI = {
  // Get current user
  getCurrentUser: () => apiCall('/me'),
  
  // Get all users
  getAllUsers: async () => {
    try {
      const response = await apiCall('/users');
      return Array.isArray(response) ? response : [];
    } catch (error) {
      console.error('Failed to get all users:', error);
      return [];
    }
  },
  
  // Get user profile
  getUserProfile: (userId) => apiCall(`/profile/${userId}`),
  
  // Update user profile
  updateProfile: (data) => apiCall('/profile', 'PUT', data),
  
  // Upload avatar
  uploadAvatar: (formData) => {
    const url = `${API_BASE_URL}/profile/avatar`;
    return fetch(url, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    });
  },
  
  // Toggle profile privacy
  togglePrivacy: () => apiCall('/profile/toggle-privacy', 'POST'),
};

// Follow-related API calls
export const followAPI = {
  followUser: (userId) => apiCall(`/follow/${userId}`, 'POST'),
  
  unfollowUser: (userId) => apiCall(`/unfollow/${userId}`, 'POST'),


  
  getFollowRequests: () => apiCall('/follow-requests'),
  
  getMyFollowRequests: () => apiCall('/my-follow-requests'),
  
  acceptFollowRequest: (requestId) => apiCall(`/follow-requests/${requestId}/accept`, 'POST'),
  
  declineFollowRequest: (requestId) => apiCall(`/follow-requests/${requestId}/decline`, 'POST'),
  
  cancelFollowRequest: (requestId) => apiCall(`/follow-requests/${requestId}/cancel`, 'POST'),
};

// Notification-related API calls
export const notificationAPI = {
  getNotifications: () => apiCall('/notifications'),
  
  markAsRead: (notificationId) => apiCall(`/notifications/${notificationId}/read`, 'POST'),
};

// Auth-related API calls
export const authAPI = {
  register: (data) => apiCall('/register', 'POST', data),
  
  login: (data) => apiCall('/login', 'POST', data),
  
  logout: () => apiCall('/logout', 'POST'),
};