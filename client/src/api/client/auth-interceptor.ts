// Attach the in-memory access token to every request. The token intentionally

import { useAuthStore } from "../../store/useAuthStore";
import { normalizeApiError } from "../errors/api-error";
import { authService } from "../services/auth.services";
import { apiClient } from "./api-client";

export function setupAuthInterceptors() {
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
        return Promise.reject(normalizeApiError(error));
      }

      originalRequest._retry = true;

      try {
        if (!refreshPromise) {
          refreshPromise = authService
            .refresh()
            .then((res) => {
              const newAccessToken = res.data?.accessToken ?? null;
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
        return Promise.reject(normalizeApiError(refreshError));
      }
    },
  );
}
