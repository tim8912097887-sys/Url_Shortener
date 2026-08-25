// Best-effort, non-verifying decode of a JWT's payload, used only to hydrate
// a display-friendly "user" object (e.g. email/username) on the client.
// The server remains the source of truth for authorization; nothing here is

import type { TokenResponse } from "../services/types";

// trusted for access-control decisions.
export function decodeAccessToken(token: string | null) {
  if (!token || typeof token !== "string") return null;

  const segments = token.split(".");
  if (segments.length < 2) return null;

  try {
    const base64 = segments[1].replace(/-/g, "+").replace(/_/g, "/");
    const padded = base64.padEnd(
      base64.length + ((4 - (base64.length % 4)) % 4),
      "=",
    );
    const payload = JSON.parse(atob(padded));

    return {
      id: payload.sub ?? payload.userId ?? payload.user_id ?? null,
      email: payload.email ?? null,
      username: payload.username ?? null,
      tokenVersion: payload.tokenVersion ?? payload.token_version ?? null,
      exp: payload.exp ?? null,
    };
  } catch {
    return null;
  }
}

export function extractAccessToken(tokenResponse: TokenResponse) {
  return tokenResponse.data.accessToken;
}
