import { apiClient } from "../api/axiosClient";
import type { UrlSchemaType } from "../schema/url";
import type { SuccessResponse } from "./types";

const BASE = "/urls";

export async function shortenUrlRequest<T>(urlInput: UrlSchemaType) {
  const response = await apiClient.post<SuccessResponse<T>>(
    `${BASE}/`,
    urlInput,
  );
  return response.data;
}
