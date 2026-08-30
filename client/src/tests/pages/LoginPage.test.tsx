import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import LoginPage from "../../pages/LoginPage";

const mockLogin = vi.fn();
const mockNavigate = vi.fn();

vi.mock("../../store/useAuthStore", () => ({
  useAuthStore: (selector: any) =>
    selector({
      login: mockLogin,
    }),
}));

vi.mock("react-router", async () => {
  const actual =
    await vi.importActual<typeof import("react-router")>("react-router");

  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const renderLoginPage = (
  initialEntries = [{ pathname: "/login", state: {} }],
) =>
  render(
    <MemoryRouter initialEntries={initialEntries as any}>
      <LoginPage />
    </MemoryRouter>,
  );

describe("Login page", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("When loads the page should display the login form", () => {
    renderLoginPage();

    expect(
      screen.getByRole("heading", { name: "Welcome back" }),
    ).toBeInTheDocument();
    expect(
      screen.getByText("Log in to continue to your account."),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Log in" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Sign up" })).toHaveAttribute(
      "href",
      "/signup",
    );
  });

  it("When loads the page with justSignedUp should show the success alert", () => {
    renderLoginPage([{ pathname: "/login", state: { justSignedUp: true } }]);

    expect(
      screen.getByText("Account created. You can log in now."),
    ).toBeInTheDocument();
  });

  it("When submits an empty form should show validation errors", async () => {
    const user = userEvent.setup();
    renderLoginPage();

    await user.click(screen.getByRole("button", { name: "Log in" }));

    expect(await screen.findByText(/invalid email/i)).toBeInTheDocument();
    expect(screen.getByText(/at least 8 characters/i)).toBeInTheDocument();
  });

  it("When submits valid credentials should call login and navigate to dashboard", async () => {
    const user = userEvent.setup();
    mockLogin.mockResolvedValueOnce({ accessToken: "token" });
    renderLoginPage();

    await user.type(screen.getByLabelText("Email"), "user@example.com");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.click(screen.getByRole("button", { name: "Log in" }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        email: "user@example.com",
        password: "password123",
      });
    });

    expect(mockNavigate).toHaveBeenCalledWith("/dashboard", { replace: true });
  });

  it("When login page state includes redirect target should navigate there after login", async () => {
    const user = userEvent.setup();
    mockLogin.mockResolvedValueOnce({ accessToken: "token" });
    renderLoginPage([
      { pathname: "/login", state: { from: { pathname: "/dashboard" } } },
    ]);

    await user.type(screen.getByLabelText("Email"), "user@example.com");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.click(screen.getByRole("button", { name: "Log in" }));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith("/dashboard", {
        replace: true,
      });
    });
  });

  it("When login fails should display the error message", async () => {
    const user = userEvent.setup();
    mockLogin.mockRejectedValueOnce(new Error("Invalid credentials"));
    renderLoginPage();

    await user.type(screen.getByLabelText("Email"), "user@example.com");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.click(screen.getByRole("button", { name: "Log in" }));

    expect(await screen.findByText("Invalid credentials")).toBeInTheDocument();
  });
});
