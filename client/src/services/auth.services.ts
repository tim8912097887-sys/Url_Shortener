import { apiClient } from "../api/axiosClient";
import type { LoginSchemaType } from "../schema/login";
import type { SignupSchemaType } from "../schema/signup";
import type { SuccessResponse } from "./types";

const BASE = "/users";

export async function loginRequest<T>(loginInput: LoginSchemaType) {
  const response = await apiClient.post<SuccessResponse<T>>(
    `${BASE}/login`,
    loginInput,
  );
  return response.data;
}

export async function signupRequest<T>(signupInput: SignupSchemaType) {
  const response = await apiClient.post<SuccessResponse<T>>(
    `${BASE}/signup`,
    signupInput,
  );
  return response.data;
}

export async function logoutRequest<T>(accessToken: string) {
  const response = await apiClient.post<SuccessResponse<T>>(
    `${BASE}/logout`,
    {},
    {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    },
  );
  return response.data;
}

export async function refreshRequest<T>() {
  const response = await apiClient.post<SuccessResponse<T>>(
    `${BASE}/refresh`,
    {},
  );
  return response.data;
}

export async function logoutAllRequest<T>(accessToken: string) {
  const response = await apiClient.post<SuccessResponse<T>>(
    `${BASE}/logout-all`,
    {},
    {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    },
  );
  return response.data;
}
