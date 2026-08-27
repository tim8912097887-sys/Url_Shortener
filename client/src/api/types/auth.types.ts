import type { ApiSuccess } from "./api.types";

export type LoginResponseData = {
  accessToken: string;
  message: string;
};

export type RefreshResponseData = {
  accessToken: string;
  message: string;
};

export type SignupResponseData = {
  message: string;
};

export type LogoutResponseData = {
  message: string;
};

export type LogoutAllResponseData = {
  message: string;
};

export type LoginSuccessResponse = ApiSuccess<LoginResponseData>;
export type RefreshSuccessResponse = ApiSuccess<RefreshResponseData>;
export type SignupSuccessResponse = ApiSuccess<SignupResponseData>;
export type LogoutSuccessResponse = ApiSuccess<LogoutResponseData>;
export type LogoutAllSuccessResponse = ApiSuccess<LogoutAllResponseData>;
export type TokenResponse = ApiSuccess<{
  accessToken: string;
  message: string;
}>;
