import type { ScenarioLine } from "../scenarios/types";

export type SessionMode =
  | "spot_objection"
  | "respond_to_objection"
  | "full_simulation";

export type SessionStatus = "active" | "completed" | "abandoned";

export type SessionEvent = {
  id: string;
  sessionId: string;
  sequenceNo: number;
  eventType:
    | "system_line"
    | "trainee_objection"
    | "trainee_response"
    | "judge_ruling"
    | "coach_feedback"
    | "missed_opportunity";
  actor?: string | null;
  text: string;
  createdAt?: string;
};

export type SessionSummary = {
  id: string;
  scenarioId: string;
  status: SessionStatus;
  currentSequenceNo: number;
  mode: SessionMode;
};

export type SessionDetail = SessionSummary & {
  startedAt?: string;
  completedAt?: string | null;
  events: SessionEvent[];
};

export type CreateSessionInput = {
  scenarioId: string;
  mode?: SessionMode;
};

export type AdvanceSessionResult = {
  session: SessionSummary;
  line: ScenarioLine | null;
  completed: boolean;
};