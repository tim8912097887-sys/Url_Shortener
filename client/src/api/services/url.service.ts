import type { UrlSchemaType } from "../../schema/url";
import { apiClient } from "../client/api-client";
import type { ShortenUrlSuccessResponse } from "../types/url.types";

const BASE = "/urls";

export const urlService = {
  async shortenUrl(urlInput: UrlSchemaType) {
    const response = await apiClient.post<ShortenUrlSuccessResponse>(
      `${BASE}/`,
      urlInput,
    );
    return response.data;
  },
};
