import { apiFetch } from "../../api/client";
import type { ScenarioDetail, ScenarioLine, ScenarioSummary } from "./types";

type ListScenariosResponse = {
  data: ScenarioSummary[];
};

type GetScenarioResponse = {
  data: ScenarioDetail;
};

type GetTranscriptResponse = {
  data: ScenarioLine[];
};

export function listScenarios(): Promise<ListScenariosResponse> {
  return apiFetch<ListScenariosResponse>("/scenarios");
}

export function getScenario(
  scenarioId: string
): Promise<GetScenarioResponse> {
  return apiFetch<GetScenarioResponse>(`/scenarios/${scenarioId}`);
}

export function getScenarioTranscript(
  scenarioId: string
): Promise<GetTranscriptResponse> {
  return apiFetch<GetTranscriptResponse>(
    `/scenarios/${scenarioId}/transcript`
  );
}