"use client";

import React, { useState } from "react";
import Link from "next/link";

export default function RegisterForm({ onSuccess, onError }) {
  const [form, setForm] = useState({
    nickname: "",
    firstName: "",
    lastName: "",
    email: "",
    dateOfBirth: "",
    avatar: null,
    aboutMe: "",
    password: "",
  });
  const [preview, setPreview] = useState(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const handleChange = (e) => {
    const { name, value, files } = e.target;
    if (name === "avatar") {
      const file = files[0];
      setForm((f) => ({ ...f, avatar: file }));
      if (file) setPreview(URL.createObjectURL(file));
    } else {
      setForm((f) => ({ ...f, [name]: value }));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const { avatar, ...rest } = form; // skip avatar for now
      const res = await fetch("http://localhost:8080/api/v1/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(rest),
        credentials: "include",
      });
      if (!res.ok) {
        const data = await res.json();
        setError(data.error || "Registration failed");
        if (onError) onError(data.error || "Registration failed");
      } else {
        if (onSuccess) {
          onSuccess();
        } else {
          window.location.href = "/auth/login";
        }
      }
    } catch (err) {
      setError("Network error");
      if (onError) onError("Network error");
    } finally {
      setLoading(false);
    }
  };

  const inputStyle = {
    width: "100%",
    padding: "10px 14px",
    border: "1px solid #ccc",
    borderRadius: 8,
    fontSize: 15,
    outline: "none",
    marginTop: 6,
    boxSizing: "border-box",
  };

  const labelStyle = {
    fontWeight: 500,
    fontSize: 14,
    marginBottom: 4,
    display: "block",
  };

  const buttonStyle = {
    width: "100%",
    background: "#f97316", 
    color: "#fff",
    border: "none",
    borderRadius: 8,
    padding: "12px 0",
    fontSize: 16,
    fontWeight: 600,
    cursor: "pointer",
  };

  return (
    <div
      style={{
        maxWidth: 420,
        margin: "2rem auto",
        background: "#fff",
        borderRadius: 12,
        boxShadow: "0 2px 8px rgba(0,0,0,0.07)",
        padding: 32,
      }}
    >
      <h1
        style={{
          textAlign: "center",
          fontSize: "22px",
          fontWeight: 700,
          marginBottom: 6,
        }}
      >
        Create Account
      </h1>
      <p style={{ textAlign: "center", marginBottom: 24, color: "#666" }}>
        Join our community 
      </p>

      <form onSubmit={handleSubmit} encType="multipart/form-data">
        <div style={{ marginBottom: 16 }}>
          <label style={labelStyle}>Nickname (Optional)</label>
          <input
            type="text"
            name="nickname"
            value={form.nickname}
            onChange={handleChange}
            style={inputStyle}
          />
        </div>

        <div style={{ display: "flex", gap: 12, marginBottom: 16 }}>
          <div style={{ flex: 1 }}>
            <label style={labelStyle}>First Name*</label>
            <input
              type="text"
              name="firstName"
              value={form.firstName}
              onChange={handleChange}
              required
              style={inputStyle}
            />
          </div>
          <div style={{ flex: 1 }}>
            <label style={labelStyle}>Last Name*</label>
            <input
              type="text"
              name="lastName"
              value={form.lastName}
              onChange={handleChange}
              required
              style={inputStyle}
            />
          </div>
        </div>

        <div style={{ marginBottom: 16 }}>
          <label style={labelStyle}>Email*</label>
          <input
            type="email"
            name="email"
            value={form.email}
            onChange={handleChange}
            required
            style={inputStyle}
          />
        </div>

        <div style={{ marginBottom: 16 }}>
          <label style={labelStyle}>Date of Birth*</label>
          <input
            type="date"
            name="dateOfBirth"
            value={form.dateOfBirth}
            onChange={handleChange}
            required
            style={inputStyle}
            placeholder="mm/dd/yyyy"
          />
        </div>

        <div style={{ marginBottom: 16 }}>
          <label style={labelStyle}>Profile Picture</label>
          <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
            {preview ? (
              <img
                src={preview}
                alt="Preview"
                style={{
                  width: 48,
                  height: 48,
                  borderRadius: "50%",
                  objectFit: "cover",
                }}
              />
            ) : (
              <div
                style={{
                  width: 48,
                  height: 48,
                  borderRadius: "50%",
                  background: "#eee",
                }}
              />
            )}
            <input
              type="file"
              name="avatar"
              accept="image/*"
              onChange={handleChange}
            />
          </div>
        </div>

        <div style={{ marginBottom: 16 }}>
          <label style={labelStyle}>About Me (Optional)</label>
          <textarea
            name="aboutMe"
            value={form.aboutMe}
            onChange={handleChange}
            style={{ ...inputStyle, minHeight: 60 }}
          />
        </div>

        <div style={{ marginBottom: 16, position: "relative" }}>
          <label style={labelStyle}>Password*</label>
          <input
            type={showPassword ? "text" : "password"}
            name="password"
            value={form.password}
            onChange={handleChange}
            required
            style={{ ...inputStyle, paddingRight: 40 }}
          />
          <span
            onClick={() => setShowPassword(!showPassword)}
            style={{
              position: "absolute",
              right: 12,
              top: 38,
              cursor: "pointer",
              color: "#f97316",
            }}
          >
            {showPassword ? "🙈" : "👁️"}
          </span>
        </div>

        {error && (
          <div style={{ color: "red", marginBottom: 12, textAlign: "center" }}>
            {error}
          </div>
        )}

        <button type="submit" disabled={loading} style={buttonStyle}>
          {loading ? "Registering..." : "Register"}
        </button>
      </form>
      <div style={{ textAlign: "center", marginTop: 16, fontSize: 15 }}>
        Already registered?{" "}
        <Link
          href="/auth/login"
          style={{ color: "#111", fontWeight: 600, textDecoration: "none" }}
        >
          Login
        </Link>
      </div>
    </div>
  );
}
