import { createBrowserRouter, Navigate } from "react-router-dom";
import { ScenarioListPage } from "../features/scenarios/ScenarioListPage";
import { ScenarioDetailPage } from "../features/scenarios/ScenarioDetailPage";
import { SessionPlayPage } from "../features/sessions/SessionPlayPage";
import { SessionDebriefPage } from "../features/debrief/SessionDebriefPage";
import { AdminScenarioListPage } from "../features/admin/AdminScenarioListPage";
import { AdminScenarioCreatePage } from "../features/admin/AdminScenarioCreatePage";
import { AdminScenarioDetailPage } from "../features/admin/AdminScenarioDetailPage";
import { LoginPage } from "../features/auth/LoginPage";
import { RequireAdmin } from "../features/auth/RequireAdmin";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <Navigate to="/scenarios" replace />,
  },
  {
    path: "/scenarios",
    element: <ScenarioListPage />,
  },
  {
    path: "/scenarios/:scenarioId",
    element: <ScenarioDetailPage />,
  },
  {
    path: "/sessions/:sessionId/play",
    element: <SessionPlayPage />,
  },
  {
    path: "/sessions/:sessionId/debrief",
    element: <SessionDebriefPage />,
  },
  {
    path: "/admin/scenarios",
    element: <AdminScenarioListPage />,
  },
  {
    path: "/admin/scenarios/new",
    element: <AdminScenarioCreatePage />,
  },
  {
    path: "/admin/scenarios/:scenarioId",
    element: <AdminScenarioDetailPage />,
  },
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    element: <RequireAdmin />,
    children: [
      {
        path: "/admin/scenarios",
        element: <AdminScenarioListPage />,
      },
      {
        path: "/admin/scenarios/new",
        element: <AdminScenarioCreatePage />,
      },
      {
        path: "/admin/scenarios/:scenarioId",
        element: <AdminScenarioDetailPage />,
      },
    ],
  },
]);