import type { ApiSuccess } from "./api.types";

export type ShortenData = {
  shortUrl: string;
  message: string;
};

export type ShortenUrlSuccessResponse = ApiSuccess<ShortenData>;
