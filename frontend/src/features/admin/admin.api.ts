import { apiFetch } from "../../api/client";
import type { ScenarioSummary } from "../scenarios/types";
import type {
  AdminScenarioDetail,
  CreateOpportunityInput,
  CreateScenarioInput,
  CreateScenarioLineInput,
  ObjectionOpportunity,
  ObjectionType,
  UpdateOpportunityInput,
  UpdateScenarioInput,
  UpdateScenarioLineInput,
} from "./admin.types";

type ListAdminScenariosResponse = {
  data: ScenarioSummary[];
};

type GetAdminScenarioResponse = {
  data: AdminScenarioDetail;
};

type ScenarioResponse = {
  data: ScenarioSummary;
};

type ScenarioLineResponse = {
  data: unknown;
};

type ListObjectionTypesResponse = {
  data: ObjectionType[];
};

type OpportunityResponse = {
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
): Promise<ScenarioResponse> {
  return apiFetch<ScenarioResponse>("/admin/scenarios", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateAdminScenario(
  scenarioId: string,
  input: UpdateScenarioInput
): Promise<ScenarioResponse> {
  return apiFetch<ScenarioResponse>(`/admin/scenarios/${scenarioId}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function publishAdminScenario(
  scenarioId: string
): Promise<ScenarioResponse> {
  return apiFetch<ScenarioResponse>(
    `/admin/scenarios/${scenarioId}/publish`,
    {
      method: "POST",
    }
  );
}

export function archiveAdminScenario(
  scenarioId: string
): Promise<ScenarioResponse> {
  return apiFetch<ScenarioResponse>(
    `/admin/scenarios/${scenarioId}/archive`,
    {
      method: "POST",
    }
  );
}

export function createAdminScenarioLine(
  scenarioId: string,
  input: CreateScenarioLineInput
): Promise<ScenarioLineResponse> {
  return apiFetch<ScenarioLineResponse>(
    `/admin/scenarios/${scenarioId}/lines`,
    {
      method: "POST",
      body: JSON.stringify(input),
    }
  );
}

export function updateAdminScenarioLine(
  lineId: string,
  input: UpdateScenarioLineInput
): Promise<ScenarioLineResponse> {
  return apiFetch<ScenarioLineResponse>(`/admin/scenario-lines/${lineId}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deleteAdminScenarioLine(lineId: string): Promise<void> {
  return apiFetch<void>(`/admin/scenario-lines/${lineId}`, {
    method: "DELETE",
  });
}

export function listObjectionTypes(): Promise<ListObjectionTypesResponse> {
  return apiFetch<ListObjectionTypesResponse>("/admin/objection-types");
}

export function createLineOpportunity(
  lineId: string,
  input: CreateOpportunityInput
): Promise<OpportunityResponse> {
  return apiFetch<OpportunityResponse>(
    `/admin/scenario-lines/${lineId}/opportunities`,
    {
      method: "POST",
      body: JSON.stringify(input),
    }
  );
}

export function updateLineOpportunity(
  opportunityId: string,
  input: UpdateOpportunityInput
): Promise<OpportunityResponse> {
  return apiFetch<OpportunityResponse>(
    `/admin/opportunities/${opportunityId}`,
    {
      method: "PUT",
      body: JSON.stringify(input),
    }
  );
}

export function deleteLineOpportunity(opportunityId: string): Promise<void> {
  return apiFetch<void>(`/admin/opportunities/${opportunityId}`, {
    method: "DELETE",
  });
}