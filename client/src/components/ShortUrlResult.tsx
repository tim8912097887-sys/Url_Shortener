import { useState } from "react";

interface ShortUrlResultProps {
  shortUrl: string;
}

const ShortUrlResult = ({ shortUrl }: ShortUrlResultProps) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(shortUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000); // Reset after 2 seconds
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
  };

  return (
    <div className="bg-white border border-slate-100 shadow-xl shadow-slate-200/50 rounded-2xl p-6 w-full mt-6 animate-fade-in">
      <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
        Your Shortened Link
      </p>

      <div className="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center justify-between bg-slate-50 border border-slate-200 p-3 rounded-xl">
        <a
          href={shortUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="text-emerald-600 font-medium hover:underline truncate px-2 text-sm sm:text-base"
        >
          {shortUrl}
        </a>

        <button
          onClick={handleCopy}
          className={`px-5 py-2.5 rounded-lg text-sm font-semibold transition-all cursor-pointer whitespace-nowrap ${
            copied
              ? "bg-slate-900 text-white shadow-none"
              : "bg-white text-slate-700 border border-slate-200 shadow-sm hover:bg-slate-100 active:scale-[0.98]"
          }`}
        >
          {copied ? "Copied!" : "Copy Link"}
        </button>
      </div>
    </div>
  );
};

export default ShortUrlResult;
