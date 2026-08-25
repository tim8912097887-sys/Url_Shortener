import axios from "axios";
import type { ErrorResponse } from "./response";

export class ApiError extends Error {
  code?: string;
  status?: number;

  constructor(message: string, code?: string, status?: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

export function normalizeApiError(error: unknown): ApiError {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status;

    const data = error.response?.data as ErrorResponse;

    return new ApiError(
      data?.error?.message ?? "An unexpected error occurred",
      data?.error?.code,
      status,
    );
  }

  return new ApiError("An unexpected error occurred", "UNKNOWN", 500);
}
