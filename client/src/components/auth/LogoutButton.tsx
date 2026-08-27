import { useState } from "react";
import { useNavigate } from "react-router";
import Button from "../ui/button/Button";
import { useAuthStore } from "../../store/useAuthStore";
import type { BUTTON_SIZES, BUTTON_VARIANTS } from "../ui/button/types";

type LogoutButtonProps = {
  allDevices?: boolean;
  variant?: keyof typeof BUTTON_VARIANTS;
  size?: keyof typeof BUTTON_SIZES;
  className?: string;
  children?: React.ReactNode;
};

// <LogoutButton /> logs out the current session/device.
// <LogoutButton allDevices /> calls logout-all instead (bumps the user's
// token version server-side, invalidating every outstanding refresh token).
export default function LogoutButton({
  allDevices = false,
  variant = "secondary",
  size = "sm",
  className = "",
  children,
}: LogoutButtonProps) {
  const navigate = useNavigate();
  const logout = useAuthStore((s) => s.logout);
  const logoutAll = useAuthStore((s) => s.logoutAll);
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const handleClick = async () => {
    setIsLoggingOut(true);
    try {
      await (allDevices ? logoutAll() : logout());
    } finally {
      setIsLoggingOut(false);
      navigate("/login", { replace: true });
    }
  };

  return (
    <Button
      type="button"
      variant={variant}
      size={size}
      isLoading={isLoggingOut}
      onClick={handleClick}
      className={className}
    >
      {children ?? (allDevices ? "Log out of all devices" : "Log out")}
    </Button>
  );
}
