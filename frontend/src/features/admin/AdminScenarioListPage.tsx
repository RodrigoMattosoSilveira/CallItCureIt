import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { listAdminScenarios } from "./admin.api";

export function AdminScenarioListPage() {
  const query = useQuery({
    queryKey: ["admin-scenarios"],
    queryFn: listAdminScenarios,
  });

  if (query.isLoading) {
    return <div className="container py-4">Loading admin scenarios...</div>;
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
      <div className="d-flex justify-content-between align-items-start mb-4">
        <div>
          <h1>Admin Scenarios</h1>
          <p className="text-muted mb-0">
            Create, review, and publish training scenarios.
          </p>
        </div>

        <Link className="btn btn-primary" to="/admin/scenarios/new">
          New Scenario
        </Link>
      </div>

      <div className="card">
        <div className="card-body">
          {scenarios.length === 0 && (
            <div className="text-muted">No scenarios found.</div>
          )}

          {scenarios.map((scenario) => (
            <div
              className="border rounded p-3 mb-3 d-flex justify-content-between gap-3"
              key={scenario.id}
            >
              <div>
                <h5 className="mb-1">{scenario.title}</h5>
                <div className="text-muted small">
                  {scenario.jurisdiction} · {scenario.practiceArea} ·{" "}
                  {scenario.hearingType}
                </div>
                <div className="mt-2">
                  <span className="badge text-bg-secondary me-2">
                    {scenario.difficulty}
                  </span>
                  <span className="badge text-bg-primary">
                    {scenario.status}
                  </span>
                </div>
              </div>

              <Link
                className="btn btn-outline-primary align-self-center"
                to={`/admin/scenarios/${scenario.id}`}
              >
                Edit
              </Link>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}