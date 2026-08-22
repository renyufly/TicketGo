import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { LoadingPanel } from "../components/Feedback";

export function ProtectedRoute({ adminOnly = false }: { adminOnly?: boolean }) {
  const { token, user, restoring } = useAuth();
  const location = useLocation();

  if (restoring) return <LoadingPanel text="正在恢复登录状态…" />;
  if (!token || !user)
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  if (adminOnly && user.role !== "admin")
    return <Navigate to="/forbidden" replace />;
  return <Outlet />;
}
