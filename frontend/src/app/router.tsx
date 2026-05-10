import { createBrowserRouter, Navigate } from "react-router-dom";
import { ScenarioListPage } from "../features/scenarios/ScenarioListPage";
import { ScenarioDetailPage } from "../features/scenarios/ScenarioDetailPage";
import { SessionPlayPage } from "../features/sessions/SessionPlayPage";
import { SessionDebriefPage } from "../features/debrief/SessionDebriefPage";

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
]);