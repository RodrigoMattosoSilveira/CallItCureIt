import type { ScenarioLine } from "../scenarios/types";

export type SessionMode =
  | "spot_objection"
  | "respond_to_objection"
  | "full_simulation";

export type SessionStatus = "active" | "completed" | "abandoned";

export type SessionEventType =
  | "system_line"
  | "trainee_objection"
  | "trainee_response"
  | "judge_ruling"
  | "coach_feedback"
  | "missed_opportunity";

export type SessionEvent = {
  id: string;
  sessionId: string;
  sequenceNo: number;
  eventType: SessionEventType;
  actor?: string | null;
  text: string;
  createdAt?: string | null;
};

export type SessionSummary = {
  id: string;
  scenarioId: string;
  status: SessionStatus;
  currentSequenceNo: number;
  mode: SessionMode;
};

export type SessionDetail = SessionSummary & {
  startedAt?: string | null;
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

export type TraineeActionType = "object" | "respond" | "pass";

export type TraineeAction = {
  id: string;
  sessionId: string;
  scenarioLineId?: string | null;
  actionType: TraineeActionType;
  rawText: string;
  normalizedObjectionTypeId?: string | null;
  createdAt?: string | null;
};

export type SubmitTraineeActionInput = {
  actionType: TraineeActionType;
  rawText: string;
};

export type ActionEvaluation = {
  id: string;
  traineeActionId: string;
  matchedOpportunityId?: string | null;
  normalizedObjectionTypeId?: string | null;
  valid: boolean;
  timely: boolean;
  ruling: "sustained" | "overruled" | "no_ruling";
  legalAccuracyScore: number;
  phrasingScore: number;
  strategyScore: number;
  feedback: string;
  createdAt?: string | null;
};

export type SubmitTraineeActionResult = {
  session: SessionSummary;
  action: TraineeAction;
  traineeEvent: SessionEvent;
  judgeEvent: SessionEvent;
  coachEvent: SessionEvent;
  evaluation: ActionEvaluation;
};