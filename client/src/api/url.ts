import { API_BASE_URL } from "./constants";

export function shortenUrl(url: string) {
  return async () => {
    const response = await fetch(`${API_BASE_URL}/urls`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url }),
    });
    return response;
  };
}
