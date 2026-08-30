import { create } from "zustand";
import { devtools } from "zustand/middleware";
import { urlService } from "../api/services/url.service";
import type { ApiError } from "../api/errors/api-error";

type Url = {
  long_url: string;
  short_url: string;
  expired_at: string;
  clicks: number;
};

type UrlsStore = {
  urls: Url[] | null;
  // Cursor used to fetch the current page.
  currentCursor: string;

  // Cursor to fetch the next page.
  nextCursor: string | null;

  // Cursor history for Previous.
  cursorHistory: string[];
  isLoading: boolean;
  error: string | null;
  hasMore: boolean;
  page: number;
  previousPage: () => Promise<void>;
  nextPage: () => Promise<void>;
  getUrlsForUser(expiredAt: string): Promise<void>;
  reset: () => void;
};

const initialState = {
  urls: null,
  hasMore: false,
  currentCursor: "",
  nextCursor: null,
  cursorHistory: [],
  isLoading: false,
  error: null,
  page: 1,
};

export const useUrlsStore = create<UrlsStore>()(
  devtools(
    (set, get) => ({
      ...initialState,

      getUrlsForUser: async (cursor = "") => {
        set({
          isLoading: true,
          error: null,
        });

        try {
          const response = await urlService.getUrlsForUser(cursor);

          const { urls, hasMore } = response.data;

          const nextCursor =
            hasMore && urls && urls.length > 0
              ? urls[urls.length - 1].expired_at
              : null;

          set({
            urls,
            currentCursor: cursor,
            nextCursor,
            hasMore,
          });
        } catch (error) {
          const axiosError = error as ApiError;

          set({
            error:
              axiosError.message ??
              "Failed to load your URLs. Please try again.",
          });
        } finally {
          set({
            isLoading: false,
          });
        }
      },

      nextPage: async () => {
        const {
          nextCursor,
          currentCursor,
          cursorHistory,
          page,
          hasMore,
          isLoading,
        } = get();

        if (!hasMore || !nextCursor || isLoading) {
          return;
        }

        set({
          cursorHistory: [...cursorHistory, currentCursor],
          page: page + 1,
        });

        await get().getUrlsForUser(nextCursor);
      },

      previousPage: async () => {
        const { cursorHistory, page, isLoading } = get();

        if (page <= 1 || isLoading) {
          return;
        }

        const previousCursor = cursorHistory[cursorHistory.length - 1];

        const newHistory = cursorHistory.slice(0, -1);

        set({
          cursorHistory: newHistory,
          page: page - 1,
        });
        await get().getUrlsForUser(previousCursor);
      },

      reset: () => {
        set(initialState);
      },
    }),
    { name: "urls" },
  ),
);
