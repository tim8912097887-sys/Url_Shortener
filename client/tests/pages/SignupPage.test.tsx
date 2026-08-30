import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import SignupPage from "../../src/pages/SignupPage";

const mockSignup = vi.fn();
const mockNavigate = vi.fn();

vi.mock("../../src/store/useAuthStore", () => ({
  useAuthStore: (selector: any) =>
    selector({
      signup: mockSignup,
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

const renderSignupPage = (initialEntries = [{ pathname: "/signup" }]) =>
  render(
    <MemoryRouter initialEntries={initialEntries as any}>
      <SignupPage />
    </MemoryRouter>,
  );

describe("Signup page", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("When loads the page should display the signup form", () => {
    renderSignupPage();

    expect(
      screen.getByRole("heading", { name: "Create an account" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Start in under a minute.")).toBeInTheDocument();
    expect(screen.getByLabelText("Username")).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(screen.getByLabelText("Confirm password")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Create account" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Log in" })).toHaveAttribute(
      "href",
      "/login",
    );
  });

  it("When submits an empty form should show validation errors", async () => {
    const user = userEvent.setup();
    renderSignupPage();

    await user.click(screen.getByRole("button", { name: "Create account" }));

    expect(
      await screen.findByText(/username must be at least 3 characters/i),
    ).toBeInTheDocument();
    expect(screen.getByText(/invalid email/i)).toBeInTheDocument();
    expect(
      screen.getAllByText(/at least 8 characters/i).length,
    ).toBeGreaterThanOrEqual(2);
  });

  it("When passwords do not match should show the mismatch validation", async () => {
    const user = userEvent.setup();
    renderSignupPage();

    await user.type(screen.getByLabelText("Username"), "janedoe");
    await user.type(screen.getByLabelText("Email"), "jane@example.com");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.type(screen.getByLabelText("Confirm password"), "different123");
    await user.click(screen.getByRole("button", { name: "Create account" }));

    expect(
      await screen.findByText("Passwords do not match"),
    ).toBeInTheDocument();
  });

  it("When submits valid signup details should call signup and navigate to login", async () => {
    const user = userEvent.setup();
    mockSignup.mockResolvedValueOnce({ ok: true });
    renderSignupPage();

    await user.type(screen.getByLabelText("Username"), "janedoe");
    await user.type(screen.getByLabelText("Email"), "jane@example.com");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.type(screen.getByLabelText("Confirm password"), "password123");
    await user.click(screen.getByRole("button", { name: "Create account" }));

    await waitFor(() => {
      expect(mockSignup).toHaveBeenCalledWith({
        username: "janedoe",
        email: "jane@example.com",
        password: "password123",
        confirmPassword: "password123",
      });
    });

    expect(mockNavigate).toHaveBeenCalledWith("/login", {
      replace: true,
      state: { justSignedUp: true },
    });
  });

  it("When signup fails should display the error message", async () => {
    const user = userEvent.setup();
    mockSignup.mockRejectedValueOnce(new Error("Email already exists"));
    renderSignupPage();

    await user.type(screen.getByLabelText("Username"), "janedoe");
    await user.type(screen.getByLabelText("Email"), "jane@example.com");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.type(screen.getByLabelText("Confirm password"), "password123");
    await user.click(screen.getByRole("button", { name: "Create account" }));

    expect(await screen.findByText("Email already exists")).toBeInTheDocument();
  });
});
