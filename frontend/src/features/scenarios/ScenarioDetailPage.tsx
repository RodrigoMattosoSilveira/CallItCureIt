// frontend/src/features/scenarios/ScenarioDetailPage.tsx

import { Link, useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery } from "@tanstack/react-query";
import { getScenario, getScenarioTranscript } from "./scenarios.api";
import { createSession } from "../sessions/sessions.api";

export function ScenarioDetailPage() {
  const { scenarioId } = useParams<{ scenarioId: string }>();
  const navigate = useNavigate();

  const createSessionMutation = useMutation({
    mutationFn: () =>
      createSession({
        scenarioId: scenarioId!,
        mode: "spot_objection",
      }),
    onSuccess: (response) => {
      navigate(`/sessions/${response.data.id}/play`);
    },
  });

  const scenarioQuery = useQuery({
    queryKey: ["scenario", scenarioId],
    queryFn: () => getScenario(scenarioId!),
    enabled: Boolean(scenarioId),
  });

  const transcriptQuery = useQuery({
    queryKey: ["scenario-transcript", scenarioId],
    queryFn: () => getScenarioTranscript(scenarioId!),
    enabled: Boolean(scenarioId),
  });

  if (scenarioQuery.isLoading || transcriptQuery.isLoading) {
    return <div className="container py-4">Loading scenario...</div>;
  }

  if (scenarioQuery.isError || transcriptQuery.isError) {
    return (
      <div className="container py-4">
        <div className="alert alert-danger">Failed to load scenario.</div>
      </div>
    );
  }

  const scenario = scenarioQuery.data?.data;
  const lines = transcriptQuery.data?.data ?? [];

  if (!scenario || !scenarioId) {
    return (
      <div className="container py-4">
        <div className="alert alert-warning">Scenario not found.</div>
      </div>
    );
  }

  return (
    <div className="container py-4">
      <Link to="/scenarios" className="btn btn-link px-0">
        ← Back to scenarios
      </Link>

      <div className="d-flex justify-content-between align-items-start gap-3 mb-3">
        <div>
          <h1>{scenario.title}</h1>
          <p className="text-muted">{scenario.description}</p>
        </div>

        <span className="badge text-bg-primary">{scenario.difficulty}</span>
      </div>

      <div className="card mb-4">
        <div className="card-header">Actors</div>
        <div className="card-body">
          <ul className="mb-0">
            {scenario.actors.map((actor) => (
              <li key={actor.id}>
                <strong>{actor.name}</strong> — {actor.actorType}
              </li>
            ))}
          </ul>
        </div>
      </div>

      <div className="card mb-4">
        <div className="card-header">Transcript Preview</div>
        <div className="card-body">
          {lines.map((line) => (
            <div className="mb-3" key={line.id}>
              <div className="fw-bold">
                {line.speakerName ?? line.speakerType}
              </div>
              <div>{line.lineText}</div>
            </div>
          ))}
        </div>
      </div>

      <button
        className="btn btn-success"
        onClick={() => createSessionMutation.mutate()}
        disabled={createSessionMutation.isPending}
      >
        {createSessionMutation.isPending
          ? "Starting..."
          : "Start Training Session"}
      </button>

      {createSessionMutation.isError && (
        <div className="alert alert-danger mt-3">
          Failed to start training session.
        </div>
      )}
    </div>
  );
}