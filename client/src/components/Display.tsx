import { useState } from "react";
import ShortUrlResult from "./ShortUrlResult";
import UrlCard from "./UrlCard";

const Display = () => {
  const [shortUrl, setShortUrl] = useState("");

  return (
    <div className="w-full max-w-xl">
      <UrlCard setShortUrl={setShortUrl} />
      <ShortUrlResult shortUrl={shortUrl} />
    </div>
  );
};

export default Display;
