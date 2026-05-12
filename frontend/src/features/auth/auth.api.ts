import { apiFetch } from "../../api/client";
import type { AuthUser, LoginInput, LoginResult } from "./auth.types";

type LoginResponse = {
  data: LoginResult;
};

type MeResponse = {
  data: AuthUser;
};

export function login(input: LoginInput): Promise<LoginResponse> {
  return apiFetch<LoginResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getMe(): Promise<MeResponse> {
  return apiFetch<MeResponse>("/auth/me");
}

export function logout(): Promise<{ data: { ok: boolean } }> {
  return apiFetch<{ data: { ok: boolean } }>("/auth/logout", {
    method: "POST",
  });
}