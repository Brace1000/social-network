"use client";

import { useState } from 'react';

export default function PostForm() {
  const [content, setContent] = useState('');
  const [image, setImage] = useState(null);
  const [privacy, setPrivacy] = useState(1); // Default to public

  const handleFileChange = (e) => {
    setImage(e.target.files[0]);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    const formData = new FormData();
    formData.append('content', content);
    // Map frontend int → backend string
    const privacyMap = { 1: "public", 2: "friends", 3: "private" };
    formData.append("privacy", privacyMap[privacy]);
    if (image) {
      formData.append('image', image);
    }

    try {
    const response = await fetch('http://localhost:8080/api/v1/posts', {
      method: 'POST',
      body: formData,
      credentials: "include",
    });

    if (response.ok) {
      // Handle success (e.g., clear form, show success message)
      setContent('');
      setImage(null);
      setPrivacy(1); //reset to public
      console.log('Post created successfully!');
    } else {
      console.error('Failed to create post');
    }
  } catch (error){
    console.log("Backend not available, simulating post creation")
      setContent("")
      setImage(null)
      setPrivacy(1)
    }
  };

  return (
    <div className="post-form-container">
      <form onSubmit={handleSubmit}>
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="What's on your mind?"
          rows="4"
          required
        />
        <div className="file-input-container">
          <label htmlFor="file-upload">
            <span className="upload-icon">📷</span> Add Photo/Video
          </label>
          <input
            id="file-upload"
            type="file"
            accept="image/*,video/*"
            onChange={handleFileChange}
          />
          {image && <p>{image.name} selected</p>}
        </div>
        <div className="privacy-select-container">
          <label>
            <input
              type="radio"
              value="1"
              checked={privacy === 1}
              onChange={() => setPrivacy(1)}
            />
            Public
          </label>
          <label>
            <input
              type="radio"
              value="2"
              checked={privacy === 2}
              onChange={() => setPrivacy(2)}
            />
            Followers
          </label>
          <label>
            <input
              type="radio"
              value="3"
              checked={privacy === 3}
              onChange={() => setPrivacy(3)}
            />
            Private
          </label>
        </div>
        <button type="submit">Post</button>
      </form>
    </div>
  );
}