import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { listScenarios } from "./scenarios.api";

export function ScenarioListPage() {
  const query = useQuery({
    queryKey: ["scenarios"],
    queryFn: listScenarios,
  });

  if (query.isLoading) {
    return <div className="container py-4">Loading scenarios...</div>;
  }

  if (query.isError) {
    return (
      <div className="container py-4">
        <div className="alert alert-danger">Failed to load scenarios.</div>
      </div>
    );
  }

  const scenarios = query.data?.data ?? [];

  return (
    <div className="container py-4">
      <h1 className="mb-3">Objection Training Scenarios</h1>

      <div className="row g-3">
        {scenarios.map((scenario) => (
          <div className="col-12 col-md-6" key={scenario.id}>
            <div className="card h-100">
              <div className="card-body">
                <div className="d-flex justify-content-between gap-2">
                  <h5 className="card-title">{scenario.title}</h5>
                  <span className="badge text-bg-primary">
                    {scenario.difficulty}
                  </span>
                </div>

                <p className="card-text">{scenario.description}</p>

                <div className="small text-muted mb-3">
                  {scenario.jurisdiction} · {scenario.practiceArea} ·{" "}
                  {scenario.hearingType}
                </div>

                <Link
                  className="btn btn-primary"
                  to={`/scenarios/${scenario.id}`}
                >
                  View Scenario
                </Link>
              </div>
            </div>
          </div>
        ))}

        {scenarios.length === 0 && (
          <div className="col-12">
            <div className="alert alert-info">
              No published scenarios found.
            </div>
          </div>
        )}
      </div>
    </div>
  );
}