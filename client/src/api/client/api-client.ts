// `rawClient` never carries the auth interceptors below. It exists solely so
// the 401-retry logic can call the refresh endpoint without re-triggering

import axios from "axios";
import { env } from "../../config/env";

// itself (a refresh call that itself 401s must not try to refresh again).
export const rawClient = axios.create({
  baseURL: env.apiBaseUrl,
  timeout: 10000,
  withCredentials: true, // send the httpOnly refresh_token cookie
});

export const apiClient = axios.create({
  baseURL: env.apiBaseUrl,
  timeout: 10000,
  withCredentials: true,
});
