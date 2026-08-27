import axios from "axios";
import type { ApiFailure } from "../types/api.types";

export class ApiError extends Error {
  readonly code?: string;
  readonly status?: number;
  readonly cause?: unknown;

  constructor(
    message: string,
    options?: {
      code?: string;
      status?: number;
      cause?: unknown;
    },
  ) {
    super(message);

    this.name = "ApiError";
    this.code = options?.code;
    this.status = options?.status;
    this.cause = options?.cause;
  }
}

export function normalizeApiError(error: unknown): ApiError {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status;

    const data = error.response?.data as ApiFailure;

    return new ApiError(
      data?.error?.message ?? "An unexpected error occurred",
      {
        code: data?.error?.code,
        status,
        cause: error,
      },
    );
  }

  return new ApiError("An unexpected error occurred", {
    cause: error,
    status: 500,
    code: "INTERNAL_SERVER_ERROR",
  });
}
