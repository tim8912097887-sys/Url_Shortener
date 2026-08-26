import { create } from "zustand";
import { devtools } from "zustand/middleware";
import { decodeAccessToken, extractAccessToken } from "../utils/jwt";
import type { SignupSchemaType } from "../schema/signup";
import type { LoginSchemaType } from "../schema/login";
import {
  loginRequest,
  logoutAllRequest,
  logoutRequest,
  refreshRequest,
  signupRequest,
} from "../services/auth.services";
import { ApiError } from "../api/error";
import { type TokenResponseData } from "../services/types";

type AuthStore = {
  accessToken: string | null;
  user: any;
  isAuthenticated: boolean;
  isInitializing: boolean;
  error: string | null;
  clearError: () => void;
  setAccessToken: (accessToken: string) => void;
  clearAuth: () => void;
  initializeAuth: () => Promise<void>;
  login: (payload: LoginSchemaType) => Promise<any>;
  logout: () => Promise<void>;
  logoutAll: () => Promise<void>;
  signup: (payload: SignupSchemaType) => Promise<any>;
};

const initialState = {
  accessToken: null,
  user: null,
  isAuthenticated: false,
  // True until the initial silent-refresh (see initializeAuth) resolves, so
  // route guards can show a loading state instead of bouncing an
  // already-logged-in user to /login on a hard page reload.
  isInitializing: true,
  error: null,
};

export const useAuthStore = create<AuthStore>()(
  devtools(
    (set, get) => ({
      ...initialState,

      clearError: () => set({ error: null }),

      // Used internally by the axios response interceptor after a silent
      // token refresh — never call this with a token you haven't gotten
      // back from the server.
      setAccessToken: (accessToken) => {
        set(
          {
            accessToken,
            user: decodeAccessToken(accessToken),
            isAuthenticated: Boolean(accessToken),
          },
          false,
          "auth/setAccessToken",
        );
      },

      clearAuth: () =>
        set(
          { ...initialState, isInitializing: false },
          false,
          "auth/clearAuth",
        ),

      // Call once on app boot. Relies on the httpOnly refresh_token cookie;
      // a 401 here just means "not logged in," not an error to surface.
      initializeAuth: async () => {
        try {
          const data = await refreshRequest<TokenResponseData>();
          const accessToken = extractAccessToken(data);
          if (!accessToken)
            throw new Error("No access token in refresh response");
          set({
            accessToken,
            user: decodeAccessToken(accessToken),
            isAuthenticated: true,
            isInitializing: false,
          });
        } catch {
          set({ ...initialState, isInitializing: false });
        }
      },

      signup: async (payload) => {
        try {
          const data = await signupRequest(payload);
          return data;
        } catch (err: ApiError | any) {
          throw new Error(err.message || "An unexpected error occurred");
        }
      },

      login: async (payload) => {
        try {
          const data = await loginRequest<TokenResponseData>(payload);
          const accessToken = extractAccessToken(data);
          set({
            accessToken,
            user: decodeAccessToken(accessToken),
            isAuthenticated: Boolean(accessToken),
          });
          return data;
        } catch (err: ApiError | any) {
          throw new Error(err.message || "An unexpected error occurred");
        }
      },

      logout: async () => {
        try {
          const accessToken = get().accessToken;
          if (!accessToken) return;
          await logoutRequest(accessToken);
        } finally {
          set({ ...initialState, isInitializing: false });
        }
      },

      logoutAll: async () => {
        try {
          const accessToken = get().accessToken;
          if (!accessToken) return;
          await logoutAllRequest(accessToken);
        } finally {
          set({ ...initialState, isInitializing: false });
        }
      },
    }),
    { name: "auth-store" },
  ),
);
