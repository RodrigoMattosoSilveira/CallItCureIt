import { apiFetch } from "../../api/client";
import type {
  AdvanceSessionResult,
  CreateSessionInput,
  SessionDetail,
} from "./types";

type GetSessionResponse = {
  data: SessionDetail;
};

type CreateSessionResponse = {
  data: SessionDetail;
};

type AdvanceSessionResponse = {
  data: AdvanceSessionResult;
};

export function createSession(
  input: CreateSessionInput
): Promise<CreateSessionResponse> {
  return apiFetch<CreateSessionResponse>("/sessions", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getSession(sessionId: string): Promise<GetSessionResponse> {
  return apiFetch<GetSessionResponse>(`/sessions/${sessionId}`);
}

export function advanceSession(
  sessionId: string
): Promise<AdvanceSessionResponse> {
  return apiFetch<AdvanceSessionResponse>(`/sessions/${sessionId}/next`, {
    method: "POST",
  });
}