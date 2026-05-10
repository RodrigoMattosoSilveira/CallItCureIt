import { apiFetch } from "../../api/client";
import type { ScenarioSummary } from "../scenarios/types";
import type {
  AdminScenarioDetail,
  CreateOpportunityInput,
  CreateScenarioInput,
  CreateScenarioLineInput,
  ObjectionOpportunity,
  ObjectionType,
} from "./admin.types";

type ListAdminScenariosResponse = {
  data: ScenarioSummary[];
};

type GetAdminScenarioResponse = {
  data: AdminScenarioDetail;
};

type CreateScenarioResponse = {
  data: ScenarioSummary;
};

type CreateScenarioLineResponse = {
  data: unknown;
};

type ListObjectionTypesResponse = {
  data: ObjectionType[];
};

type CreateOpportunityResponse = {
  data: ObjectionOpportunity;
};

export function listAdminScenarios(): Promise<ListAdminScenariosResponse> {
  return apiFetch<ListAdminScenariosResponse>("/admin/scenarios");
}

export function getAdminScenario(
  scenarioId: string
): Promise<GetAdminScenarioResponse> {
  return apiFetch<GetAdminScenarioResponse>(`/admin/scenarios/${scenarioId}`);
}

export function createAdminScenario(
  input: CreateScenarioInput
): Promise<CreateScenarioResponse> {
  return apiFetch<CreateScenarioResponse>("/admin/scenarios", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function publishAdminScenario(
  scenarioId: string
): Promise<CreateScenarioResponse> {
  return apiFetch<CreateScenarioResponse>(
    `/admin/scenarios/${scenarioId}/publish`,
    {
      method: "POST",
    }
  );
}

export function archiveAdminScenario(
  scenarioId: string
): Promise<CreateScenarioResponse> {
  return apiFetch<CreateScenarioResponse>(
    `/admin/scenarios/${scenarioId}/archive`,
    {
      method: "POST",
    }
  );
}

export function createAdminScenarioLine(
  scenarioId: string,
  input: CreateScenarioLineInput
): Promise<CreateScenarioLineResponse> {
  return apiFetch<CreateScenarioLineResponse>(
    `/admin/scenarios/${scenarioId}/lines`,
    {
      method: "POST",
      body: JSON.stringify(input),
    }
  );
}

export function listObjectionTypes(): Promise<ListObjectionTypesResponse> {
  return apiFetch<ListObjectionTypesResponse>("/admin/objection-types");
}

export function createLineOpportunity(
  lineId: string,
  input: CreateOpportunityInput
): Promise<CreateOpportunityResponse> {
  return apiFetch<CreateOpportunityResponse>(
    `/admin/scenario-lines/${lineId}/opportunities`,
    {
      method: "POST",
      body: JSON.stringify(input),
    }
  );
}