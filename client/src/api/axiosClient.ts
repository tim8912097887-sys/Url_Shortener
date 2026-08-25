import axios from "axios";
import { env } from "../config/env";
import { useAuthStore } from "../store/useAuthStore";
import { normalizeApiError } from "./error";

// `rawClient` never carries the auth interceptors below. It exists solely so
// the 401-retry logic can call the refresh endpoint without re-triggering
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

// Attach the in-memory access token to every request. The token intentionally
// never touches localStorage/sessionStorage to keep it out of reach of XSS.
apiClient.interceptors.request.use((config) => {
  const { accessToken } = useAuthStore.getState();
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

// Concurrent 401s should trigger exactly one refresh call; every other
// request waits on the same in-flight promise instead of racing it.
let refreshPromise: Promise<string> | null = null;

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    const status = error.response?.status;
    const isRefreshCall = originalRequest?.url?.includes("/users/refresh");

    if (
      status !== 401 ||
      !originalRequest ||
      originalRequest._retry ||
      isRefreshCall
    ) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    try {
      if (!refreshPromise) {
        refreshPromise = rawClient
          .post("/users/refresh")
          .then((res) => {
            const newAccessToken =
              res.data?.data?.accessToken ?? res.data?.accessToken ?? null;
            useAuthStore.getState().setAccessToken(newAccessToken);
            return newAccessToken;
          })
          .finally(() => {
            refreshPromise = null;
          });
      }

      const newAccessToken = await refreshPromise;
      if (!newAccessToken) throw error;

      originalRequest.headers = originalRequest.headers ?? {};
      originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
      return apiClient(originalRequest);
    } catch (refreshError) {
      useAuthStore.getState().clearAuth();
      return Promise.reject(refreshError);
    }
  },
);

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    return Promise.reject(normalizeApiError(error));
  },
);
