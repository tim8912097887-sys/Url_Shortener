import { useState } from "react";
import ShortUrlResult from "../components/url/ShortUrlResult";
import UrlCard from "../components/url/UrlCard";

const HomePage = () => {
  const [shortUrl, setShortUrl] = useState("");

  return (
    <div className="flex min-h-[calc(100vh-5rem)] items-center justify-center px-4 py-12">
      <div className="w-full max-w-xl flex flex-col items-center">
        <UrlCard setShortUrl={setShortUrl} />
        <ShortUrlResult shortUrl={shortUrl} />
      </div>
    </div>
  );
};

export default HomePage;
