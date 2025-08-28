"use client";
import { useEffect, useState } from "react";

export default function PostList({ userId }) {
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`http://localhost:8080/api/v1/posts`) // adjust to your Go API
      .then((res) => res.json())
      .then((data) => {
        setPosts(data);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  if (loading) return <p>Loading posts...</p>;
  if (!posts.length) return <p>No posts yet.</p>;

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
      {posts.map((post) => (
        <div key={post.id} style={styles.card}>
          <h3>{post.authorName}</h3>
          <p>{post.content}</p>
        </div>
      ))}
    </div>
  );
}

const styles = {
  card: {
    border: "1px solid #ddd",
    borderRadius: "12px",
    padding: "16px",
    background: "#fff",
    boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
  },
};
