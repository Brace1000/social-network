"use client";
import React from "react";
import LogoutForm from "../../components/auth/LogoutForm";

export default function LogoutPage() {
  return (
    <div style={{ maxWidth: 400, margin: "4rem auto", textAlign: "center" }}>
      <h2>Logout</h2>
      <p>Click below to log out of your account.</p>
      <LogoutForm />
    </div>
  );
} 