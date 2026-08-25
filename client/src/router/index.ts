import { createBrowserRouter } from "react-router";
import RootLayout from "../components/layout/RootLayout";
import PublicOnlyRoute from "../components/layout/PublicOnlyRoute";
import LoginPage from "../pages/LoginPage";
import NotFoundPage from "../pages/NotFoundPage";
import SignupPage from "../pages/SignupPage";

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

      { path: "*", Component: NotFoundPage },
    ],
  },
]);
