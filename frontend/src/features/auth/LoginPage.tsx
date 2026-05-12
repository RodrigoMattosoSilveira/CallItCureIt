import { FormEvent, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useMutation } from "@tanstack/react-query";
import { login } from "./auth.api";
import { setAuthSession } from "./auth.store";

export function LoginPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const [email, setEmail] = useState("admin@example.com");
  const [password, setPassword] = useState("admin123");

  const redirectTo = searchParams.get("redirectTo") || "/admin/scenarios";

  const loginMutation = useMutation({
    mutationFn: login,
    onSuccess: (response) => {
      setAuthSession(response.data.token, response.data.user);
      navigate(redirectTo, { replace: true });
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!email.trim() || !password) {
      return;
    }

    loginMutation.mutate({
      email,
      password,
    });
  }

  return (
    <div className="container py-5" style={{ maxWidth: 480 }}>
      <h1 className="mb-3">Admin Login</h1>

      <form className="card" onSubmit={handleSubmit}>
        <div className="card-body">
          <div className="mb-3">
            <label className="form-label" htmlFor="email">
              Email
            </label>
            <input
              id="email"
              className="form-control"
              type="email"
              autoComplete="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
            />
          </div>

          <div className="mb-3">
            <label className="form-label" htmlFor="password">
              Password
            </label>
            <input
              id="password"
              className="form-control"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
          </div>

          {loginMutation.isError && (
            <div className="alert alert-danger">
              Login failed. Check your email and password.
            </div>
          )}

          <button
            className="btn btn-primary w-100"
            disabled={loginMutation.isPending}
          >
            {loginMutation.isPending ? "Logging in..." : "Login"}
          </button>
        </div>
      </form>
    </div>
  );
}