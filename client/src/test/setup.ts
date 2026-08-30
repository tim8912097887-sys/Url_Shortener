import { beforeEach, vi } from "vitest";
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";

beforeEach(() => {
  cleanup();
  vi.clearAllMocks();
  vi.resetAllMocks();
});
