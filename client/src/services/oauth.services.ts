import { apiClient } from "../api/axiosClient";
import type { TokenResponse } from "./types";

const BASE = "/auth";

export async function oauthRequest() {
  const response = await apiClient.get<TokenResponse>(`${BASE}/google/login`);
  return response.data;
}
