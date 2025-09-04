"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useAuth } from "../../store/AuthContext";
import { useRouter } from "next/navigation";

export default function LoginForm({ onSuccess, onError }) {
  const { login } = useAuth();
  const router = useRouter();
  const [form, setForm] = useState({ email: "", password: "" });
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setForm((f) => ({ ...f, [name]: value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await login(form.email, form.password);
      if (onSuccess) {
        onSuccess();
      } else {
        router.push("/home");
      }
    } catch (err) {
      setError(err.message || "Wrong email or password");
      if (onError) onError(err.message || "Wrong email or password");
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
        Login
      </h1>
      <p style={{ textAlign: "center", marginBottom: 24, color: "#666" }}>
        Welcome back! Please enter your details.
      </p>

      <form onSubmit={handleSubmit}>
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
          {loading ? "Logging in..." : "Login"}
        </button>
      </form>

      <div style={{ textAlign: "center", marginTop: 16, fontSize: 15 }}>
        Don&apos;t have an account?{" "}
        <Link
          href="/auth/register"
          style={{ color: "#111", fontWeight: 600, textDecoration: "none" }}
        >
          Sign Up
        </Link>
      </div>
    </div>
  );
}
