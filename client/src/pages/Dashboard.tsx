import { useEffect, useState } from "react";
import { urlService } from "../api/services/url.service";
import UrlCard from "../components/url/UrlCard";
import type { ApiError } from "../api/errors/api-error";

type Url = {
  long_url: string;
  short_url: string;
  expired_at: string;
  clicks: number;
};

const Dashboard = () => {
  const [urls, setUrls] = useState<Url[] | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUrls = async () => {
      try {
        setError(null);
        const data = await urlService.getUrlsForUser();
        await new Promise((resolve) => setTimeout(resolve, 3000));
        setUrls(data.data.urls);
      } catch (error) {
        const axiosError = error as ApiError;
        setError(
          axiosError.message ?? "Failed to load your URLs. Please try again.",
        );
      } finally {
        setIsLoading(false);
      }
    };
    fetchUrls();
  }, []);

  return (
    <main className="min-h-[calc(100vh-5rem)] bg-slate-50">
      <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold tracking-tight text-slate-900 sm:text-3xl">
            Your URLs
          </h1>

          <p className="mt-2 text-sm text-slate-500 sm:text-base">
            Manage and track the URLs you have shortened.
          </p>
        </div>

        {/* Loading */}
        {isLoading && (
          <div className="space-y-4" aria-label="Loading URLs">
            {Array.from({ length: 5 }).map((_, index) => (
              <UrlCardSkeleton key={index} />
            ))}
          </div>
        )}

        {/* Error */}
        {!isLoading && error && (
          <div
            role="alert"
            className="rounded-xl border border-red-200 bg-red-50 px-4 py-4 text-sm text-red-700"
          >
            <p className="font-medium">Unable to load URLs</p>
            <p className="mt-1 text-red-600">{error}</p>
          </div>
        )}

        {/* Empty */}
        {!isLoading && !error && (!urls || urls.length === 0) && (
          <div className="flex min-h-80 flex-col items-center justify-center rounded-2xl border border-dashed border-slate-300 bg-white px-6 text-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-slate-100">
              <svg
                className="h-6 w-6 text-slate-400"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.8"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M13.5 6.5l-1.4-1.4a2 2 0 00-2.8 0l-5.8 5.8a2 2 0 000 2.8l1.4 1.4a2 2 0 002.8 0l1.8-1.8m1-1l1.8-1.8a2 2 0 012.8 0l1.4 1.4a2 2 0 010 2.8l-5.8 5.8a2 2 0 01-2.8 0l-1.4-1.4"
                />
              </svg>
            </div>

            <h2 className="mt-4 text-base font-semibold text-slate-900">
              No URLs yet
            </h2>

            <p className="mt-1 max-w-sm text-sm text-slate-500">
              Create your first shortened URL and it will appear here.
            </p>
          </div>
        )}

        {/* URL list */}
        {!isLoading && !error && urls && urls.length > 0 && (
          <div className="space-y-4">
            {urls.map((url) => (
              <UrlCard key={url.short_url} {...url} />
            ))}
          </div>
        )}
      </div>
    </main>
  );
};

const UrlCardSkeleton = () => {
  return (
    <div className="animate-pulse rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
        <div className="min-w-0 flex-1 space-y-3">
          <div className="h-5 w-48 rounded bg-slate-200" />
          <div className="h-4 w-full max-w-md rounded bg-slate-200" />
        </div>

        <div className="space-y-2 sm:w-40">
          <div className="ml-auto h-4 w-24 rounded bg-slate-200" />
          <div className="ml-auto h-4 w-32 rounded bg-slate-200" />
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
