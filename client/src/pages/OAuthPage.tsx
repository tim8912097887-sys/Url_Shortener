// React / JavaScript Example
import { useEffect } from "react";
import { useNavigate } from "react-router";
import { useAuthStore } from "../store/useAuthStore";

export function OAuthPage() {
  const navigate = useNavigate();
  const setAccessToken = useAuthStore((state) => state.setAccessToken);

  useEffect(() => {
    // Extract access_token from URL hash fragment (#access_token=xyz)
    const hashParams = new URLSearchParams(window.location.hash.substring(1));
    const accessToken = hashParams.get("access_token");

    if (accessToken) {
      setAccessToken(accessToken);
      // Clear hash from URL for security
      window.history.replaceState(null, "", window.location.pathname);

      // Redirect user to Home page
      navigate("/");
    } else {
      navigate("/login?error=token_missing");
    }
  }, [navigate, setAccessToken]);

  return <div>Logging you in...</div>;
}
