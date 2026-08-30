import { beforeEach, describe, expect, it, vi } from "vitest";
import { urlService } from "../../api/services/url.service";
import { useUrlsStore } from "../../store/useUrlsStore";

vi.mock("../../api/services/url.service", () => ({
  urlService: {
    getUrlsForUser: vi.fn(),
  },
}));

const mockedGetUrlsForUser = vi.mocked(urlService.getUrlsForUser);

const page1 = {
  urls: [
    {
      long_url: "https://example.com/1",
      short_url: "abc",
      expired_at: "2026-08-30T10:00:00Z",
      clicks: 10,
    },
    {
      long_url: "https://example.com/2",
      short_url: "def",
      expired_at: "2026-08-29T10:00:00Z",
      clicks: 20,
    },
  ],
  hasMore: true,
};

const page2 = {
  urls: [
    {
      long_url: "https://example.com/3",
      short_url: "ghi",
      expired_at: "2026-08-28T10:00:00Z",
      clicks: 5,
    },
  ],
  hasMore: false,
};

describe("useUrlsStore", () => {
  beforeEach(() => {
    useUrlsStore.getState().reset();
  });

  it("Wehn loads the first page, page info should be set correctly", async () => {
    mockedGetUrlsForUser.mockResolvedValueOnce({
      data: page1,
    } as any);

    await useUrlsStore.getState().getUrlsForUser("");

    const state = useUrlsStore.getState();

    expect(mockedGetUrlsForUser).toHaveBeenCalledWith("");

    expect(state.urls).toEqual(page1.urls);
    expect(state.hasMore).toBe(true);
    expect(state.currentCursor).toBe("");
    expect(state.nextCursor).toBe(page1.urls[page1.urls.length - 1].expired_at);
    expect(state.page).toBe(1);
    expect(state.isLoading).toBe(false);
    expect(state.error).toBeNull();
  });

  it("When clicks the next page, page info should be set correctly", async () => {
    mockedGetUrlsForUser
      .mockResolvedValueOnce({
        data: page1,
      } as any)
      .mockResolvedValueOnce({
        data: page2,
      } as any);

    await useUrlsStore.getState().getUrlsForUser("");

    await useUrlsStore.getState().nextPage();

    const state = useUrlsStore.getState();

    expect(mockedGetUrlsForUser).toHaveBeenNthCalledWith(1, "");
    expect(mockedGetUrlsForUser).toHaveBeenNthCalledWith(
      2,
      "2026-08-29T10:00:00Z",
    );

    expect(state.page).toBe(2);
    expect(state.currentCursor).toBe(
      page1.urls[page1.urls.length - 1].expired_at,
    );
    expect(state.urls).toEqual(page2.urls);
    expect(state.hasMore).toBe(false);
    expect(state.nextCursor).toBeNull();
    expect(state.cursorHistory.length).toEqual(1);
  });

  it("When clicks the previous page, page info should be set correctly", async () => {
    mockedGetUrlsForUser
      .mockResolvedValueOnce({
        data: page1,
      } as any)
      .mockResolvedValueOnce({
        data: page2,
      } as any)
      .mockResolvedValueOnce({
        data: page1,
      } as any);

    await useUrlsStore.getState().getUrlsForUser("");

    await useUrlsStore.getState().nextPage();

    await useUrlsStore.getState().previousPage();

    const state = useUrlsStore.getState();

    expect(state.page).toBe(1);
    expect(state.currentCursor).toBe("");
    expect(state.urls).toEqual(page1.urls);
    expect(state.hasMore).toBe(true);
    expect(state.nextCursor).toBe(page1.urls[page1.urls.length - 1].expired_at);
    expect(state.cursorHistory.length).toEqual(0);
  });
});
