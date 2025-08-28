"use client";
import { useState } from "react";

export default function PostForm({ onSuccess }) {
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e) {
    e.preventDefault();
    setLoading(true);

    const res = await fetch("http://localhost:8080/api/v1/posts", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ content }),
    });

    if (res.ok) {
      setContent("");
      onSuccess && onSuccess(); // refresh posts
    }

    setLoading(false);
  }

  return (
    <form onSubmit={handleSubmit} style={styles.form}>
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="What's on your mind?"
        required
        style={styles.textarea}
      />
      <button type="submit" disabled={loading} style={styles.button}>
        {loading ? "Posting..." : "Post"}
      </button>
    </form>
  );
}

const styles = {
  form: {
    display: "flex",
    flexDirection: "column",
    gap: "12px",
    marginBottom: "20px",
  },
  textarea: {
    width: "100%",
    minHeight: "80px",
    borderRadius: "8px",
    border: "1px solid #ccc",
    padding: "10px",
    fontSize: "14px",
  },
  button: {
    backgroundColor: "#b74115",
    color: "white",
    padding: "10px",
    border: "none",
    borderRadius: "8px",
    cursor: "pointer",
  },
};
