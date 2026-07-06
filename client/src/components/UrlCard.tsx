import { useState } from "react";
import { fetchWrapper } from "../api/fetch-wrapper";
import { shortenUrl } from "../api/url";
import type {
  SuccessResponse,
  ShortenData,
  ErrorResponse,
} from "../api/response";
import { API_BASE_URL } from "../api/constants";

type UrlCardProps = {
  setShortUrl: (shortUrl: string) => void;
};

const UrlCard = ({ setShortUrl }: UrlCardProps) => {
  const [longUrl, setLongUrl] = useState("");
  const [error, setError] = useState("");

  async function handleSubmit(e: React.SubmitEvent<HTMLFormElement>) {
    e.preventDefault();

    const response = await fetchWrapper<
      SuccessResponse<ShortenData>,
      ErrorResponse
    >(shortenUrl(longUrl));
    if (response.ok) {
      setShortUrl(`${API_BASE_URL}/urls/${response.data?.data.shortUrl}`);
    } else {
      setError(response.error?.error.message);
    }
  }
  return (
    <form
      onSubmit={handleSubmit}
      className="bg-white border border-slate-100 shadow-xl shadow-slate-200/50 rounded-2xl p-8 md:p-10 w-full transition-all"
    >
      <h2 className="font-bold text-2xl text-slate-900 tracking-tight mb-2">
        Shorten your long URL
      </h2>
      <p className="text-sm text-slate-500 mb-6">
        Paste your link below to create a clean, shareable short link instantly.
      </p>

      <div className="space-y-4">
        <div className="relative space-y-2">
          <input
            value={longUrl}
            onChange={(e) => setLongUrl(e.target.value)}
            className="w-full px-4 py-3.5 bg-slate-50 border border-slate-200 text-slate-900 text-sm rounded-xl block placeholder-slate-400 focus:outline-none focus:border-emerald-500 focus:ring-4 focus:ring-emerald-500/10 transition-all"
            placeholder="https://example.com/your-long-painful-url-here"
            type="url"
          />
          {error && <p className="text-xs text-rose-500">{error}</p>}
        </div>

        <button
          type="submit"
          className="bg-emerald-600 hover:bg-emerald-700 text-white font-semibold py-3.5 px-6 rounded-xl w-full shadow-md shadow-emerald-600/10 hover:shadow-lg hover:shadow-emerald-600/20 active:scale-[0.99] transition-all cursor-pointer text-center"
        >
          Shorten URL
        </button>
      </div>
    </form>
  );
};

export default UrlCard;
