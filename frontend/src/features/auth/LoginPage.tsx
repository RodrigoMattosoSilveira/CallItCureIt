// frontend/src/features/auth/LoginPage.tsx

import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useMutation } from "@tanstack/react-query";
import { login } from "./auth.api";
import { setAuthSession } from "./auth.store";

export function LoginPage() {
  const navigate = useNavigate();

  const [email, setEmail] = useState("admin@example.com");
  const [password, setPassword] = useState("admin123");

  const loginMutation = useMutation({
    mutationFn: () =>
      login({
        email,
        password,
      }),
    onSuccess: (response) => {
      setAuthSession(response.data);
      navigate("/admin/scenarios");
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    loginMutation.mutate();
  }

  return (
    <div className="container py-4" style={{ maxWidth: 480 }}>
      <h1 className="mb-3">Admin Login</h1>

      <form onSubmit={handleSubmit}>
        <div className="mb-3">
          <label htmlFor="email" className="form-label">
            Email
          </label>

          <input
            id="email"
            type="email"
            className="form-control"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            autoComplete="email"
          />
        </div>

        <div className="mb-3">
          <label htmlFor="password" className="form-label">
            Password
          </label>

          <input
            id="password"
            type="password"
            className="form-control"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            autoComplete="current-password"
          />
        </div>

        {loginMutation.isError && (
          <div className="alert alert-danger">
            Login failed. Check your email and password.
          </div>
        )}

        <button
          type="submit"
          className="btn btn-primary"
          disabled={loginMutation.isPending}
        >
          {loginMutation.isPending ? "Logging in..." : "Login"}
        </button>
      </form>
    </div>
  );
}