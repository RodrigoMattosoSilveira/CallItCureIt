export type ScenarioSummary = {
  id: string;
  title: string;
  description?: string | null;
  jurisdiction: string;
  practiceArea: string;
  hearingType: string;
  difficulty: "beginner" | "intermediate" | "advanced";
  status: "draft" | "published" | "archived";
};

export type ScenarioActor = {
  id: string;
  name: string;
  actorType: "judge" | "witness" | "opposing_counsel" | "trainee_counsel";
  persona?: string | null;
};

export type ScenarioDetail = ScenarioSummary & {
  actors: ScenarioActor[];
};

export type ScenarioLine = {
  id: string;
  scenarioId: string;
  sequenceNo: number;
  speakerType:
    | "judge"
    | "witness"
    | "opposing_counsel"
    | "trainee_counsel"
    | "coach"
    | "system";
  speakerName?: string | null;
  lineText: string;
  lineKind: "question" | "answer" | "argument" | "ruling" | "instruction";
};