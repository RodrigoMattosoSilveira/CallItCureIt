import { createBrowserRouter, Navigate } from "react-router-dom";
import { ScenarioListPage } from "../features/scenarios/ScenarioListPage";
import { ScenarioDetailPage } from "../features/scenarios/ScenarioDetailPage";

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
]);