export type ShortenData = {
  shortUrl: string;
  message: string;
};
export type SuccessResponse<T> = {
  state: "success";
  data: T;
  error: null;
  meta: {
    timestamp: string;
  };
};

export type ErrorResponse = {
  state: "error";
  data: null;
  error: {
    code: string;
    message: string;
  };
  meta: {
    timestamp: string;
  };
};
