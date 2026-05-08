import { FormEvent, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  advanceSession,
  getSession,
  submitTraineeAction,
} from "./sessions.api";

export function SessionPlayPage() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const queryClient = useQueryClient();

  const [objectionText, setObjectionText] = useState("");
  const [lastEvaluation, setLastEvaluation] = useState<string | null>(null);

  const sessionQuery = useQuery({
    queryKey: ["session", sessionId],
    queryFn: () => getSession(sessionId!),
    enabled: Boolean(sessionId),
  });

  const advanceMutation = useMutation({
    mutationFn: () => advanceSession(sessionId!),
    onSuccess: () => {
      setObjectionText("");
      setLastEvaluation(null);
      queryClient.invalidateQueries({
        queryKey: ["session", sessionId],
      });
    },
  });

  const submitActionMutation = useMutation({
    mutationFn: () =>
      submitTraineeAction(sessionId!, {
        actionType: "object",
        rawText: objectionText.trim(),
      }),
    onSuccess: (response) => {
      setObjectionText("");
      setLastEvaluation(response.data.evaluation.feedback);
      queryClient.invalidateQueries({
        queryKey: ["session", sessionId],
      });
    },
  });

  const passMutation = useMutation({
    mutationFn: () =>
      submitTraineeAction(sessionId!, {
        actionType: "pass",
        rawText: "Pass",
      }),
    onSuccess: (response) => {
      setObjectionText("");
      setLastEvaluation(response.data.evaluation.feedback);
      queryClient.invalidateQueries({
        queryKey: ["session", sessionId],
      });
    },
  });

  const session = sessionQuery.data?.data;

  const sortedEvents = useMemo(() => {
    return [...(session?.events ?? [])].sort((a, b) => {
      if (a.sequenceNo !== b.sequenceNo) {
        return a.sequenceNo - b.sequenceNo;
      }

      return (a.createdAt ?? "").localeCompare(b.createdAt ?? "");
    });
  }, [session?.events]);

  function handleSubmitObjection(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!objectionText.trim()) {
      return;
    }

    submitActionMutation.mutate();
  }

  if (sessionQuery.isLoading) {
    return <div className="container py-4">Loading session...</div>;
  }

  if (sessionQuery.isError || !session || !sessionId) {
    return (
      <div className="container py-4">
        <div className="alert alert-danger">Failed to load session.</div>
      </div>
    );
  }

  const isCompleted = session.status === "completed";
  const hasStarted = session.currentSequenceNo > 0;
  const isBusy =
    advanceMutation.isPending ||
    submitActionMutation.isPending ||
    passMutation.isPending;

  return (
    <div className="container py-4">
      <div className="d-flex justify-content-between align-items-start mb-3">
        <div>
          <h1>Training Session</h1>
          <div className="text-muted">
            Scenario: {session.scenarioId} · Mode: {session.mode}
          </div>
        </div>

        <span
          className={
            isCompleted ? "badge text-bg-success" : "badge text-bg-primary"
          }
        >
          {session.status}
        </span>
      </div>

      <div className="card mb-4">
        <div className="card-header">Courtroom Transcript</div>

        <div className="card-body">
          {sortedEvents.length === 0 && (
            <div className="text-muted">
              No transcript lines yet. Click “Next Line” to begin.
            </div>
          )}

          {sortedEvents.map((event) => (
            <div className="mb-3" key={event.id}>
              <div className="fw-bold">
                {event.actor || "Courtroom"}
                <span className="text-muted fw-normal">
                  {" "}
                  · line {event.sequenceNo}
                </span>
              </div>

              <div className={getEventClassName(event.eventType)}>
                {event.text}
              </div>
            </div>
          ))}
        </div>
      </div>

      {!hasStarted && !isCompleted && (
        <div className="alert alert-info">
          Start the transcript before submitting an objection.
        </div>
      )}

      {hasStarted && !isCompleted && (
        <div className="card mb-4">
          <div className="card-header">Your Objection</div>
          <div className="card-body">
            <form onSubmit={handleSubmitObjection}>
              <label htmlFor="objectionText" className="form-label">
                Type your objection
              </label>

              <textarea
                id="objectionText"
                className="form-control"
                rows={3}
                value={objectionText}
                onChange={(event) => setObjectionText(event.target.value)}
                placeholder="Example: Objection, hearsay."
                disabled={isBusy}
              />

              <div className="d-flex gap-2 mt-3">
                <button
                  type="submit"
                  className="btn btn-warning"
                  disabled={isBusy || !objectionText.trim()}
                >
                  {submitActionMutation.isPending
                    ? "Submitting..."
                    : "Submit Objection"}
                </button>

                <button
                  type="button"
                  className="btn btn-outline-secondary"
                  disabled={isBusy}
                  onClick={() => passMutation.mutate()}
                >
                  Pass
                </button>
              </div>
            </form>

            {lastEvaluation && (
              <div className="alert alert-info mt-3">
                <strong>Coach Feedback:</strong> {lastEvaluation}
              </div>
            )}
          </div>
        </div>
      )}

      {(advanceMutation.isError ||
        submitActionMutation.isError ||
        passMutation.isError) && (
        <div className="alert alert-danger">
          The session action failed. Check the backend logs.
        </div>
      )}

      <div className="d-flex gap-2">
        <button
          className="btn btn-primary"
          onClick={() => advanceMutation.mutate()}
          disabled={isBusy || isCompleted}
        >
          {advanceMutation.isPending ? "Loading..." : "Next Line"}
        </button>

        <Link to={`/scenarios/${session.scenarioId}`} className="btn btn-light">
          Back to Scenario
        </Link>
      </div>

      {isCompleted && (
        <div className="alert alert-success mt-4">
          Session complete. Debrief will be added in a later phase.
        </div>
      )}
    </div>
  );
}

function getEventClassName(eventType: string): string {
  switch (eventType) {
    case "trainee_objection":
    case "trainee_response":
      return "border-start border-4 ps-3";
    case "judge_ruling":
      return "border-start border-4 ps-3 fw-semibold";
    case "coach_feedback":
      return "border-start border-4 ps-3 text-muted";
    default:
      return "";
  }
}