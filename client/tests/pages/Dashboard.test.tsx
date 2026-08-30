import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { urlService } from "../../src/api/services/url.service";
import { useUrlsStore } from "../../src/store/useUrlsStore";
import Dashboard from "../../src/pages/Dashboard";
import { env } from "../../src/config/env";

vi.mock("../../src/api/services/url.service", () => ({
  urlService: {
    getUrlsForUser: vi.fn(),
  },
}));

const mockedGetUrlsForUser = vi.mocked(urlService.getUrlsForUser);

describe("Dashboard page", () => {
  beforeEach(() => {
    useUrlsStore.getState().reset();
  });

  it("When loads the page with urls should displays the user's URLs", async () => {
    mockedGetUrlsForUser.mockResolvedValueOnce({
      data: {
        urls: [
          {
            long_url: "https://google.com",
            short_url: "abc123",
            expired_at: "2026-09-01T00:00:00Z",
            clicks: 10,
          },
        ],
        hasMore: false,
      },
    } as any);

    render(<Dashboard />);

    expect(screen.getByLabelText("Loading URLs")).toBeInTheDocument();

    expect(await screen.findByText("https://google.com")).toBeInTheDocument();

    expect(
      screen.getByText(`${env.apiBaseUrl}/urls/abc123`),
    ).toBeInTheDocument();

    expect(screen.getByText("Page 1")).toBeInTheDocument();
  });

  it("When loads the page without urls should displays empty state", async () => {
    mockedGetUrlsForUser.mockResolvedValueOnce({
      data: {
        urls: [],
        hasMore: false,
      },
    } as any);

    render(<Dashboard />);

    expect(await screen.findByText("No URLs yet")).toBeInTheDocument();
  });

  it("When loads the page with error should displays error message", async () => {
    mockedGetUrlsForUser.mockRejectedValueOnce(
      new Error("Something went wrong"),
    );

    render(<Dashboard />);

    expect(await screen.findByText("Something went wrong")).toBeInTheDocument();
  });
});
