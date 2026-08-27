export type ApiMeta = {
  timestamp: string;
};

export type ApiSuccess<T> = {
  state: "success";
  data: T;
  error: null;
  meta: ApiMeta;
};

export type ApiFailure = {
  state: "error";
  data: null;
  error: {
    code: string;
    message: string;
  };
  meta: ApiMeta;
};

export type ApiResponse<T> = ApiSuccess<T> | ApiFailure;
