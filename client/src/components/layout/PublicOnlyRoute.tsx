import { Navigate, Outlet } from "react-router";
import { useAuthStore } from "../../store/useAuthStore";

// Keeps a logged-in user off /login and /signup.
export default function PublicOnlyRoute() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const isInitializing = useAuthStore((s) => s.isInitializing);

  if (!isInitializing && isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
}
