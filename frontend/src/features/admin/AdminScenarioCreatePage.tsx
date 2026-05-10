import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useMutation } from "@tanstack/react-query";
import { createAdminScenario } from "./admin.api";
import type { CreateScenarioInput } from "./admin.types";

const initialForm: CreateScenarioInput = {
  title: "",
  description: "",
  jurisdiction: "federal",
  practiceArea: "civil",
  hearingType: "trial_direct_examination",
  difficulty: "beginner",
  status: "draft",
};

export function AdminScenarioCreatePage() {
  const navigate = useNavigate();
  const [form, setForm] = useState<CreateScenarioInput>(initialForm);

  const mutation = useMutation({
    mutationFn: createAdminScenario,
    onSuccess: (response) => {
      navigate(`/admin/scenarios/${response.data.id}`);
    },
  });

  function updateField<K extends keyof CreateScenarioInput>(
    key: K,
    value: CreateScenarioInput[K]
  ) {
    setForm((current) => ({
      ...current,
      [key]: value,
    }));
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!form.title.trim()) {
      return;
    }

    mutation.mutate(form);
  }

  return (
    <div className="container py-4">
      <h1>New Scenario</h1>

      <form className="card mt-4" onSubmit={handleSubmit}>
        <div className="card-body">
          <div className="mb-3">
            <label className="form-label" htmlFor="title">
              Title
            </label>
            <input
              id="title"
              className="form-control"
              value={form.title}
              onChange={(event) => updateField("title", event.target.value)}
            />
          </div>

          <div className="mb-3">
            <label className="form-label" htmlFor="description">
              Description
            </label>
            <textarea
              id="description"
              className="form-control"
              rows={3}
              value={form.description}
              onChange={(event) =>
                updateField("description", event.target.value)
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
                value={form.jurisdiction}
                onChange={(event) =>
                  updateField("jurisdiction", event.target.value)
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
                value={form.practiceArea}
                onChange={(event) =>
                  updateField("practiceArea", event.target.value)
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
                value={form.hearingType}
                onChange={(event) =>
                  updateField("hearingType", event.target.value)
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
                value={form.difficulty}
                onChange={(event) =>
                  updateField(
                    "difficulty",
                    event.target.value as CreateScenarioInput["difficulty"]
                  )
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
                value={form.status}
                onChange={(event) =>
                  updateField(
                    "status",
                    event.target.value as CreateScenarioInput["status"]
                  )
                }
              >
                <option value="draft">Draft</option>
                <option value="published">Published</option>
                <option value="archived">Archived</option>
              </select>
            </div>
          </div>

          {mutation.isError && (
            <div className="alert alert-danger mt-3">
              Failed to create scenario.
            </div>
          )}

          <div className="d-flex gap-2 mt-4">
            <button className="btn btn-primary" disabled={mutation.isPending}>
              {mutation.isPending ? "Creating..." : "Create Scenario"}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}