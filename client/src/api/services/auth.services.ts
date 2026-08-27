import type { LoginSchemaType } from "../../schema/login";
import type { SignupSchemaType } from "../../schema/signup";
import { apiClient, rawClient } from "../client/api-client";
import type {
  LoginSuccessResponse,
  LogoutAllSuccessResponse,
  LogoutSuccessResponse,
  RefreshSuccessResponse,
  SignupSuccessResponse,
} from "../types/auth.types";

const BASE = "/users";

export const authService = {
  async login(loginInput: LoginSchemaType) {
    const response = await apiClient.post<LoginSuccessResponse>(
      `${BASE}/login`,
      loginInput,
    );
    return response.data;
  },
  async signup(signupInput: SignupSchemaType) {
    const response = await apiClient.post<SignupSuccessResponse>(
      `${BASE}/signup`,
      signupInput,
    );
    return response.data;
  },
  async logout() {
    const response = await apiClient.post<LogoutSuccessResponse>(
      `${BASE}/logout`,
    );
    return response.data;
  },
  async refresh() {
    const response = await rawClient.post<RefreshSuccessResponse>(
      `${BASE}/refresh`,
    );
    return response.data;
  },
  async logoutAll() {
    const response = await apiClient.post<LogoutAllSuccessResponse>(
      `${BASE}/logout-all`,
    );
    return response.data;
  },
};
