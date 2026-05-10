import { apiFetch } from "../../api/client";
import type {
  AdvanceSessionResult,
  CreateSessionInput,
  GetSessionDebriefResult,
  GetSessionScoreResult,
  SessionDetail,
  SubmitTraineeActionInput,
  SubmitTraineeActionResult,
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

type SubmitTraineeActionResponse = {
  data: SubmitTraineeActionResult;
};

type GetSessionScoreResponse = {
  data: GetSessionScoreResult;
};

type GetSessionDebriefResponse = {
  data: GetSessionDebriefResult;
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

export function submitTraineeAction(
  sessionId: string,
  input: SubmitTraineeActionInput
): Promise<SubmitTraineeActionResponse> {
  return apiFetch<SubmitTraineeActionResponse>(
    `/sessions/${sessionId}/actions`,
    {
      method: "POST",
      body: JSON.stringify(input),
    }
  );
}

export function getSessionScore(
  sessionId: string
): Promise<GetSessionScoreResponse> {
  return apiFetch<GetSessionScoreResponse>(`/sessions/${sessionId}/score`);
}

export function getSessionDebrief(
  sessionId: string
): Promise<GetSessionDebriefResponse> {
  return apiFetch<GetSessionDebriefResponse>(
    `/sessions/${sessionId}/debrief`
  );
}