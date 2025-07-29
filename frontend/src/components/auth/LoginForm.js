"use client";

import React, { useState } from 'react';
import Link from 'next/link';
import { useAuth } from '../../store/AuthContext';
import { useRouter } from 'next/navigation';

export default function LoginForm({ onSuccess, onError }) {
  const { login } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await login(email, password);
      if (onSuccess) {
        onSuccess();
      } else {
        router.push('/home');
      }
    } catch (err) {
      setError(err.message || 'Wrong email or password');
      if (onError) onError(err.message || 'Wrong email or password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: 400, margin: '2rem auto', background: '#fff', borderRadius: 16, boxShadow: '0 2px 8px rgba(0,0,0,0.07)', padding: 32 }}>
      {/* Added Login Title */}
      <h1 style={{
        textAlign: 'center',
        color: '#b74115',
        fontSize: '28px',
        fontWeight: '600',
        marginBottom: '24px'
      }}>
        Login
      </h1>
      
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: 16 }}>
          <input
            type="text"
            placeholder="Enter your email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            required
            style={{
              width: '100%',
              padding: '12px 16px',
              border: '2px solid #c44a1b',
              borderRadius: 8,
              fontSize: 16,
              outline: 'none',
              marginBottom: 8,
              boxSizing: 'border-box',
            }}
          />
        </div>
        <div style={{ marginBottom: 16, position: 'relative' }}>
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            required
            style={{
              width: '100%',
              padding: '12px 40px 12px 16px',
              border: '2px solid #e0e0e0',
              borderRadius: 8,
              fontSize: 16,
              outline: 'none',
              boxSizing: 'border-box',
            }}
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
          Don&apos;t have an account?{' '}
          <Link href="/auth/register" style={{ color: '#111', fontWeight: 600, textDecoration: 'none' }}>Sign Up</Link>
        </div>
        {error && <div style={{ color: 'red', marginBottom: 8, textAlign: 'center' }}>{error}</div>}
        <button
          type="submit"
          disabled={loading}
          style={{
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
          }}
        >
          {loading ? 'Logging in...' : 'Login'}
        </button>
      </form>
    </div>
  );
}