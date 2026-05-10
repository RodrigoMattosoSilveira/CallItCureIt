import { FormEvent, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  archiveAdminScenario,
  createAdminScenarioLine,
  createLineOpportunity,
  getAdminScenario,
  listObjectionTypes,
  publishAdminScenario,
} from "./admin.api";
import type {
  CreateOpportunityInput,
  CreateScenarioLineInput,
} from "./admin.types";

const initialLineForm: CreateScenarioLineInput = {
  sequenceNo: 1,
  speakerType: "opposing_counsel",
  speakerName: "",
  lineText: "",
  lineKind: "question",
};

const initialOpportunityForm: CreateOpportunityInput = {
  objectionTypeId: "",
  strength: "strong",
  timingWindow: "after_answer",
  explanation: "",
  expectedPhrase: "",
  isPrimary: true,
};

export function AdminScenarioDetailPage() {
  const { scenarioId } = useParams<{ scenarioId: string }>();
  const queryClient = useQueryClient();

  const [lineForm, setLineForm] =
    useState<CreateScenarioLineInput>(initialLineForm);

  const [selectedLineId, setSelectedLineId] = useState<string>("");
  const [opportunityForm, setOpportunityForm] =
    useState<CreateOpportunityInput>(initialOpportunityForm);

  const scenarioQuery = useQuery({
    queryKey: ["admin-scenario", scenarioId],
    queryFn: () => getAdminScenario(scenarioId!),
    enabled: Boolean(scenarioId),
  });

  const objectionTypesQuery = useQuery({
    queryKey: ["admin-objection-types"],
    queryFn: listObjectionTypes,
  });

  const publishMutation = useMutation({
    mutationFn: () => publishAdminScenario(scenarioId!),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["admin-scenario", scenarioId],
      });
      queryClient.invalidateQueries({
        queryKey: ["admin-scenarios"],
      });
    },
  });

  const archiveMutation = useMutation({
    mutationFn: () => archiveAdminScenario(scenarioId!),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["admin-scenario", scenarioId],
      });
      queryClient.invalidateQueries({
        queryKey: ["admin-scenarios"],
      });
    },
  });

  const createLineMutation = useMutation({
    mutationFn: () => createAdminScenarioLine(scenarioId!, lineForm),
    onSuccess: () => {
      setLineForm((current) => ({
        ...initialLineForm,
        sequenceNo: current.sequenceNo + 1,
      }));

      queryClient.invalidateQueries({
        queryKey: ["admin-scenario", scenarioId],
      });
    },
  });

  const createOpportunityMutation = useMutation({
    mutationFn: () => createLineOpportunity(selectedLineId, opportunityForm),
    onSuccess: () => {
      setOpportunityForm(initialOpportunityForm);
      setSelectedLineId("");

      queryClient.invalidateQueries({
        queryKey: ["admin-scenario", scenarioId],
      });
    },
  });

  if (scenarioQuery.isLoading) {
    return <div className="container py-4">Loading scenario...</div>;
  }

  if (scenarioQuery.isError || !scenarioQuery.data || !scenarioId) {
    return (
      <div className="container py-4">
        <div className="alert alert-danger">Failed to load scenario.</div>
      </div>
    );
  }

  const detail = scenarioQuery.data.data;
  const scenario = detail.scenario;
  const lines = detail.lines;
  const objectionTypes = objectionTypesQuery.data?.data ?? [];

  function handleCreateLine(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!lineForm.lineText.trim()) {
      return;
    }

    createLineMutation.mutate();
  }

  function handleCreateOpportunity(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (
      !selectedLineId ||
      !opportunityForm.objectionTypeId ||
      !opportunityForm.explanation.trim()
    ) {
      return;
    }

    createOpportunityMutation.mutate();
  }

  return (
    <div className="container py-4">
      <Link to="/admin/scenarios" className="btn btn-link px-0">
        ← Back to admin scenarios
      </Link>

      <div className="d-flex justify-content-between align-items-start gap-3 mb-4">
        <div>
          <h1>{scenario.title}</h1>
          <p className="text-muted mb-1">{scenario.description}</p>
          <div className="small text-muted">
            {scenario.jurisdiction} · {scenario.practiceArea} ·{" "}
            {scenario.hearingType}
          </div>
        </div>

        <div className="text-end">
          <div className="mb-2">
            <span className="badge text-bg-primary">{scenario.status}</span>
          </div>

          <div className="d-flex gap-2">
            <button
              className="btn btn-success btn-sm"
              onClick={() => publishMutation.mutate()}
              disabled={publishMutation.isPending}
            >
              Publish
            </button>

            <button
              className="btn btn-outline-secondary btn-sm"
              onClick={() => archiveMutation.mutate()}
              disabled={archiveMutation.isPending}
            >
              Archive
            </button>
          </div>
        </div>
      </div>

      <div className="row g-4">
        <div className="col-12 col-lg-7">
          <div className="card mb-4">
            <div className="card-header">Transcript Lines</div>
            <div className="card-body">
              {lines.length === 0 && (
                <div className="text-muted">No transcript lines yet.</div>
              )}

              {lines.map((line) => (
                <div className="border rounded p-3 mb-3" key={line.id}>
                  <div className="fw-bold">
                    {line.sequenceNo}. {line.speakerName || line.speakerType}
                  </div>
                  <div>{line.lineText}</div>
                  <div className="small text-muted mt-1">
                    {line.lineKind} · {line.speakerType}
                  </div>

                  {line.opportunities.length > 0 && (
                    <div className="mt-3">
                      <div className="small fw-bold">Opportunities</div>
                      {line.opportunities.map((opportunity) => (
                        <div
                          className="small border-start border-4 ps-2 mt-2"
                          key={opportunity.id}
                        >
                          {opportunity.objectionTypeId} ·{" "}
                          {opportunity.strength} ·{" "}
                          {opportunity.timingWindow}
                          <div>{opportunity.explanation}</div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          <form className="card" onSubmit={handleCreateLine}>
            <div className="card-header">Add Transcript Line</div>
            <div className="card-body">
              <div className="row g-3">
                <div className="col-4">
                  <label className="form-label" htmlFor="sequenceNo">
                    Sequence
                  </label>
                  <input
                    id="sequenceNo"
                    type="number"
                    className="form-control"
                    value={lineForm.sequenceNo}
                    onChange={(event) =>
                      setLineForm((current) => ({
                        ...current,
                        sequenceNo: Number(event.target.value),
                      }))
                    }
                  />
                </div>

                <div className="col-8">
                  <label className="form-label" htmlFor="speakerName">
                    Speaker Name
                  </label>
                  <input
                    id="speakerName"
                    className="form-control"
                    value={lineForm.speakerName}
                    onChange={(event) =>
                      setLineForm((current) => ({
                        ...current,
                        speakerName: event.target.value,
                      }))
                    }
                  />
                </div>
              </div>

              <div className="row g-3 mt-1">
                <div className="col-6">
                  <label className="form-label" htmlFor="speakerType">
                    Speaker Type
                  </label>
                  <select
                    id="speakerType"
                    className="form-select"
                    value={lineForm.speakerType}
                    onChange={(event) =>
                      setLineForm((current) => ({
                        ...current,
                        speakerType:
                          event.target
                            .value as CreateScenarioLineInput["speakerType"],
                      }))
                    }
                  >
                    <option value="judge">Judge</option>
                    <option value="witness">Witness</option>
                    <option value="opposing_counsel">Opposing Counsel</option>
                    <option value="trainee_counsel">Trainee Counsel</option>
                    <option value="coach">Coach</option>
                    <option value="system">System</option>
                  </select>
                </div>

                <div className="col-6">
                  <label className="form-label" htmlFor="lineKind">
                    Line Kind
                  </label>
                  <select
                    id="lineKind"
                    className="form-select"
                    value={lineForm.lineKind}
                    onChange={(event) =>
                      setLineForm((current) => ({
                        ...current,
                        lineKind:
                          event.target.value as CreateScenarioLineInput["lineKind"],
                      }))
                    }
                  >
                    <option value="question">Question</option>
                    <option value="answer">Answer</option>
                    <option value="argument">Argument</option>
                    <option value="ruling">Ruling</option>
                    <option value="instruction">Instruction</option>
                  </select>
                </div>
              </div>

              <div className="mt-3">
                <label className="form-label" htmlFor="lineText">
                  Line Text
                </label>
                <textarea
                  id="lineText"
                  className="form-control"
                  rows={3}
                  value={lineForm.lineText}
                  onChange={(event) =>
                    setLineForm((current) => ({
                      ...current,
                      lineText: event.target.value,
                    }))
                  }
                />
              </div>

              {createLineMutation.isError && (
                <div className="alert alert-danger mt-3">
                  Failed to create line.
                </div>
              )}

              <button
                className="btn btn-primary mt-3"
                disabled={createLineMutation.isPending}
              >
                {createLineMutation.isPending ? "Adding..." : "Add Line"}
              </button>
            </div>
          </form>
        </div>

        <div className="col-12 col-lg-5">
          <form className="card" onSubmit={handleCreateOpportunity}>
            <div className="card-header">Add Objection Opportunity</div>
            <div className="card-body">
              <div className="mb-3">
                <label className="form-label" htmlFor="lineId">
                  Transcript Line
                </label>
                <select
                  id="lineId"
                  className="form-select"
                  value={selectedLineId}
                  onChange={(event) => setSelectedLineId(event.target.value)}
                >
                  <option value="">Select line...</option>
                  {lines.map((line) => (
                    <option key={line.id} value={line.id}>
                      {line.sequenceNo}. {line.lineText.slice(0, 80)}
                    </option>
                  ))}
                </select>
              </div>

              <div className="mb-3">
                <label className="form-label" htmlFor="objectionTypeId">
                  Objection Type
                </label>
                <select
                  id="objectionTypeId"
                  className="form-select"
                  value={opportunityForm.objectionTypeId}
                  onChange={(event) =>
                    setOpportunityForm((current) => ({
                      ...current,
                      objectionTypeId: event.target.value,
                    }))
                  }
                >
                  <option value="">Select objection...</option>
                  {objectionTypes.map((type) => (
                    <option key={type.id} value={type.id}>
                      {type.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="row g-3">
                <div className="col-6">
                  <label className="form-label" htmlFor="strength">
                    Strength
                  </label>
                  <select
                    id="strength"
                    className="form-select"
                    value={opportunityForm.strength}
                    onChange={(event) =>
                      setOpportunityForm((current) => ({
                        ...current,
                        strength:
                          event.target.value as CreateOpportunityInput["strength"],
                      }))
                    }
                  >
                    <option value="weak">Weak</option>
                    <option value="moderate">Moderate</option>
                    <option value="strong">Strong</option>
                  </select>
                </div>

                <div className="col-6">
                  <label className="form-label" htmlFor="timingWindow">
                    Timing
                  </label>
                  <select
                    id="timingWindow"
                    className="form-select"
                    value={opportunityForm.timingWindow}
                    onChange={(event) =>
                      setOpportunityForm((current) => ({
                        ...current,
                        timingWindow:
                          event.target
                            .value as CreateOpportunityInput["timingWindow"],
                      }))
                    }
                  >
                    <option value="after_question">After Question</option>
                    <option value="after_answer">After Answer</option>
                    <option value="before_answer">Before Answer</option>
                  </select>
                </div>
              </div>

              <div className="mt-3">
                <label className="form-label" htmlFor="expectedPhrase">
                  Expected Phrase
                </label>
                <input
                  id="expectedPhrase"
                  className="form-control"
                  value={opportunityForm.expectedPhrase}
                  onChange={(event) =>
                    setOpportunityForm((current) => ({
                      ...current,
                      expectedPhrase: event.target.value,
                    }))
                  }
                  placeholder="Objection, hearsay."
                />
              </div>

              <div className="mt-3">
                <label className="form-label" htmlFor="explanation">
                  Explanation
                </label>
                <textarea
                  id="explanation"
                  className="form-control"
                  rows={4}
                  value={opportunityForm.explanation}
                  onChange={(event) =>
                    setOpportunityForm((current) => ({
                      ...current,
                      explanation: event.target.value,
                    }))
                  }
                />
              </div>

              <div className="form-check mt-3">
                <input
                  id="isPrimary"
                  className="form-check-input"
                  type="checkbox"
                  checked={opportunityForm.isPrimary}
                  onChange={(event) =>
                    setOpportunityForm((current) => ({
                      ...current,
                      isPrimary: event.target.checked,
                    }))
                  }
                />
                <label className="form-check-label" htmlFor="isPrimary">
                  Primary objection opportunity
                </label>
              </div>

              {createOpportunityMutation.isError && (
                <div className="alert alert-danger mt-3">
                  Failed to create objection opportunity.
                </div>
              )}

              <button
                className="btn btn-warning mt-3"
                disabled={createOpportunityMutation.isPending}
              >
                {createOpportunityMutation.isPending
                  ? "Adding..."
                  : "Add Opportunity"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}