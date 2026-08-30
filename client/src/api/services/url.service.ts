import type { UrlSchemaType } from "../../schema/url";
import { apiClient } from "../client/api-client";
import type {
  GetUrlsSuccessResponse,
  ShortenUrlSuccessResponse,
} from "../types/url.types";

const BASE = "/urls";

const MAX_URL_LIMIT = 5;

export const urlService = {
  async shortenUrl(urlInput: UrlSchemaType) {
    const response = await apiClient.post<ShortenUrlSuccessResponse>(
      `${BASE}/`,
      urlInput,
    );
    return response.data;
  },

  async getUrlsForUser(expiredAt: string) {
    const response = await apiClient.get<GetUrlsSuccessResponse>(
      `${BASE}/?expiredAt=${expiredAt}&limit=${MAX_URL_LIMIT}`,
    );
    return response.data;
  },
};
