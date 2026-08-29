import type { ApiSuccess } from "./api.types";

export type UrlSchemaType = {
  short_url: string;
  long_url: string;
  expired_at: string;
  clicks: number;
};

export type GetUrlsData = {
  urls: UrlSchemaType[] | null;
  message: string;
};

export type GetUrlsSuccessResponse = ApiSuccess<GetUrlsData>;

export type ShortenData = {
  shortUrl: string;
  message: string;
};

export type ShortenUrlSuccessResponse = ApiSuccess<ShortenData>;
