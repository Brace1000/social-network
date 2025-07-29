"use client";
import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "../../store/AuthContext";

export default function LogoutForm() {
  const { logout } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  const handleLogout = async (e) => {
    e.preventDefault();
    setLoading(true);
    try {
      await logout();
      router.push("/auth/login");
    } catch (err) {
      alert("Logout failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleLogout} style={{ display: "inline" }}>
      <button type="submit" disabled={loading} style={{ marginTop: 16, background: "rgba(255, 255, 255, 0.2)", color: "#fff", border: "1px solid rgba(255, 255, 255, 0.3)", borderRadius: 999, padding: "8px 24px", fontWeight: 600, cursor: loading ? "not-allowed" : "pointer", opacity: loading ? 0.7 : 1 }}>
        {loading ? "Logging out..." : "Logout"}
      </button>
    </form>
  );
} 