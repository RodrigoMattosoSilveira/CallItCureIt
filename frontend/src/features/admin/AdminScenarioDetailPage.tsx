import { FormEvent, useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  archiveAdminScenario,
  createAdminScenarioLine,
  createLineOpportunity,
  deleteAdminScenarioLine,
  deleteLineOpportunity,
  getAdminScenario,
  listObjectionTypes,
  publishAdminScenario,
  updateAdminScenario,
  updateAdminScenarioLine,
  updateLineOpportunity,
} from "./admin.api";
import type {
  AdminScenarioLine,
  CreateOpportunityInput,
  CreateScenarioLineInput,
  ObjectionOpportunity,
  UpdateScenarioInput,
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
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [metadataForm, setMetadataForm] =
    useState<UpdateScenarioInput | null>(null);

  const [lineForm, setLineForm] =
    useState<CreateScenarioLineInput>(initialLineForm);

  const [editingLineId, setEditingLineId] = useState<string | null>(null);
  const [editingLineForm, setEditingLineForm] =
    useState<CreateScenarioLineInput>(initialLineForm);

  const [selectedLineId, setSelectedLineId] = useState<string>("");
  const [opportunityForm, setOpportunityForm] =
    useState<CreateOpportunityInput>(initialOpportunityForm);

  const [editingOpportunityId, setEditingOpportunityId] =
    useState<string | null>(null);
  const [editingOpportunityForm, setEditingOpportunityForm] =
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

  const detail = scenarioQuery.data?.data;
  const scenario = detail?.scenario;
  const lines = useMemo(() => detail?.lines ?? [], [detail?.lines]);
  
  const objectionTypes = useMemo(() => {
    return objectionTypesQuery.data?.data ?? [];
  }, [objectionTypesQuery.data?.data]);

  const objectionTypeNameById = useMemo(() => {
    return new Map(objectionTypes.map((type) => [type.id, type.name]));
  }, [objectionTypes]);

  useEffect(() => {
    if (!scenario) {
      return;
    }

    setMetadataForm({
      title: scenario.title,
      description: scenario.description ?? "",
      jurisdiction: scenario.jurisdiction,
      practiceArea: scenario.practiceArea,
      hearingType: scenario.hearingType,
      difficulty: scenario.difficulty,
      status: scenario.status,
    });
  }, [scenario]);

  useEffect(() => {
    if (lines.length === 0) {
      setLineForm((current) => ({
        ...current,
        sequenceNo: 1,
      }));
      return;
    }

    const maxSequence = Math.max(...lines.map((line) => line.sequenceNo));
    setLineForm((current) => ({
      ...current,
      sequenceNo: maxSequence + 1,
    }));
  }, [lines]);

  function invalidateScenario() {
    queryClient.invalidateQueries({
      queryKey: ["admin-scenario", scenarioId],
    });
    queryClient.invalidateQueries({
      queryKey: ["admin-scenarios"],
    });
  }

  const updateMetadataMutation = useMutation({
    mutationFn: () => updateAdminScenario(scenarioId!, metadataForm!),
    onSuccess: invalidateScenario,
  });

  const publishMutation = useMutation({
    mutationFn: () => publishAdminScenario(scenarioId!),
    onSuccess: invalidateScenario,
  });

  const archiveMutation = useMutation({
    mutationFn: () => archiveAdminScenario(scenarioId!),
    onSuccess: invalidateScenario,
  });

  const createLineMutation = useMutation({
    mutationFn: () => createAdminScenarioLine(scenarioId!, lineForm),
    onSuccess: () => {
      setLineForm((current) => ({
        ...initialLineForm,
        sequenceNo: current.sequenceNo + 1,
      }));
      invalidateScenario();
    },
  });

  const updateLineMutation = useMutation({
    mutationFn: () => updateAdminScenarioLine(editingLineId!, editingLineForm),
    onSuccess: () => {
      setEditingLineId(null);
      invalidateScenario();
    },
  });

  const deleteLineMutation = useMutation({
    mutationFn: deleteAdminScenarioLine,
    onSuccess: invalidateScenario,
  });

  const createOpportunityMutation = useMutation({
    mutationFn: () => createLineOpportunity(selectedLineId, opportunityForm),
    onSuccess: () => {
      setOpportunityForm(initialOpportunityForm);
      setSelectedLineId("");
      invalidateScenario();
    },
  });

  const updateOpportunityMutation = useMutation({
    mutationFn: () =>
      updateLineOpportunity(editingOpportunityId!, editingOpportunityForm),
    onSuccess: () => {
      setEditingOpportunityId(null);
      invalidateScenario();
    },
  });

  const deleteOpportunityMutation = useMutation({
    mutationFn: deleteLineOpportunity,
    onSuccess: invalidateScenario,
  });

  function handleUpdateMetadata(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!metadataForm?.title.trim()) {
      return;
    }

    updateMetadataMutation.mutate();
  }

  function handleCreateLine(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!lineForm.lineText.trim()) {
      return;
    }

    createLineMutation.mutate();
  }

  function startEditingLine(line: AdminScenarioLine) {
    setEditingLineId(line.id);
    setEditingLineForm({
      sequenceNo: line.sequenceNo,
      speakerType: line.speakerType,
      speakerName: line.speakerName ?? "",
      lineText: line.lineText,
      lineKind: line.lineKind,
    });
  }

  function handleUpdateLine(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!editingLineId || !editingLineForm.lineText.trim()) {
      return;
    }

    updateLineMutation.mutate();
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

  function startEditingOpportunity(opportunity: ObjectionOpportunity) {
    setEditingOpportunityId(opportunity.id);
    setEditingOpportunityForm({
      objectionTypeId: opportunity.objectionTypeId,
      strength: opportunity.strength,
      timingWindow: opportunity.timingWindow,
      explanation: opportunity.explanation,
      expectedPhrase: opportunity.expectedPhrase ?? "",
      isPrimary: opportunity.isPrimary,
    });
  }

  function handleUpdateOpportunity(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (
      !editingOpportunityId ||
      !editingOpportunityForm.objectionTypeId ||
      !editingOpportunityForm.explanation.trim()
    ) {
      return;
    }

    updateOpportunityMutation.mutate();
  }

  if (scenarioQuery.isLoading) {
    return <div className="container py-4">Loading scenario...</div>;
  }

  if (scenarioQuery.isError || !scenario || !metadataForm || !scenarioId) {
    return (
      <div className="container py-4">
        <div className="alert alert-danger">Failed to load scenario.</div>
      </div>
    );
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

          <div className="d-flex flex-wrap gap-2 justify-content-end">
            <button
              className="btn btn-outline-primary btn-sm"
              onClick={() => navigate(`/scenarios/${scenario.id}`)}
            >
              Preview as Trainee
            </button>

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

      <form className="card mb-4" onSubmit={handleUpdateMetadata}>
        <div className="card-header">Scenario Metadata</div>
        <div className="card-body">
          <div className="mb-3">
            <label className="form-label" htmlFor="title">
              Title
            </label>
            <input
              id="title"
              className="form-control"
              value={metadataForm.title}
              onChange={(event) =>
                setMetadataForm((current) => ({
                  ...current!,
                  title: event.target.value,
                }))
              }
            />
          </div>

          <div className="mb-3">
            <label className="form-label" htmlFor="description">
              Description
            </label>
            <textarea
              id="description"
              className="form-control"
              rows={2}
              value={metadataForm.description}
              onChange={(event) =>
                setMetadataForm((current) => ({
                  ...current!,
                  description: event.target.value,
                }))
              }
            />
          </div>

          <div className="row g-3">
            <div className="col-12 col-md-4">
              <label className="form-label" htmlFor="jurisdiction">
                Jurisdiction
              </label>
              <input
                id="jurisdiction"
                className="form-control"
                value={metadataForm.jurisdiction}
                onChange={(event) =>
                  setMetadataForm((current) => ({
                    ...current!,
                    jurisdiction: event.target.value,
                  }))
                }
              />
            </div>

            <div className="col-12 col-md-4">
              <label className="form-label" htmlFor="practiceArea">
                Practice Area
              </label>
              <input
                id="practiceArea"
                className="form-control"
                value={metadataForm.practiceArea}
                onChange={(event) =>
                  setMetadataForm((current) => ({
                    ...current!,
                    practiceArea: event.target.value,
                  }))
                }
              />
            </div>

            <div className="col-12 col-md-4">
              <label className="form-label" htmlFor="hearingType">
                Hearing Type
              </label>
              <input
                id="hearingType"
                className="form-control"
                value={metadataForm.hearingType}
                onChange={(event) =>
                  setMetadataForm((current) => ({
                    ...current!,
                    hearingType: event.target.value,
                  }))
                }
              />
            </div>
          </div>

          <div className="row g-3 mt-1">
            <div className="col-12 col-md-6">
              <label className="form-label" htmlFor="difficulty">
                Difficulty
              </label>
              <select
                id="difficulty"
                className="form-select"
                value={metadataForm.difficulty}
                onChange={(event) =>
                  setMetadataForm((current) => ({
                    ...current!,
                    difficulty: event.target
                      .value as UpdateScenarioInput["difficulty"],
                  }))
                }
              >
                <option value="beginner">Beginner</option>
                <option value="intermediate">Intermediate</option>
                <option value="advanced">Advanced</option>
              </select>
            </div>

            <div className="col-12 col-md-6">
              <label className="form-label" htmlFor="status">
                Status
              </label>
              <select
                id="status"
                className="form-select"
                value={metadataForm.status}
                onChange={(event) =>
                  setMetadataForm((current) => ({
                    ...current!,
                    status: event.target.value as UpdateScenarioInput["status"],
                  }))
                }
              >
                <option value="draft">Draft</option>
                <option value="published">Published</option>
                <option value="archived">Archived</option>
              </select>
            </div>
          </div>

          <button
            className="btn btn-primary mt-3"
            disabled={updateMetadataMutation.isPending}
          >
            {updateMetadataMutation.isPending
              ? "Saving..."
              : "Save Metadata"}
          </button>
        </div>
      </form>

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
                  {editingLineId === line.id ? (
                    <form onSubmit={handleUpdateLine}>
                      <LineFields
                        value={editingLineForm}
                        onChange={setEditingLineForm}
                      />

                      <div className="d-flex gap-2 mt-3">
                        <button
                          className="btn btn-primary btn-sm"
                          disabled={updateLineMutation.isPending}
                        >
                          Save Line
                        </button>
                        <button
                          type="button"
                          className="btn btn-light btn-sm"
                          onClick={() => setEditingLineId(null)}
                        >
                          Cancel
                        </button>
                      </div>
                    </form>
                  ) : (
                    <>
                      <div className="d-flex justify-content-between gap-3">
                        <div>
                          <div className="fw-bold">
                            {line.sequenceNo}.{" "}
                            {line.speakerName || line.speakerType}
                          </div>
                          <div>{line.lineText}</div>
                          <div className="small text-muted mt-1">
                            {line.lineKind} · {line.speakerType}
                          </div>
                        </div>

                        <div className="d-flex gap-2 align-self-start">
                          <button
                            className="btn btn-outline-primary btn-sm"
                            onClick={() => startEditingLine(line)}
                          >
                            Edit
                          </button>
                          <button
                            className="btn btn-outline-danger btn-sm"
                            onClick={() => deleteLineMutation.mutate(line.id)}
                          >
                            Delete
                          </button>
                        </div>
                      </div>

                      {(line.opportunities ?? []).length > 0 && (
                        <div className="mt-3">
                          <div className="small fw-bold">Opportunities</div>

                          {(line.opportunities ?? []).map((opportunity) => (
                            <div
                              className="small border-start border-4 ps-2 mt-2"
                              key={opportunity.id}
                            >
                              {editingOpportunityId === opportunity.id ? (
                                <form onSubmit={handleUpdateOpportunity}>
                                  <OpportunityFields
                                    value={editingOpportunityForm}
                                    objectionTypes={objectionTypes}
                                    onChange={setEditingOpportunityForm}
                                  />

                                  <div className="d-flex gap-2 mt-2">
                                    <button
                                      className="btn btn-primary btn-sm"
                                      disabled={
                                        updateOpportunityMutation.isPending
                                      }
                                    >
                                      Save Opportunity
                                    </button>
                                    <button
                                      type="button"
                                      className="btn btn-light btn-sm"
                                      onClick={() =>
                                        setEditingOpportunityId(null)
                                      }
                                    >
                                      Cancel
                                    </button>
                                  </div>
                                </form>
                              ) : (
                                <>
                                  <div className="d-flex justify-content-between gap-2">
                                    <div>
                                      <strong>
                                        {objectionTypeNameById.get(
                                          opportunity.objectionTypeId
                                        ) ?? opportunity.objectionTypeId}
                                      </strong>{" "}
                                      · {opportunity.strength} ·{" "}
                                      {opportunity.timingWindow}
                                    </div>

                                    <div className="d-flex gap-2">
                                      <button
                                        className="btn btn-outline-primary btn-sm"
                                        onClick={() =>
                                          startEditingOpportunity(opportunity)
                                        }
                                      >
                                        Edit
                                      </button>
                                      <button
                                        className="btn btn-outline-danger btn-sm"
                                        onClick={() =>
                                          deleteOpportunityMutation.mutate(
                                            opportunity.id
                                          )
                                        }
                                      >
                                        Delete
                                      </button>
                                    </div>
                                  </div>

                                  <div>{opportunity.explanation}</div>

                                  {opportunity.expectedPhrase && (
                                    <div className="text-muted">
                                      Expected: {opportunity.expectedPhrase}
                                    </div>
                                  )}
                                </>
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                    </>
                  )}
                </div>
              ))}
            </div>
          </div>

          <form className="card" onSubmit={handleCreateLine}>
            <div className="card-header">Add Transcript Line</div>
            <div className="card-body">
              <LineFields value={lineForm} onChange={setLineForm} />

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

              <OpportunityFields
                value={opportunityForm}
                objectionTypes={objectionTypes}
                onChange={setOpportunityForm}
              />

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

function LineFields({
  value,
  onChange,
}: {
  value: CreateScenarioLineInput;
  onChange: (value: CreateScenarioLineInput) => void;
}) {
  return (
    <>
      <div className="row g-3">
        <div className="col-4">
          <label className="form-label">Sequence</label>
          <input
            type="number"
            className="form-control"
            value={value.sequenceNo}
            onChange={(event) =>
              onChange({
                ...value,
                sequenceNo: Number(event.target.value),
              })
            }
          />
        </div>

        <div className="col-8">
          <label className="form-label">Speaker Name</label>
          <input
            className="form-control"
            value={value.speakerName}
            onChange={(event) =>
              onChange({
                ...value,
                speakerName: event.target.value,
              })
            }
          />
        </div>
      </div>

      <div className="row g-3 mt-1">
        <div className="col-6">
          <label className="form-label">Speaker Type</label>
          <select
            className="form-select"
            value={value.speakerType}
            onChange={(event) =>
              onChange({
                ...value,
                speakerType: event.target
                  .value as CreateScenarioLineInput["speakerType"],
              })
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
          <label className="form-label">Line Kind</label>
          <select
            className="form-select"
            value={value.lineKind}
            onChange={(event) =>
              onChange({
                ...value,
                lineKind: event.target
                  .value as CreateScenarioLineInput["lineKind"],
              })
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
        <label className="form-label">Line Text</label>
        <textarea
          className="form-control"
          rows={3}
          value={value.lineText}
          onChange={(event) =>
            onChange({
              ...value,
              lineText: event.target.value,
            })
          }
        />
      </div>
    </>
  );
}

function OpportunityFields({
  value,
  objectionTypes,
  onChange,
}: {
  value: CreateOpportunityInput;
  objectionTypes: { id: string; name: string }[];
  onChange: (value: CreateOpportunityInput) => void;
}) {
  return (
    <>
      <div className="mb-3">
        <label className="form-label">Objection Type</label>
        <select
          className="form-select"
          value={value.objectionTypeId}
          onChange={(event) =>
            onChange({
              ...value,
              objectionTypeId: event.target.value,
            })
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
          <label className="form-label">Strength</label>
          <select
            className="form-select"
            value={value.strength}
            onChange={(event) =>
              onChange({
                ...value,
                strength: event.target.value as CreateOpportunityInput["strength"],
              })
            }
          >
            <option value="weak">Weak</option>
            <option value="moderate">Moderate</option>
            <option value="strong">Strong</option>
          </select>
        </div>

        <div className="col-6">
          <label className="form-label">Timing</label>
          <select
            className="form-select"
            value={value.timingWindow}
            onChange={(event) =>
              onChange({
                ...value,
                timingWindow: event.target
                  .value as CreateOpportunityInput["timingWindow"],
              })
            }
          >
            <option value="after_question">After Question</option>
            <option value="after_answer">After Answer</option>
            <option value="before_answer">Before Answer</option>
          </select>
        </div>
      </div>

      <div className="mt-3">
        <label className="form-label">Expected Phrase</label>
        <input
          className="form-control"
          value={value.expectedPhrase}
          onChange={(event) =>
            onChange({
              ...value,
              expectedPhrase: event.target.value,
            })
          }
          placeholder="Objection, hearsay."
        />
      </div>

      <div className="mt-3">
        <label className="form-label">Explanation</label>
        <textarea
          className="form-control"
          rows={4}
          value={value.explanation}
          onChange={(event) =>
            onChange({
              ...value,
              explanation: event.target.value,
            })
          }
        />
      </div>

      <div className="form-check mt-3">
        <input
          className="form-check-input"
          type="checkbox"
          checked={value.isPrimary}
          onChange={(event) =>
            onChange({
              ...value,
              isPrimary: event.target.checked,
            })
          }
        />
        <label className="form-check-label">
          Primary objection opportunity
        </label>
      </div>
    </>
  );
}