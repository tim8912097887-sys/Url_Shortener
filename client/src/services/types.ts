export type SuccessResponse<T> = {
  state: "success";
  data: T;
  error: null;
  meta: {
    timestamp: string;
  };
  timestamp: string;
};

export type ErrorResponse = {
  state: "error";
  data: null;
  error: {
    code: string;
    message: string;
  };
  code: string;
  message: string;
  meta: {
    timestamp: string;
  };
  timestamp: string;
};

export type TokenResponseData = {
  accessToken: string;
  message: string;
};

export type TokenResponse = SuccessResponse<TokenResponseData>;
