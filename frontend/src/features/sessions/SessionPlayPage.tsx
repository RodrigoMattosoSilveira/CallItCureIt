import { useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { advanceSession, getSession } from "./sessions.api";

export function SessionPlayPage() {
  const { sessionId } = useParams();
  const queryClient = useQueryClient();

  const sessionQuery = useQuery({
    queryKey: ["session", sessionId],
    queryFn: () => getSession(sessionId!),
    enabled: Boolean(sessionId),
  });

  const advanceMutation = useMutation({
    mutationFn: () => advanceSession(sessionId!),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["session", sessionId],
      });
    },
  });

  const session = sessionQuery.data?.data;

  const sortedEvents = useMemo(() => {
    return [...(session?.events ?? [])].sort(
      (a, b) => a.sequenceNo - b.sequenceNo
    );
  }, [session?.events]);

  if (sessionQuery.isLoading) {
    return <div className="container py-4">Loading session...</div>;
  }

  if (sessionQuery.isError || !session) {
    return (
      <div className="container py-4">
        <div className="alert alert-danger">Failed to load session.</div>
      </div>
    );
  }

  const isCompleted = session.status === "completed";

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
              <div>{event.text}</div>
            </div>
          ))}
        </div>
      </div>

      {advanceMutation.isError && (
        <div className="alert alert-danger">
          Could not advance the session.
        </div>
      )}

      <div className="d-flex gap-2">
        <button
          className="btn btn-primary"
          onClick={() => advanceMutation.mutate()}
          disabled={advanceMutation.isPending || isCompleted}
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