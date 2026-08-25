import { useEffect } from "react";
import { RouterProvider } from "react-router";
import { router } from "./router";
import { useAuthStore } from "./store/useAuthStore";
import ErrorBoundary from "./components/layout/ErrorBoundary";

export default function App() {
  const initializeAuth = useAuthStore((s) => s.initializeAuth);

  // Access tokens live only in memory (see useAuthStore), so every hard
  // reload needs one silent refresh call, riding on the httpOnly cookie, to
  // find out whether the user is actually still logged in.
  useEffect(() => {
    initializeAuth();
  }, [initializeAuth]);

  return (
    <ErrorBoundary>
      <RouterProvider router={router} />
    </ErrorBoundary>
  );
}
