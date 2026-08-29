import { useState } from "react";
import { env } from "../../config/env";

type UrlCardProps = {
  clicks: number;
  expired_at: string;
  short_url: string;
  long_url: string;
};

const UrlCard = ({ clicks, expired_at, short_url, long_url }: UrlCardProps) => {
  const [copied, setCopied] = useState(false);

  const formattedExpiry = new Date(expired_at).toLocaleString();

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(short_url);

      setCopied(true);

      window.setTimeout(() => {
        setCopied(false);
      }, 2000);
    } catch {
      // Clipboard failures should not break the card.
    }
  };

  const shortUrl = `${env.apiBaseUrl}/urls/${short_url}`;

  return (
    <article className="group rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition hover:border-slate-300 hover:shadow-md">
      <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
        {/* URL information */}
        <div className="min-w-0 flex-1">
          <div className="flex min-w-0 items-center gap-2">
            <a
              href={shortUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="min-w-0 truncate text-base font-semibold text-blue-600 transition hover:text-blue-700 hover:underline sm:text-lg"
              title={shortUrl}
            >
              {shortUrl}
            </a>

            <button
              type="button"
              onClick={handleCopy}
              aria-label="Copy short URL"
              className="shrink-0 rounded-md p-1.5 text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1"
            >
              {copied ? (
                <svg
                  className="h-4 w-4 text-emerald-600"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              ) : (
                <svg
                  className="h-4 w-4"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.8"
                >
                  <rect width="9" height="9" x="9" y="9" rx="1" />
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M5 15H4a1 1 0 01-1-1V4a1 1 0 011-1h10a1 1 0 011 1v1"
                  />
                </svg>
              )}
            </button>
          </div>

          {copied && (
            <p className="mt-1 text-xs font-medium text-emerald-600">
              Copied to clipboard
            </p>
          )}

          <p
            className="mt-2 max-w-2xl truncate text-sm text-slate-500"
            title={long_url}
          >
            {long_url}
          </p>
        </div>

        {/* Metadata */}
        <div className="flex shrink-0 flex-wrap items-center gap-x-6 gap-y-3 border-t border-slate-100 pt-4 lg:border-l lg:border-t-0 lg:pl-6 lg:pt-0">
          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-slate-400">
              Clicks
            </p>

            <p className="mt-1 text-sm font-semibold text-slate-800">
              {clicks.toLocaleString()}
            </p>
          </div>

          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-slate-400">
              Expires
            </p>

            <p
              className="mt-1 text-sm font-medium text-slate-600"
              title={formattedExpiry}
            >
              {formattedExpiry}
            </p>
          </div>
        </div>
      </div>
    </article>
  );
};

export default UrlCard;
