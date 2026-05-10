import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { getSessionDebrief } from "../sessions/sessions.api";
import type { SessionEventType } from "../sessions/types";

export function SessionDebriefPage() {
  const { sessionId } = useParams<{ sessionId: string }>();

  const debriefQuery = useQuery({
    queryKey: ["session-debrief", sessionId],
    queryFn: () => getSessionDebrief(sessionId!),
    enabled: Boolean(sessionId),
  });

  if (debriefQuery.isLoading) {
    return <div className="container py-4">Loading debrief...</div>;
  }

  if (debriefQuery.isError || !debriefQuery.data) {
    return (
      <div className="container py-4">
        <div className="alert alert-danger">Failed to load debrief.</div>
      </div>
    );
  }

  const debrief = debriefQuery.data.data;
  const { session, score, summary, events, actions } = debrief;

  return (
    <div className="container py-4">
      <div className="d-flex justify-content-between align-items-start mb-3">
        <div>
          <h1>Session Debrief</h1>
          <div className="text-muted">
            Scenario: {session.scenarioId} · Mode: {session.mode}
          </div>
        </div>

        <span className="badge text-bg-primary">{session.status}</span>
      </div>

      <div className="d-flex gap-2 mb-4">
        <Link to={`/sessions/${session.id}/play`} className="btn btn-light">
          Back to Session
        </Link>

        <Link to={`/scenarios/${session.scenarioId}`} className="btn btn-light">
          Back to Scenario
        </Link>
      </div>

      <div className="row g-3 mb-4">
        <ScoreMetric label="Overall" value={score.overallScore} />
        <ScoreMetric label="Spotting" value={score.spottingAccuracy} />
        <ScoreMetric label="Legal Accuracy" value={score.legalAccuracy} />
        <ScoreMetric label="Timeliness" value={score.timeliness} />
        <ScoreMetric label="Phrasing" value={score.phrasing} />
        <ScoreMetric label="Strategy" value={score.strategy} />
      </div>

      <div className="card mb-4">
        <div className="card-header">Summary</div>
        <div className="card-body">
          <div className="row g-3">
            <SummaryItem
              label="Correct Actions"
              value={summary.correctActionCount}
            />
            <SummaryItem
              label="Missed / Incorrect"
              value={summary.missedOrIncorrectActionCount}
            />
            <SummaryItem
              label="Strongest Skill"
              value={summary.strongestSkill}
            />
            <SummaryItem
              label="Weakest Skill"
              value={summary.weakestSkill}
            />
          </div>
        </div>
      </div>

      <div className="card mb-4">
        <div className="card-header">Action Review</div>
        <div className="card-body">
          {actions.length === 0 && (
            <div className="text-muted">No trainee actions were evaluated.</div>
          )}

          {actions.map((item) => (
            <div className="border rounded p-3 mb-3" key={item.action.id}>
              <div className="d-flex justify-content-between gap-3">
                <div>
                  <div className="fw-bold">Your action</div>
                  <div>{item.action.rawText}</div>
                </div>

                <span
                  className={
                    item.evaluation.valid
                      ? "badge text-bg-success align-self-start"
                      : "badge text-bg-danger align-self-start"
                  }
                >
                  {item.evaluation.ruling}
                </span>
              </div>

              <hr />

              <div className="small text-muted mb-1">Coach feedback</div>
              <div>{item.evaluation.feedback}</div>

              <div className="row g-2 mt-3">
                <MiniScore
                  label="Legal"
                  value={item.evaluation.legalAccuracyScore}
                />
                <MiniScore
                  label="Phrasing"
                  value={item.evaluation.phrasingScore}
                />
                <MiniScore
                  label="Strategy"
                  value={item.evaluation.strategyScore}
                />
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="card">
        <div className="card-header">Full Transcript</div>
        <div className="card-body">
          {events.length === 0 && (
            <div className="text-muted">No transcript events yet.</div>
          )}

          {events.map((event) => (
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
    </div>
  );
}

function ScoreMetric({
  label,
  value,
}: {
  label: string;
  value: number;
}) {
  return (
    <div className="col-6 col-md-2">
      <div className="border rounded p-3 text-center h-100">
        <div className="small text-muted">{label}</div>
        <div className="fs-4 fw-bold">{Math.round(value)}</div>
      </div>
    </div>
  );
}

function SummaryItem({
  label,
  value,
}: {
  label: string;
  value: string | number;
}) {
  return (
    <div className="col-6 col-md-3">
      <div className="border rounded p-3 h-100">
        <div className="small text-muted">{label}</div>
        <div className="fw-semibold">{value}</div>
      </div>
    </div>
  );
}

function MiniScore({
  label,
  value,
}: {
  label: string;
  value: number;
}) {
  return (
    <div className="col-4">
      <div className="border rounded p-2 text-center">
        <div className="small text-muted">{label}</div>
        <div className="fw-bold">{Math.round(value)}</div>
      </div>
    </div>
  );
}

function getEventClassName(eventType: SessionEventType): string {
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