"use client";

import React, { useState } from 'react';
import Link from 'next/link';

export default function RegisterForm({ onSuccess, onError }) {
  const [form, setForm] = useState({
    nickname: '',
    firstName: '',
    lastName: '',
    email: '',
    dateOfBirth: '',
    avatar: null,
    aboutMe: '',
    password: '',
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleChange = (e) => {
    const { name, value, files } = e.target;
    if (name === 'avatar') {
      setForm(f => ({ ...f, avatar: files[0] }));
    } else {
      setForm(f => ({ ...f, [name]: value }));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const { avatar, ...rest } = form; // ignore avatar for now
      const res = await fetch('http://localhost:8080/api/v1/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(rest),
        credentials: 'include',
      });
      if (!res.ok) {
        const data = await res.json();
        setError(data.error || 'Registration failed');
        if (onError) onError(data.error || 'Registration failed');
      } else {
        if (onSuccess) {
          onSuccess();
        } else {
          window.location.href = '/auth/login';
        }
      }
    } catch (err) {
      setError('Network error');
      if (onError) onError('Network error');
    } finally {
      setLoading(false);
    }
  };

  // Input style to match LoginForm
  const inputStyle = {
    width: '100%',
    padding: '12px 16px',
    border: '2px solid #e0e0e0',
    borderRadius: 8,
    fontSize: 16,
    outline: 'none',
    marginBottom: 8,
    boxSizing: 'border-box',
  };

  // Button style to match LoginForm
  const buttonStyle = {
    width: '100%',
    background: '#b74115',
    color: '#fff',
    border: 'none',
    borderRadius: 999,
    padding: '14px 0',
    fontSize: 18,
    fontWeight: 600,
    cursor: 'pointer',
    marginTop: 8,
    marginBottom: 8,
    boxShadow: '0 2px 4px rgba(180,65,21,0.08)',
    transition: 'background 0.2s',
  };

  return (
    <div style={{ maxWidth: 400, margin: '2rem auto', background: '#fff', borderRadius: 16, boxShadow: '0 2px 8px rgba(0,0,0,0.07)', padding: 32 }}>
      <h1 style={{
        textAlign: 'center',
        color: '#b74115',
        fontSize: '28px',
        fontWeight: '600',
        marginBottom: '24px'
      }}>
        Register
      </h1>
      <form onSubmit={handleSubmit} encType="multipart/form-data">
        <div style={{ marginBottom: 16 }}>
          <input type="text" name="nickname" placeholder="Nickname (Optional)" value={form.nickname} onChange={handleChange} style={inputStyle} />
        </div>
        <div style={{ marginBottom: 16 }}>
          <input type="text" name="firstName" placeholder="First Name" value={form.firstName} onChange={handleChange} required style={inputStyle} />
        </div>
        <div style={{ marginBottom: 16 }}>
          <input type="text" name="lastName" placeholder="Last Name" value={form.lastName} onChange={handleChange} required style={inputStyle} />
        </div>
        <div style={{ marginBottom: 16 }}>
          <input type="email" name="email" placeholder="Email" value={form.email} onChange={handleChange} required style={inputStyle} />
        </div>
        <div style={{ marginBottom: 16 }}>
          <input type="date" name="dateOfBirth" placeholder="Date of Birth" value={form.dateOfBirth} onChange={handleChange} required style={inputStyle} />
        </div>
        <div style={{ marginBottom: 16 }}>
          <input type="file" name="avatar" accept="image/*" onChange={handleChange} style={inputStyle} />
        </div>
        <div style={{ marginBottom: 16 }}>
          <textarea name="aboutMe" placeholder="About Me (Optional)" value={form.aboutMe} onChange={handleChange} style={{ ...inputStyle, minHeight: 60 }} />
        </div>
        <div style={{ marginBottom: 16, position: 'relative' }}>
          <input
            type="password"
            name="password"
            placeholder="Password"
            value={form.password}
            onChange={handleChange}
            required
            style={{ ...inputStyle, paddingRight: 40 }}
          />
          {/* Password icon (lock) */}
          <span style={{
            position: 'absolute',
            right: 12,
            top: '50%',
            transform: 'translateY(-50%)',
            color: '#c44a1b',
            fontSize: 20,
            pointerEvents: 'none',
          }}>
            <svg width="20" height="20" fill="none" viewBox="0 0 24 24"><path fill="#c44a1b" d="M17 10V7a5 5 0 0 0-10 0v3H5v10a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V10h-2Zm-8-3a3 3 0 1 1 6 0v3H9V7Zm8 13H7V12h10v8Zm-5-3a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3Z"/></svg>
          </span>
        </div>
        <div style={{ textAlign: 'center', marginBottom: 24, fontSize: 15 }}>
          Already have an account?{' '}
          <Link href="/auth/login" style={{ color: '#111', fontWeight: 600, textDecoration: 'none' }}>Login</Link>
        </div>
        {error && <div style={{ color: 'red', marginBottom: 8, textAlign: 'center' }}>{error}</div>}
        <button type="submit" disabled={loading} style={buttonStyle}>
          {loading ? 'Registering...' : 'Register'}
        </button>
      </form>
    </div>
  );
} 