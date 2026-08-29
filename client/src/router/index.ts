import { createBrowserRouter } from "react-router";
import RootLayout from "../components/layout/RootLayout";
import PublicOnlyRoute from "../components/layout/PublicOnlyRoute";
import LoginPage from "../pages/LoginPage";
import NotFoundPage from "../pages/NotFoundPage";
import SignupPage from "../pages/SignupPage";
import HomePage from "../pages/HomePage";
import Dashboard from "../pages/Dashboard";
import { OAuthPage } from "../pages/OAuthPage";
import { ProtectedRoute } from "../components/layout/ProtectedRoute";

export const router = createBrowserRouter([
  {
    Component: RootLayout,
    children: [
      {
        Component: PublicOnlyRoute,
        children: [
          { path: "/login", Component: LoginPage },
          { path: "/signup", Component: SignupPage },
        ],
      },
      {
        Component: ProtectedRoute,
        children: [
          {
            path: "/dashboard",
            Component: Dashboard,
          },
        ],
      },
      { path: "/", Component: HomePage },
      { path: "/oauth-callback", Component: OAuthPage },
      { path: "*", Component: NotFoundPage },
    ],
  },
]);
