import type { ScenarioLine, ScenarioSummary } from "../scenarios/types";

export type ObjectionType = {
  id: string;
  code: string;
  name: string;
  description: string;
  defaultPhrase: string;
};

export type ObjectionOpportunity = {
  id: string;
  scenarioLineId: string;
  objectionTypeId: string;
  strength: "weak" | "moderate" | "strong";
  timingWindow: "after_question" | "after_answer" | "before_answer";
  explanation: string;
  expectedPhrase?: string | null;
  isPrimary: boolean;
};

export type AdminScenarioLine = ScenarioLine & {
  opportunities: ObjectionOpportunity[];
};

export type AdminScenarioDetail = {
  scenario: ScenarioSummary;
  lines: AdminScenarioLine[];
};

export type CreateScenarioInput = {
  title: string;
  description: string;
  jurisdiction: string;
  practiceArea: string;
  hearingType: string;
  difficulty: "beginner" | "intermediate" | "advanced";
  status: "draft" | "published" | "archived";
};

export type UpdateScenarioInput = CreateScenarioInput;

export type CreateScenarioLineInput = {
  sequenceNo: number;
  speakerType:
    | "judge"
    | "witness"
    | "opposing_counsel"
    | "trainee_counsel"
    | "coach"
    | "system";
  speakerName: string;
  lineText: string;
  lineKind: "question" | "answer" | "argument" | "ruling" | "instruction";
};

export type UpdateScenarioLineInput = CreateScenarioLineInput;

export type CreateOpportunityInput = {
  objectionTypeId: string;
  strength: "weak" | "moderate" | "strong";
  timingWindow: "after_question" | "after_answer" | "before_answer";
  explanation: string;
  expectedPhrase: string;
  isPrimary: boolean;
};

export type UpdateOpportunityInput = CreateOpportunityInput;