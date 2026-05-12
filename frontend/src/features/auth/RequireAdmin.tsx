import { Navigate, Outlet, useLocation } from "react-router-dom";
import { getAuthToken, getStoredUser } from "./auth.store";

export function RequireAdmin() {
  const location = useLocation();

  const token = getAuthToken();
  const user = getStoredUser();

  if (!token || !user) {
    const redirectTo = encodeURIComponent(location.pathname);
    return <Navigate to={`/login?redirectTo=${redirectTo}`} replace />;
  }

  if (user.role !== "admin") {
    return (
      <div className="container py-4">
        <div className="alert alert-danger">
          You do not have permission to access this page.
        </div>
      </div>
    );
  }

  return <Outlet />;
}